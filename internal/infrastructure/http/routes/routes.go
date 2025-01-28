package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/inview-team/gorynych/internal/application"
	"github.com/inview-team/gorynych/internal/infrastructure/http/handlers"
	httpSwagger "github.com/swaggo/http-swagger"
)

func Make(app *application.Application) http.Handler {
	r := mux.NewRouter()
	r.PathPrefix("/docs/").Handler(httpSwagger.WrapHandler)

	r.MethodNotAllowedHandler = handlers.NotAllowedHandler()
	r.NotFoundHandler = handlers.NotFoundHandler()
	makeFileRoutes(r, app)
	return r
}
