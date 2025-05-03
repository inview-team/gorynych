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

func ReplicateFile(s *service.WorkerService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorMessage := "Error replicate object"

		objectID := mux.Vars(r)["object_id"]

		cTask := new(controllers.ReplicateInput)
		if err := json.NewDecoder(r.Body).Decode(&cTask); err != nil {
			log.Errorf("failed to decode payload")
			http.Error(w, errorMessage, http.StatusBadRequest)
			return
		}

		task := entity.NewReplicationTask(objectID, entity.Storage(cTask.SourceStorage), entity.Storage(cTask.TargetStorage))
		s.Submit(*task)
		w.WriteHeader(http.StatusAccepted)
	})
}

func makeTaskRoutes(r *mux.Router, app *application.Application) {
	path := "/tasks"
	serviceRouter := r.PathPrefix(path).Subrouter()
	serviceRouter.Handle("/replicate/{object_id}", ReplicateFile(app.WorkerService)).Methods("POST")
}
