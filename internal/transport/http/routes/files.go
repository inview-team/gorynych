package routes

import (
	"encoding/json"
	"net/http"

	"github.com/D3vR4pt0rs/logger"
	"github.com/gorilla/mux"
	"github.com/inview-team/gorynych/internal/application"
	"github.com/inview-team/gorynych/internal/service/object"
)

// uploadFile godoc
//
//	@Summary		Upload file
//	@Description	upload file
//	@Tags			Files
//	@Accept			mpfd
//	@Produce		json
//	@Param			file	formData	file	true	"Body with file"
//	@Success		200
//	@Router			/files [post]
func uploadCompetitionResult(service object.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorMessage := "failed to upload result"

		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			logger.Error.Printf("failed to upload file: %v", err.Error())
			http.Error(w, errorMessage, http.StatusBadRequest)
			return
		}

		f, _, err := r.FormFile("file")
		if err != nil {
			logger.Error.Println("failed to get file")
			http.Error(w, errorMessage, http.StatusBadRequest)
			return
		}
		defer f.Close()

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode("File uploaded successfully")
	})
}

func makeObjectRoutes(r *mux.Router, app *application.Application) {
	path := "/files"
	serviceRouter := r.PathPrefix(path).Subrouter()

}
