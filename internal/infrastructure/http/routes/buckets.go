package routes

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/inview-team/gorynych/internal/application"
)

func getBuckets() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorMessage := "failed to get buckets"
		ctx := r.Context()

	})
}

func makeBucketRoutes(r *mux.Router, app *application.Application) {
	path := "/buckets"
	serviceRouter := r.PathPrefix(path).Subrouter()
	serviceRouter.Handle("").Methods("GET")
}
