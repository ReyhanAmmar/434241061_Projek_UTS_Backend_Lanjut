package service

import (
	"api-buku-kas/app/model"
	"api-buku-kas/helper"
)

func CanAccessCategory(
	current model.AuthUser,
	ownerID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	if current.UserID == ownerID {
		return true
	}

	return perms.Can(current.Role, anyPermission)
}