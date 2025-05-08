package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/inview-team/gorynych/internal/application"
	"github.com/inview-team/gorynych/internal/domain/entity"
	"github.com/inview-team/gorynych/internal/domain/service"
	"github.com/inview-team/gorynych/internal/infrastructure/http/controllers"
	"github.com/inview-team/gorynych/internal/infrastructure/http/views"
	log "github.com/sirupsen/logrus"
)

func ReplicateFile(s *service.TaskService) http.Handler {
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

		taskID, err := s.Replication(ctx, objectID, entity.Storage(cTask.SourceStorage), entity.Storage(cTask.TargetStorage))
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusAccepted)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&views.ID{ID: taskID})
	})
}

func makeTaskRoutes(r *mux.Router, app *application.Application) {
	path := "/tasks"
	serviceRouter := r.PathPrefix(path).Subrouter()
	serviceRouter.Handle("/replicate/{object_id}", ReplicateFile(app.TaskService)).Methods("POST")
}
