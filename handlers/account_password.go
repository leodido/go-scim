package handlers

import (
	"context"
	"fmt"
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

	username := resource.GetData()["username"].(string)

	fmt.Println("username ", username)

	list, err := repo.Search(shared.SearchRequest{
		Filter:     fmt.Sprintf("userName eq \"%s\"", username),
		StartIndex: 1,
	})

	// TODO: Check returned list
	ref := list.Resources[0]
	user := list.Resources[0].GetData()

	id := user["id"].(string)
	version := user["meta"].(map[string]interface{})["version"].(string)
	newPassword := resource.GetData()["password"].(string)

	user["password"] = newPassword

	fmt.Println("New password ", newPassword)

	err = repo.Update(id, version, ref)
	ErrorCheck(err)

	ri.Status(http.StatusCreated)
	return
}
