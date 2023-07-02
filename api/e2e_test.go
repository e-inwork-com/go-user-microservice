package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/e-inwork-com/go-user-service/internal/data"
	"github.com/e-inwork-com/go-user-service/internal/jsonlog"
	"github.com/stretchr/testify/assert"
)

func TestE2E(t *testing.T) {
	// Team Service
	var cfg Config
	cfg.Db.Dsn = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	cfg.Auth.Secret = "secret"
	cfg.Db.MaxOpenConn = 25
	cfg.Db.MaxIdleConn = 25
	cfg.Db.MaxIdleTime = "15m"
	cfg.Limiter.Enabled = true
	cfg.Limiter.Rps = 2
	cfg.Limiter.Burst = 6

	// Set logger
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// Set Database
	db, err := OpenDB(cfg)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer db.Close()

	// Set Applcation
	app := Application{
		Config: cfg,
		Logger: logger,
		Models: data.InitModels(db),
	}

	// Server Routes API
	ts := httptest.NewTLSServer(app.Routes())
	defer ts.Close()

	// Read a SQL file for the deleting all records
	script, err := os.ReadFile("./test/sql/delete_all.sql")
	if err != nil {
		t.Fatal(err)
	}

	// Delete all records
	_, err = db.Exec(string(script))
	if err != nil {
		t.Fatal(err)
	}

	// Initial email & password
	email := "jon@doe.com"
	password := "pa55word"

	// Initial user resposnse
	var userResponse map[string]data.User

	t.Run("Register User", func(t *testing.T) {
		data := fmt.Sprintf(
			`{"email_t": "%v", "password": "%v", "first_name_t": "Jon", "last_name_t": "Doe"}`,
			email,
			password)
		req, _ := http.NewRequest(
			"POST",
			ts.URL+"/service/users",
			bytes.NewReader([]byte(data)))
		req.Header.Add("Content-Type", "application/json")

		res, err := ts.Client().Do(req)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, res.StatusCode)

		body, err := ioutil.ReadAll(res.Body)
		defer res.Body.Close()
		assert.Nil(t, err)

		err = json.Unmarshal(body, &userResponse)
		assert.Nil(t, err)
		assert.Equal(t, email, userResponse["user"].Email)
	})

	// Initial Authentication
	type authType struct {
		Token string `json:"token"`
	}
	var authentication authType

	t.Run("Login User", func(t *testing.T) {
		data := fmt.Sprintf(
			`{"email_t": "%v", "password": "%v"}`,
			email,
			password)
		req, _ := http.NewRequest(
			"POST",
			ts.URL+"/service/users/authentication",
			bytes.NewReader([]byte(data)))
		req.Header.Add("Content-Type", "application/json")

		res, err := ts.Client().Do(req)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, err := ioutil.ReadAll(res.Body)
		defer res.Body.Close()
		assert.Nil(t, err)

		err = json.Unmarshal(body, &authentication)
		assert.Nil(t, err)
		assert.NotNil(t, authentication.Token)
	})

	t.Run("Get User with Authication Token ", func(t *testing.T) {
		req, _ := http.NewRequest(
			"GET",
			ts.URL+"/service/users/me",
			nil)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Set(
			"Authorization",
			fmt.Sprintf("Bearer %v", authentication.Token))

		res, err := ts.Client().Do(req)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, err := ioutil.ReadAll(res.Body)
		defer res.Body.Close()
		assert.Nil(t, err)

		err = json.Unmarshal(body, &userResponse)
		assert.Nil(t, err)
		assert.Equal(t, email, userResponse["user"].Email)
	})

	// Initail new email & new passs
	newEmail := "test@email.com"
	newPassword := "password"

	t.Run("Patch User with New Email & Password", func(t *testing.T) {
		data := fmt.Sprintf(
			`{"email_t": "%v", "password": "%v"}`,
			newEmail,
			newPassword)
		req, _ := http.NewRequest(
			"PATCH",
			ts.URL+"/service/users/"+userResponse["user"].ID.String(),
			bytes.NewReader([]byte(data)))
		req.Header.Add("Content-Type", "application/json")
		req.Header.Set(
			"Authorization",
			fmt.Sprintf("Bearer %v", authentication.Token))

		res, err := ts.Client().Do(req)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, err := ioutil.ReadAll(res.Body)
		defer res.Body.Close()
		assert.Nil(t, err)

		err = json.Unmarshal(body, &userResponse)
		assert.Nil(t, err)
		assert.Equal(t, newEmail, userResponse["user"].Email)
	})

	t.Run("Login User with New Email & Password", func(t *testing.T) {
		data := fmt.Sprintf(
			`{"email_t": "%v", "password": "%v"}`,
			newEmail,
			newPassword)
		req, _ := http.NewRequest(
			"POST",
			ts.URL+"/service/users/authentication",
			bytes.NewReader([]byte(data)))
		req.Header.Add("Content-Type", "application/json")

		res, err := ts.Client().Do(req)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body, err := ioutil.ReadAll(res.Body)
		defer res.Body.Close()
		assert.Nil(t, err)

		err = json.Unmarshal(body, &authentication)
		assert.Nil(t, err)
		assert.NotNil(t, authentication.Token)
	})
}
