package handlers

import (
	"context"
	"net/http"

	"github.com/leodido/go-scim/shared"
)

func CreatePasswordResetRequestHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()

	ri.Status(http.StatusCreated)
	return
}
