package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/e-inwork-com/go-user-service/internal/data"
	"github.com/e-inwork-com/go-user-service/internal/data/mocks"
	"github.com/e-inwork-com/go-user-service/internal/jsonlog"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func testApplication(t *testing.T) *Application {

	var cfg Config
	cfg.Auth.Secret = "secret"

	return &Application{
		Config: cfg,
		Logger: jsonlog.New(os.Stdout, jsonlog.LevelInfo),
		Models: data.Models{
			Users: &mocks.UserModel{},
		},
	}

}

type httpTestServer struct {
	*httptest.Server
}

func testServer(t *testing.T, h http.Handler) *httpTestServer {
	ts := httptest.NewTLSServer(h)

	return &httpTestServer{ts}
}

func (ts *httpTestServer) request(t *testing.T, method string, urlPath string, contentType string, authToken string, body io.Reader) (int, http.Header, string) {
	rq, _ := http.NewRequest(method, ts.URL+urlPath, body)

	if contentType != "" {
		rq.Header.Add("Content-Type", contentType)
	}

	if authToken != "" {
		rq.Header.Set("Authorization", fmt.Sprintf("Bearer %v", authToken))
	}

	rs, err := ts.Client().Do(rq)
	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()
	bd, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	bytes.TrimSpace(bd)

	return rs.StatusCode, rs.Header, string(bd)
}

func (app *Application) testCreateToken(t *testing.T, id uuid.UUID) string {
	// Set Signing Key from the Config Environment
	signingKey := []byte(app.Config.Auth.Secret)

	// Set an expired time for a week
	expirationTime := time.Now().Add((24 * 7) * time.Hour)

	// Set the ID of the user in the Claim token
	claims := &Claims{
		ID: id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Create a signed token
	signed := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := signed.SignedString(signingKey)
	if err != nil {
		t.Fatal(err)
	}

	return token
}

func (app *Application) testFirstToken(t *testing.T) string {
	// Create UUID
	id := mocks.MockFirstUUID()

	return app.testCreateToken(t, id)
}

func (app *Application) testSecondToken(t *testing.T) string {
	// Create UUID
	id := mocks.MockSecondUUID()

	return app.testCreateToken(t, id)
}

func (app *Application) testBodyCreateUser(t *testing.T) io.Reader {
	user := `{"email_t": "jon@doe.com", "password": "pa55word", "first_name_t": "Jon", "last_name_t": "Doe"}`
	return bytes.NewReader([]byte(user))
}

func (app *Application) testBodyLoginUser(t *testing.T) io.Reader {
	user := `{"email_t": "jon@doe.com", "password": "pa55word"}`
	return bytes.NewReader([]byte(user))
}

func (app *Application) testBodyUpdateUser(t *testing.T) io.Reader {
	user := `{"password": "pa00word", "first_name_t": "Nina"}`
	return bytes.NewReader([]byte(user))
}

func (app *Application) testBodyUpdateUserFobidden(t *testing.T) io.Reader {
	user := `{"password": "pa11w0rd", "email_t": "lee@john.com"}`
	return bytes.NewReader([]byte(user))
}
