package service

import (
	"strings"

	"api-buku-kas/app/model"
	"api-buku-kas/helper"
)

func IsEmptyCategoryPatch(req model.PatchCategoryRequest) bool {
	return req.Name == nil && req.Description == nil
}

func ApplyCategoryPatch(
	current model.Category,
	req model.PatchCategoryRequest,
) (model.Category, map[string]string) {
	updated := current

	if req.Name != nil {
		updated.Name = strings.TrimSpace(*req.Name)
	}

	if req.Description != nil {
		updated.Description = strings.TrimSpace(*req.Description)
	}

	errs := helper.ValidateStruct(model.ReplaceCategoryRequest{
		Name:        updated.Name,
		Description: updated.Description,
	})

	return updated, errs
}