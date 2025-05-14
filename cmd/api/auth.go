package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/elhambadri2411/social/internal/mailer"
	"github.com/elhambadri2411/social/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type registerUserPayload struct {
	Username string `json:"username" validate:"required,max=100"`
	Email    string `json:"email" validate:"required,email,max=225"`
	Password string `json:"password" validate:"required,max=100"`
}

type userWithToken struct {
	*store.User
	Token string `json:"token"`
}

type createUserTokenPayload struct {
	Email    string `json:"email" validate:"required,email,max=225"`
	Password string `json:"password" validate:"required,max=100"`
}

// registerUserHandle godoc
//
//	@Summary		Registers a user
//	@Description	Registers a user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		registerUserPayload	true	"User credentials"
//	@Success		201		{object}	userWithToken		"User registered"
//	@Failiure		400 {object} error
//	@Failiure		500 {object} error
//	@Router			/authentication/user [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload registerUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
	}

	// hash user password
	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	plainToken := uuid.New().String()

	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	// store user
	err := app.store.UsersRepository.CreateAndInvite(ctx, user, hashToken, app.config.mail.exp)
	if err != nil {
		switch err {
		case store.ErrDuplicateEmail:
		case store.ErrDuplicateUsername:
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}
	userWithToken := userWithToken{
		User:  user,
		Token: plainToken,
	}

	activationUrl := fmt.Sprintf("%s/confirm/%s", app.config.frontendUrl, plainToken)

	isProdEnv := app.config.env == "PROD"
	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: activationUrl,
	}

	// mail
	status, err := app.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, !isProdEnv)
	if err != nil {
		app.logger.Errorw("error sending welcome email", "error", err)

		// rollback user creating if email fails
		if err := app.store.UsersRepository.Delete(ctx, user.ID); err != nil {
			app.logger.Errorw("error deleting user after email sending fails", "error", err)
		}

		app.internalServerError(w, r, err)
		return
	}

	app.logger.Infow("Email sent", "status code", status)

	if err := app.jsonResponse(w, http.StatusCreated, userWithToken); err != nil {
		app.internalServerError(w, r, err)
	}
}

// activateuserHandle godoc
//
//	@Summary		Activates a user
//	@Description	Activates a user
//	@Tags			authentication
//	@Produce		json
//	@Param			token	path		string	true	"Invitation token"
//	@Success		204		{string}	string	"User activated"
//	@Failiure		400 {object} error
//	@Failiure		500 {object} error
//	@Security		ApiKeyAuth
//	@Router			/users/activate/{token} [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	err := app.store.UsersRepository.Activate(r.Context(), token)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, "User activated"); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// createTokenHandler godoc
//
//	@Summary		creates a token
//	@Description	creates a user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload			body		createUserTokenPayload	true	"User credentials"
//	@Success		200				{string}	string					"Token"
//	@Failiure		400 {object} 	error
//	@Failiure		500 {object} 	error
//	@Security
//	@Router	/authentication/token [post]
func (app *application) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	var payload createUserTokenPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	user, err := app.store.UsersRepository.GetByEmail(r.Context(), payload.Email)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.unauthorizedError(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(app.config.auth.token.expiration).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": app.config.auth.token.issuer,
		"aud": app.config.auth.token.issuer,
	}

	token, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, token); err != nil {
		app.internalServerError(w, r, err)
	}
}
