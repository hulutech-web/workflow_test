package controllers

import (
	GlobalModel "goravel/app/models"
	"goravel/packages/goravel-workflow/models"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	httpfacades "github.com/hulutech-web/http_result"
)

type CcController struct{}

func NewCcController() *CcController {
	return &CcController{}
}

// Index 获取当前用户的抄送列表
func (r *CcController) Index(ctx http.Context) http.Response {

	var ccRecords []models.CcRecord
	var user GlobalModel.User
	facades.Auth(ctx).User(&user)
	queries := ctx.Request().Queries()
	res, _ := httpfacades.NewResult(ctx).SearchByParams(queries, map[string]interface{}{"user_id": user.ID}).ResultPagination(&ccRecords, nil)

	//facades.Orm().Query().Model(&models.CcRecord{}).Where("emp_id=?", emp.ID).Order("id desc").Find(&ccRecords)
	return httpfacades.NewResult(ctx).Success("", res)
}

// GetEntryCC 获取某流程的抄送记录
func (r *CcController) GetEntryCC(ctx http.Context) http.Response {
	entry_id := ctx.Request().InputInt("entry_id")
	var ccRecords []models.CcRecord
	facades.Orm().Query().Model(&models.CcRecord{}).Where("entry_id=?", entry_id).Order("id asc").Find(&ccRecords)
	return httpfacades.NewResult(ctx).Success("", ccRecords)
}
