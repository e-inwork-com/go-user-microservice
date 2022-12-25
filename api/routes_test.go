package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/e-inwork-com/go-user-service/pkg/data"
	"github.com/e-inwork-com/go-user-service/pkg/jsonlog"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	log "github.com/sirupsen/logrus"
)

var db *sql.DB

func TestRoutes(t *testing.T) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "11",
		Env: []string{
			"POSTGRES_PASSWORD=postgres",
			"POSTGRES_USER=postgres",
			"POSTGRES_DB=postgres",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://postgres:postgres@%s/postgres?sslmode=disable", hostAndPort)

	log.Println("Connecting to database on url: ", databaseUrl)

	// Tell docker to hard kill the container in 120 seconds
	resource.Expire(120)

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 120 * time.Second
	if err = pool.Retry(func() error {
		db, err = sql.Open("postgres", databaseUrl)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// Server Setup
	var cfg Config
	cfg.Db.Dsn = databaseUrl
	cfg.Auth.Secret = "secret"
	cfg.Db.MaxOpenConn = 25
	cfg.Db.MaxIdleConn = 25
	cfg.Db.MaxIdleTime = "15m"
	cfg.Limiter.Enabled = true
	cfg.Limiter.Rps = 2
	cfg.Limiter.Burst = 4

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := OpenDB(cfg)
	assert.Nil(t, err)
	defer db.Close()

	_, err = db.Exec("" +
		"CREATE EXTENSION IF NOT EXISTS pgcrypto;" +
		"CREATE TABLE IF NOT EXISTS users (" +
		"id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid()," +
		"created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()," +
		"email text UNIQUE NOT NULL," +
		"password_hash bytea NOT NULL," +
		"first_name char varying(100) NOT NULL," +
		"last_name char varying(100) NOT NULL," +
		"activated bool NOT NULL DEFAULT false," +
		"version integer NOT NULL DEFAULT 1);")
	assert.Nil(t, err)

	app := &Application{
		Config: cfg,
		Logger: logger,
		Models: data.InitModels(db),
	}

	ts := httptest.NewTLSServer(app.Routes())
	defer ts.Close()

	// Register
	user := `{"email": "test@example.com", "password": "pa55word", "first_name": "Jon", "last_name": "Doe"}`
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

	// Sign in
	user = `{"email": "test@example.com", "password": "pa55word"}`
	res, err = ts.Client().Post(ts.URL+"/api/authentication", "application/json", bytes.NewReader([]byte(user)))
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusCreated)

	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.Nil(t, err)

	type authType struct {
		Token string `json:"token"`
	}
	var authResult authType
	err = json.Unmarshal(body, &authResult)
	assert.Nil(t, err)
	assert.NotNil(t, authResult.Token)

	// Get a User with the Authorization token
	req, _ := http.NewRequest("GET", ts.URL+"/api/users/me", nil)

	bearer := fmt.Sprintf("Bearer %v", authResult.Token)
	req.Header.Set("Authorization", bearer)

	res, err = ts.Client().Do(req)
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusOK)

	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.Nil(t, err)

	var mUser map[string]data.User
	err = json.Unmarshal(body, &mUser)
	assert.Nil(t, err)
	assert.Equal(t, mUser["user"].Email, "test@example.com")

	// Patch User with the Authorization token
	email := "test@email.com"
	password := "password"
	user = fmt.Sprintf(`{"email": "%v", "password": "%v"}`, email, password)
	req, _ = http.NewRequest("PATCH", ts.URL+"/api/users/"+mUser["user"].ID.String(), bytes.NewReader([]byte(user)))

	bearer = fmt.Sprintf("Bearer %v", authResult.Token)
	req.Header.Set("Authorization", bearer)

	res, err = ts.Client().Do(req)
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusOK)

	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	assert.Nil(t, err)

	err = json.Unmarshal(body, &mUser)
	assert.Nil(t, err)
	assert.Equal(t, mUser["user"].Email, email)

	// Sign in again with a new email and a new password
	res, err = ts.Client().Post(ts.URL+"/api/authentication", "application/json", bytes.NewReader([]byte(user)))
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusCreated)

	err = json.Unmarshal(body, &authResult)
	assert.Nil(t, err)
	assert.NotNil(t, authResult.Token)
}
