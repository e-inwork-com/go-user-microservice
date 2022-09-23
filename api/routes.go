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

	router.HandlerFunc(http.MethodGet, "/api/health", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/api/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPost, "/api/authentication", app.createAuthenticationTokenHandler)
	router.HandlerFunc(http.MethodGet, "/api/users/me", app.requireAuthenticated(app.getUserHandler))
	router.HandlerFunc(http.MethodPatch, "/api/users/:id", app.requireAuthenticated(app.patchUserHandler))

	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
