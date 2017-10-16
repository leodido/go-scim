package handlers

import (
	"context"
	"net/http"

	"github.com/leodido/go-scim/shared"
)

func CreatePasswordResetRequestHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()
	sch := server.InternalSchema(shared.PasswordResetRequestUrn)
	repo := server.Repository(shared.UserResourceType)

	resource, err := ParseBodyAsResource(r)

	err = server.ValidateType(resource, sch, ctx)
	ErrorCheck(err)

	username := resource.GetData()["username"].(map[string]interface{})["username"].(string)

	user, err := repo.GetByUsername(username)
	ErrorCheck(err)

	id := user.GetData()["id"].(map[string]interface{})["id"].(string)
	version := user.GetData()["version"].(map[string]interface{})["version"].(string)

	user["password"] = resource.GetData()["password"]

	err = repo.Update(id, version, user)
	ErrorCheck(err)

	ri.Status(http.StatusCreated)
	return
}
