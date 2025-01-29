package routes

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/inview-team/gorynych/internal/application"
	"github.com/inview-team/gorynych/internal/domain/entity"
	"github.com/inview-team/gorynych/internal/domain/service"
	"github.com/inview-team/gorynych/internal/infrastructure/http/controllers"
	"github.com/inview-team/gorynych/internal/infrastructure/http/views"
)

func CreateUpload(s *service.UploadService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		config := s.GetServiceConfiguration()

		size, err := strconv.ParseInt(r.Header.Get("Upload-Length"), 10, 64)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
		}

		if config.MaxSize > 0 && size > config.MaxSize {
			http.Error(w, "", http.StatusRequestEntityTooLarge)
			return
		}

		meta := controllers.NewMetadata(r.Header.Get("Upload-Metadata"))

		id, err := s.CreateUpload(ctx, size, meta)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
		}

		w.Header().Set("Location", fmt.Sprintf("%s/%s", r.URL.String(), string(id)))
		w.WriteHeader(http.StatusCreated)
	})
}

func WriteChunk(s *service.UploadService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		uploadID := mux.Vars(r)["upload_id"]

		if r.Header.Get("Content-Type") != "application/offset+octet-stream" {
			http.Error(w, "wrong content type", http.StatusBadRequest)
		}

		// Check for presence of a valid Upload-Offset Header
		offset, err := strconv.ParseInt(r.Header.Get("Upload-Offset"), 10, 64)
		if err != nil || offset < 0 {
			http.Error(w, "wrong offset", http.StatusBadRequest)
			return
		}

		var bodyBuffer []byte
		bodyBuffer, err = io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
		}

		newOffset, err := s.WritePart(ctx, entity.UploadID(uploadID), offset, bodyBuffer)
		if err != nil {
			if errors.Is(err, service.ErrUploadNotFound) {
				http.Error(w, "", http.StatusNotFound)
				return
			}

			if errors.Is(err, service.ErrWrongOffset) {
				http.Error(w, "", http.StatusConflict)
				return
			}

			if errors.Is(err, service.ErrUploadBig) {
				http.Error(w, "", http.StatusRequestEntityTooLarge)
				return
			}
			http.Error(w, "", http.StatusInternalServerError)
		}

		w.Header().Add("Upload-Offset", strconv.Itoa(int(newOffset)))
		w.Header().Add("Tus-Resumable", "1.0.0")
		w.WriteHeader(http.StatusNoContent)
	})
}

func GetUploadInformation(s *service.UploadService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		uploadID := mux.Vars(r)["upload_id"]

		uploadInfo, err := s.GetUpload(ctx, entity.UploadID(uploadID))
		if err != nil {
			if errors.Is(err, service.ErrUploadNotFound) {
				http.Error(w, "", http.StatusNotFound)
				return
			}
		}

		metaHeader := views.NewResponseMetadata(uploadInfo.Metadata)
		if metaHeader != "" {
			w.Header().Add("Upload-Metadata", metaHeader)
		}
		w.Header().Add("Upload-Offset", strconv.Itoa(int(uploadInfo.Offset)))
		w.Header().Add("Upload-Length", strconv.Itoa(int(uploadInfo.Size)))
		w.Header().Add("Cache-Control", "no-store")
		w.Header().Add("Tus-Resumable", "1.0.0")
		w.WriteHeader(http.StatusNoContent)
	})
}

func GetServerInformation(service *service.UploadService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Tus-Version", "1.0.0")
		w.Header().Add("Tus-Resumable", "1.0.0")
		w.Header().Add("Tus-Extension", "creation")
		config := service.GetServiceConfiguration()
		if config.MaxSize > 0 {
			w.Header().Set("Tus-Max-Size", strconv.Itoa(int(config.MaxSize)))
		}
		w.WriteHeader(http.StatusNoContent)
		return
	})
}

func makeFileRoutes(r *mux.Router, app *application.Application) {
	path := "/files"
	serviceRouter := r.PathPrefix(path).Subrouter()
	serviceRouter.Handle("", CreateUpload(app.UploadService)).Methods("POST")
	serviceRouter.Handle(fmt.Sprintf("/%s", patternUploadID), GetUploadInformation(app.UploadService)).Methods("HEAD")
	serviceRouter.Handle("", GetServerInformation(app.UploadService)).Methods("OPTIONS")
	serviceRouter.Handle(fmt.Sprintf("/%s", patternUploadID), WriteChunk(app.UploadService)).Methods("PATCH")
}
