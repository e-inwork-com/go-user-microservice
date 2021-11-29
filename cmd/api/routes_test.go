package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"user.services.e-inwork.com/internal/data"
	"user.services.e-inwork.com/internal/jsonlog"

	"github.com/cockroachdb/cockroach-go/v2/testserver"
	"github.com/stretchr/testify/assert"
)

func TestRoutes(t *testing.T) {
	// Server Setup
	tsDB, err := testserver.NewTestServer()
	assert.Nil(t, err)
	urlDB := tsDB.PGURL()

	var cfg config
	cfg.db.dsn = urlDB.String()
	cfg.auth.secret = "secret"
	cfg.db.maxOpenConn = 25
	cfg.db.maxIdleConn = 25
	cfg.db.maxIdleTime = "15m"
	cfg.limiter.enabled = true
	cfg.limiter.rps = 2
	cfg.limiter.burst = 4

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cfg)
	assert.Nil(t, err)
	defer db.Close()

	_, err = db.Exec("" +
		"CREATE TABLE IF NOT EXISTS users (" +
		"id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid()," +
		"created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()," +
		"name text NOT NULL," +
		"email text UNIQUE NOT NULL," +
		"password_hash bytea NOT NULL," +
		"activated bool NOT NULL DEFAULT false," +
		"version integer NOT NULL DEFAULT 1);")
	assert.Nil(t, err)

	app := &application{
		config: cfg,
		logger: logger,
		models: data.InitModels(db),
	}

	ts := httptest.NewTLSServer(app.routes())
	defer ts.Close()

	// Register
	user := `{"name": "Test", "email": "test@example.com", "password": "pa55word"}`
	res, err := ts.Client().Post(ts.URL+"/api/users", "application/json", bytes.NewReader([]byte(user)))
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusAccepted)

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)

	var userResult map[string]data.User
	err = json.Unmarshal(body, &userResult)
	assert.Nil(t, err)
	assert.Equal(t, userResult["user"].Email, "test@example.com")

	// User Token Authentication Sign In
	user = `{"email": "test@example.com", "password": "pa55word"}`
	res, err = ts.Client().Post(ts.URL+"/api/authentication", "application/json", bytes.NewReader([]byte(user)))
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusCreated)

	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.Nil(t, err)

	type authType struct{
		Token string `json:"token"`
	}
	var authResult authType
	err = json.Unmarshal(body, &authResult)
	assert.Nil(t, err)
	assert.NotNil(t, authResult.Token)
}