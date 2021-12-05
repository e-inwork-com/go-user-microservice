package api

import (
	"errors"
	"github.com/google/uuid"
	"net/http"
	"time"

	"github.com/e-inwork-com/golang-user-microservice/internal/validator"
	"github.com/e-inwork-com/golang-user-microservice/pkg/data"

	"github.com/golang-jwt/jwt/v4"
)

// Claims define a claim of JSON Web Token
type Claims struct {
	ID		uuid.UUID	`json:"id"`
	jwt.RegisteredClaims
}

// Func to create a JSON Web Token
func (app *Application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Sign  in with email & password
	// to request a token for the current user
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// The JSON should be match with define input
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Validate the email and the password
	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Get the user by the input email
	user, err := app.Models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Check if the inout password is match
	// with the existing password in the database
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	// Set Signing Key from the Config Environment
	signingKey := []byte(app.Config.Auth.Secret)

	// Set an expired time for a week
	expirationTime := time.Now().Add((24 * 7) * time.Hour)

	// Set the ID of the user in the Claim token
	claims := &Claims{
		ID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Create a signed token
	signed := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := signed.SignedString(signingKey)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Set response with a token
	err = app.writeJSON(w, http.StatusCreated, envelope{"token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
