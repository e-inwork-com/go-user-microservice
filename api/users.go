package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/e-inwork-com/go-user-service/internal/data"
	"github.com/e-inwork-com/go-user-service/internal/validator"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// Claims define a claim of JSON Web Token
type Claims struct {
	ID uuid.UUID `json:"id"`
	jwt.RegisteredClaims
}

func (app *Application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email     string `json:"email_t"`
		Password  string `json:"password"`
		FirstName string `json:"first_name_t"`
		LastName  string `json:"last_name_t"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Email:     input.Email,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Activated: true,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.Models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// getUserHandler Function to get a current User
func (app *Application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	// Get the current user as the owner of the Profile
	owner := app.contextGetUser(r)

	// Get user by ID
	user, err := app.Models.Users.GetByID(owner.ID)

	// Check error
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Send a request response
	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// patchUserHandler Function to update a User record
func (app *Application) patchUserHandler(w http.ResponseWriter, r *http.Request) {
	// Get ID from the request parameters
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Get User from the database
	user, err := app.Models.Users.GetByID(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Get the current user
	owner := app.contextGetUser(r)

	// Check if the User has a related to the owner
	// Only the Owner of the User can update the own User
	if user.ID != owner.ID {
		app.notPermittedResponse(w, r)
		return
	}

	// User input
	var input struct {
		Email     *string `json:"email_t"`
		Password  *string `json:"password"`
		FirstName *string `json:"first_name_t"`
		LastName  *string `json:"last_name_t"`
	}

	// Read JSON from input
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Assign input email if exist
	if input.Email != nil {
		user.Email = *input.Email
	}

	// Assign input password if exist
	if input.Password != nil {
		err = user.Password.Set(*input.Password)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	// Assign input FirstName if exist
	if input.FirstName != nil {
		user.FirstName = *input.FirstName
	}

	// Assign input LastName if exist
	if input.LastName != nil {
		user.LastName = *input.LastName
	}

	// Create a Validator
	v := validator.New()

	// Check if the Profile is valid
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Update the User
	err = app.Models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Send back the User to the request response
	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// Func to create a JSON Web Token
func (app *Application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Sign  in with email & password
	// to request a token for the current user
	var input struct {
		Email    string `json:"email_t"`
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
	err = app.writeJSON(w, http.StatusOK, envelope{"token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
