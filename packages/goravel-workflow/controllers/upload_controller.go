package controllers

import (
	"goravel/packages/goravel-workflow/services/Upload"

	"github.com/goravel/framework/contracts/http"
	httpfacades "github.com/hulutech-web/http_result"
)

// UploadController 文件上传控制器，处理文件上传相关的 HTTP 请求。
type UploadController struct {
	//Dependent services (依赖服务)
}

// NewUploadController 创建并返回一个新的 UploadController 实例。
// 用于依赖注入和服务初始化。
func NewUploadController() *UploadController {
	return &UploadController{
		//Inject services (注入服务)
	}
}

// Upload 处理文件上传请求。
// 从 HTTP 请求中获取上传文件，调用 UploadService 执行上传逻辑，
// 并返回上传结果（成功时返回文件路径，失败时返回错误信息）。
func (r *UploadController) Upload(ctx http.Context) http.Response {
	// 从请求中获取上传的文件
	file, err := ctx.Request().File("file")
	if err != nil {
		// 获取文件失败，返回上传失败响应
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "上传失败", nil)
	}
	// 调用 UploadService 执行文件上传
	if att, err := Upload.NewUploadService().Upload(ctx, file); err != nil {
		// 上传服务返回错误，返回上传失败响应
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "上传失败", nil)
	} else {
		// 上传成功，返回文件路径
		return httpfacades.NewResult(ctx).Success("上传成功", att.Path)
	}
}
