package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/martinezmoises/comments/internal/data"
	"github.com/martinezmoises/comments/internal/validator"
)

func (a *applicationDependencies) registerUserHandler(w http.ResponseWriter,
	r *http.Request) {
	// Get the passed in data from the request body and store in a temporary struct
	var incomingData struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := a.readJSON(w, r, &incomingData)

	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	// we will add the password later after we have hashed it
	user := &data.User{
		Username:  incomingData.Username,
		Email:     incomingData.Email,
		Activated: false,
	}

	// hash the password and store it along with the cleartext version
	err = user.Password.Set(incomingData.Password)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	// Perform validation for the User
	v := validator.New()

	data.ValidateUser(v, user)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = a.userModel.Insert(user) // we will add userModel to main later
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			a.failedValidationResponse(w, r, v.Errors)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return

	}

	token, err := a.tokenModel.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := map[string]any{
		"activationToken": token.Plaintext,
		"userID":          user.ID,
	}

	err = a.mailer.Send(user.Email, "user_welcome.tmpl", data)
	if err != nil {
		a.logger.Error(err.Error())
	}

	// Status code 201 resource created
	err = a.writeJSON(w, http.StatusCreated, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Get the body from the request and store in temporary struct
	var incomingData struct {
		TokenPlaintext string `json:"token"`
	}
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}
	// Validate the data
	v := validator.New()
	data.ValidateTokenPlaintext(v, incomingData.TokenPlaintext)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Let's check if the token provided belongs to the user
	// We will implement the GetForToken() method later
	user, err := a.userModel.GetForToken(data.ScopeActivation, incomingData.TokenPlaintext)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			a.failedValidationResponse(w, r, v.Errors)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	user.Activated = true
	err = a.userModel.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			a.editConflictResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// User has been activated so let's delete the activation token to
	// prevent reuse.
	err = a.tokenModel.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Send a response
	data := envelope{
		"user": user,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
