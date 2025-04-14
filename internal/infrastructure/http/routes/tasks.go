package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/inview-team/gorynych/internal/application"
	"github.com/inview-team/gorynych/internal/domain/entity"
	"github.com/inview-team/gorynych/internal/domain/service"
	"github.com/inview-team/gorynych/internal/infrastructure/http/controllers"
	log "github.com/sirupsen/logrus"
)

func ReplicateFile(s *service.ReplicationService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorMessage := "Error replicate object"
		ctx := r.Context()
		objectID := mux.Vars(r)["object_id"]

		cTask := new(controllers.ReplicateInput)
		if err := json.NewDecoder(r.Body).Decode(&cTask); err != nil {
			log.Errorf("failed to decode payload")
			http.Error(w, errorMessage, http.StatusBadRequest)
			return
		}

		err := s.Replicate(ctx, objectID, 0, entity.Storage(cTask.SourceStorage), entity.Storage(cTask.TargetStorage))
		if err != nil {
			switch err {
			case service.ErrObjectNotFound:
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			case service.ErrBucketNotFound:
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			default:
				log.Errorf(err.Error())
				http.Error(w, errorMessage, http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusAccepted)
	})
}

func makeTaskRoutes(r *mux.Router, app *application.Application) {
	path := "/replicate"
	serviceRouter := r.PathPrefix(path).Subrouter()
	serviceRouter.Handle("/{object_id}", ReplicateFile(app.ReplicationService)).Methods("POST")
}
