package controllers

import (
	"goravel/packages/goravel-workflow/models"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	httpfacades "github.com/hulutech-web/http_result"
)

// EntryArchiveController 流程归档控制器，管理已归档的流程实例快照数据
type EntryArchiveController struct{}

// NewEntryArchiveController 创建流程归档控制器实例
func NewEntryArchiveController() *EntryArchiveController {
	return &EntryArchiveController{}
}

// Index returns paginated list of archived entries
// Index 分页查询已归档的流程实例列表
func (r *EntryArchiveController) Index(ctx http.Context) http.Response {
	// 初始化归档记录切片，用于存放查询结果
	archives := []models.EntryArchive{}
	// 获取请求中的所有查询参数（分页、筛选条件等）
	queries := ctx.Request().Queries()
	// 根据查询参数分页查询归档记录，返回标准分页响应
	result, _ := httpfacades.NewResult(ctx).SearchByParams(queries, nil).ResultPagination(&archives, nil)
	return result
}

// Show returns a single archive with all JSON snapshot fields
// Show 查询单条归档记录的完整 JSON 快照数据
func (r *EntryArchiveController) Show(ctx http.Context) http.Response {
	// 从路由参数中获取归档记录 ID
	id := ctx.Request().RouteInt("id")
	var archive models.EntryArchive
	// 根据主键 ID 查询归档记录
	facades.Orm().Query().Model(&models.EntryArchive{}).Where("id", id).First(&archive)
	// 如果未找到记录（ID 为 0），返回 404 错误
	if archive.ID == 0 {
		return httpfacades.NewResult(ctx).Error(http.StatusNotFound, "归档记录不存在", "")
	}
	// 返回查询成功的归档记录数据
	return httpfacades.NewResult(ctx).Success("", archive)
}
