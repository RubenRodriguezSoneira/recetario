package handlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"recipe-app/internal/appmiddleware"
)

// withParams attaches chi route params and, when userID is non-empty, an
// authenticated user id to the request context. It supports handlers that read
// more than one URL parameter (e.g. {id} + {ingredientID}).
func withParams(req *http.Request, params map[string]string, userID string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	if userID != "" {
		ctx = context.WithValue(ctx, appmiddleware.UserIDKey, userID)
	}
	return req.WithContext(ctx)
}
