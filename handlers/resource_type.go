package handlers

import (
	"context"
	"net/http"

	"github.com/leodido/go-scim/shared"
)

func GetAllResourceTypeHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()

	repo := server.Repository(shared.ResourceTypeResourceType)
	userResourceType, err := repo.Get(shared.UserResourceType, "")
	ErrorCheck(err)
	groupResourceType, err := repo.Get(shared.GroupResourceType, "")
	ErrorCheck(err)
	passwordPolicyResourceType, err := repo.Get(shared.PasswordPolicyResourceType, "")
	ErrorCheck(err)

	jsonBytes, err := server.MarshalJSON([]interface{}{
		userResourceType.GetData(),
		groupResourceType.GetData(),
		passwordPolicyResourceType.GetData(),
	}, nil, nil, nil)
	ErrorCheck(err)

	ri.Status(http.StatusOK)
	ri.Body(jsonBytes)
	return
}
