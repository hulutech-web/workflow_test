package controllers

import (
	"goravel/packages/goravel-workflow/models"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/validation"
	httpfacades "github.com/hulutech-web/http_result"
)

// TemplateController 表单模板控制器，负责模板的增删改查操作
type TemplateController struct {
	//Dependent services 依赖服务
}

// NewTemplateController 创建模板控制器实例
func NewTemplateController() *TemplateController {
	return &TemplateController{
		//Inject services 注入服务
	}
}

// Index 模板列表分页查询，支持通过 URL 查询参数进行搜索和过滤
func (r *TemplateController) Index(ctx http.Context) http.Response {
	temps := []models.Template{}
	queries := ctx.Request().Queries()
	result, _ := httpfacades.NewResult(ctx).SearchByParams(queries, nil).ResultPagination(&temps)
	return result
}

// Option 获取模板下拉选项列表，返回 label/value 格式，用于前端下拉选择组件
func (r *TemplateController) Option(ctx http.Context) http.Response {
	type Option struct {
		Label string `json:"label"`
		Value string `json:"value"`
	}
	opts := []Option{}
	facades.Orm().Query().Model(&models.Template{}).Select("template_name as label,id as value").Scan(&opts)
	return httpfacades.NewResult(ctx).Success("", opts)
}

// Show 查看模板详情，当前未实现
func (r *TemplateController) Show(ctx http.Context) http.Response {
	return nil
}

// Store 新增模板，对模板名称进行必填和长度校验后写入数据库
func (r *TemplateController) Store(ctx http.Context) http.Response {
	// 构建表单验证器，校验模板名称字段
	validator, err := facades.Validation().Make(ctx, map[string]any{
		"template_name": ctx.Request().Input("template_name"),
	}, map[string]any{
		"template_name": "required|max_len:255"},
		validation.Messages(map[string]string{
			"template_name.required": "标题不能为空",
		}))
	if err != nil {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "验证失败", err)
	}
	var template models.Template
	// 验证未通过，返回验证错误信息
	if validator.Fails() {
		return httpfacades.NewResult(ctx).ValidError("验证失败", validator.Errors().All())
	}
	template.TemplateName = ctx.Request().Input("template_name")
	// 将模板数据写入数据库
	err = facades.Orm().Query().Model(&models.Template{}).Create(&template)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "添加失败", err)
	}
	return httpfacades.NewResult(ctx).Success("添加成功", template)
}

// Update 更新模板信息，绑定请求参数后更新模板名称
func (r *TemplateController) Update(ctx http.Context) http.Response {
	template := models.Template{}
	// 将请求参数绑定到模板结构体
	ctx.Request().Bind(&template)
	// 根据主键ID更新模板名称
	facades.Orm().Query().Model(&models.Template{}).Where("id=?", template.ID).Update("template_name", template.TemplateName)
	return httpfacades.NewResult(ctx).Success("更新成功", template)
}

// Destroy 删除模板，先删除关联的表单字段记录，再删除模板本身
func (r *TemplateController) Destroy(ctx http.Context) http.Response {
	idInt := ctx.Request().RouteInt64("id")
	//删除关联 先删除该模板关联的所有表单字段数据
	facades.Orm().Query().Model(&models.TemplateForm{}).Where("template_id=?", idInt).Delete(&models.TemplateForm{})
	//删除模板 再删除模板主记录
	facades.Orm().Query().Model(&models.Template{}).Where("id=?", idInt).Delete(&models.Template{})
	return httpfacades.NewResult(ctx).Success("删除成功", nil)

}
