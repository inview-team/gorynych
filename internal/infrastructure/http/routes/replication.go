package routes

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/inview-team/gorynych/internal/application"
	"github.com/inview-team/gorynych/internal/domain/service"
	"github.com/inview-team/gorynych/internal/infrastructure/http/controllers"
	"github.com/inview-team/gorynych/internal/infrastructure/http/views"
)

func ReplicateObject(s *service.ReplicationService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorMessage := "Error replicate file"
		ctx := r.Context()

		mFile := new(controllers.File)
		if err := json.NewDecoder(r.Body).Decode(&mFile); err != nil {
			http.Error(w, errorMessage, http.StatusBadRequest)
			return
		}

		objectID, bucket, err := s.CreateReplication(ctx, mFile.ToEntity())
		if err != nil {
			if errors.Is(err, service.ErrObjectNotFound) {
				http.Error(w, errorMessage, http.StatusNotFound)
				return
			}
			http.Error(w, errorMessage, http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(views.NewReplicatedFile(objectID, bucket))
	})
}

func makeReplicationRoutes(r *mux.Router, app *application.Application) {
	path := "/replication"
	serviceRouter := r.PathPrefix(path).Subrouter()
	serviceRouter.Handle("", ReplicateObject(app.ReplicationService)).Methods("POST")
}
