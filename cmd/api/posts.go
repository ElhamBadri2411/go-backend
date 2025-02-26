package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/elhambadri2411/social/internal/store"
	"github.com/go-chi/chi/v5"
)

type createPostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

func (app *application) getPostByIdHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	postIdParam := chi.URLParam(r, "postId")
	postId, err := strconv.ParseInt(postIdParam, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	post, err := app.store.PostsRepository.GetById(ctx, postId)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := writeJSON(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) createPostsHandler(w http.ResponseWriter, r *http.Request) {
	var payload createPostPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		// TODO: change after auth
		UserId: 1,
	}

	ctx := r.Context()
	if err := app.store.PostsRepository.Create(ctx, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := writeJSON(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
