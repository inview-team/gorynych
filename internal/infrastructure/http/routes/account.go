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

func AddAccount(s *service.AccountService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorMessage := "Error creating account"
		ctx := r.Context()

		cAccount := new(controllers.Account)
		if err := json.NewDecoder(r.Body).Decode(&cAccount); err != nil {
			log.Errorf("failed to decode payload")
			http.Error(w, errorMessage, http.StatusBadRequest)
			return
		}

		id, err := s.AddAccount(ctx, entity.Provider(cAccount.Provider), cAccount.KeyID, cAccount.Secret)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&views.ID{ID: id})
	})
}

func makeAccountRoutes(r *mux.Router, app *application.Application) {
	path := "/accounts"
	serviceRouter := r.PathPrefix(path).Subrouter()
	serviceRouter.Handle("", AddAccount(app.AccountService)).Methods("POST")
}
