package api

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *Application) Routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/service/users/health", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/service/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPost, "/service/users/authentication", app.createAuthenticationTokenHandler)
	router.HandlerFunc(http.MethodGet, "/service/users/me", app.requireAuthenticated(app.getUserHandler))
	router.HandlerFunc(http.MethodPatch, "/service/users/:id", app.requireAuthenticated(app.patchUserHandler))

	router.Handler(http.MethodGet, "/service/users/debug/vars", expvar.Handler())

	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
