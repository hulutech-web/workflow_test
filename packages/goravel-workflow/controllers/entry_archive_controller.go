package controllers

import (
	"goravel/packages/goravel-workflow/models"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	httpfacades "github.com/hulutech-web/http_result"
)

type EntryArchiveController struct{}

func NewEntryArchiveController() *EntryArchiveController {
	return &EntryArchiveController{}
}

// Index returns paginated list of archived entries
func (r *EntryArchiveController) Index(ctx http.Context) http.Response {
	archives := []models.EntryArchive{}
	queries := ctx.Request().Queries()
	result, _ := httpfacades.NewResult(ctx).SearchByParams(queries, nil).ResultPagination(&archives, nil)
	return result
}

// Show returns a single archive with all JSON snapshot fields
func (r *EntryArchiveController) Show(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt("id")
	var archive models.EntryArchive
	facades.Orm().Query().Model(&models.EntryArchive{}).Where("id", id).First(&archive)
	if archive.ID == 0 {
		return httpfacades.NewResult(ctx).Error(http.StatusNotFound, "归档记录不存在", "")
	}
	return httpfacades.NewResult(ctx).Success("", archive)
}
