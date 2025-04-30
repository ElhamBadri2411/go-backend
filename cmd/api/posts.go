package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/elhambadri2411/social/internal/store"
	"github.com/go-chi/chi/v5"
)

type postKey string

const postCtx postKey = "post"

type createPostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

type updatePostPayload struct {
	Title   *string `json:"title" validate:"omitempty,max=100"`
	Content *string `json:"content" validate:"omitempty,max=1000"`
}

func (app *application) getPostsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	limitParam := chi.URLParam(r, "limit")
	offsetParam := chi.URLParam(r, "offset")

	if limitParam == "" {
		limitParam = "10"
	}
	if offsetParam == "" {
		offsetParam = "0"
	}

	limit, err := strconv.ParseInt(limitParam, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	offset, err := strconv.ParseInt(offsetParam, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	posts, err := app.store.PostsRepository.GetAll(ctx, limit, offset)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, posts); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getPostByIdHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	comments, err := app.store.CommentsRepository.GetByPostId(r.Context(), post.ID)
	if err != nil {
		app.internalServerError(w, r, err)
	}

	post.Comments = comments

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) deletePostByIdHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	postIdParam := chi.URLParam(r, "postId")
	postId, err := strconv.ParseInt(postIdParam, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.store.PostsRepository.DeleteById(ctx, postId); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) updatePostByIdHandler(w http.ResponseWriter, r *http.Request) {
	var payload updatePostPayload
	post := getPostFromCtx(r)

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if payload.Content != nil {
		post.Content = *payload.Content
	}

	if payload.Title != nil {
		post.Title = *payload.Content
	}

	err := app.store.PostsRepository.UpdateById(r.Context(), post)
	if err != nil {
		app.internalServerError(w, r, err)
	}

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
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

	if err := app.jsonResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) postsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		ctx = context.WithValue(ctx, postCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromCtx(r *http.Request) *store.Post {
	post := r.Context().Value(postCtx)
	return post.(*store.Post)
}
