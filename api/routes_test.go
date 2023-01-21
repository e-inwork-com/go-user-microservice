package api

import (
	"io"
	"net/http"
	"testing"

	"github.com/e-inwork-com/go-user-service/internal/data/mocks"
	"github.com/stretchr/testify/assert"
)

func TestRoutes(t *testing.T) {
	app := testApplication(t)

	ts := testServer(t, app.Routes())
	defer ts.Close()

	tBodyCreateUser := app.testBodyCreateUser(t)
	tBodyLoginUser := app.testBodyLoginUser(t)
	tBodyUpdateUser := app.testBodyUpdateUser(t)
	firstToken := app.testFirstToken(t)
	tBodyUpdateUserForbidden := app.testBodyUpdateUserFobidden(t)
	secondToken := app.testSecondToken(t)

	tests := []struct {
		name         string
		method       string
		urlPath      string
		contentType  string
		token        string
		body         io.Reader
		expectedCode int
	}{
		{
			name:         "Register User",
			method:       "POST",
			urlPath:      "/service/users",
			contentType:  "application/json",
			token:        "",
			body:         tBodyCreateUser,
			expectedCode: http.StatusCreated,
		},
		{
			name:         "Login User",
			method:       "POST",
			urlPath:      "/service/users/authentication",
			contentType:  "application/json",
			token:        "",
			body:         tBodyLoginUser,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Update User",
			method:       "PATCH",
			urlPath:      "/service/users/" + mocks.MockFirstUUID().String(),
			contentType:  "application/json",
			token:        firstToken,
			body:         tBodyUpdateUser,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Update User Forbidden",
			method:       "PATCH",
			urlPath:      "/service/users/" + mocks.MockFirstUUID().String(),
			contentType:  "application/json",
			token:        secondToken,
			body:         tBodyUpdateUserForbidden,
			expectedCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualCode, _, _ := ts.request(t, tt.method, tt.urlPath, tt.contentType, tt.token, tt.body)
			assert.Equal(t, tt.expectedCode, actualCode)
		})
	}
}
