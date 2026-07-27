package controllers

import (
	"goravel/packages/goravel-workflow/models"
	"goravel/packages/goravel-workflow/requests"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	httpfacades "github.com/hulutech-web/http_result"
)

// TemplateformController 模板表单控制器，负责管理流程模板中的表单字段配置。
// 每个流程模板（Template）可以关联多个表单字段（TemplateForm），
// 用于定义发起流程时需要填写的表单内容（字段名、字段类型、校验规则等）。
type TemplateformController struct {
	//Dependent services
	// 依赖的服务
}

// NewTemplateformController 创建 TemplateformController 实例
func NewTemplateformController() *TemplateformController {
	return &TemplateformController{
		//Inject services
		// 注入服务
	}
}

// Index 获取指定模板下的所有表单字段列表
// 根据路由参数中的模板 ID (template_id) 查询关联的表单字段，
// 按 sort 升序、id 降序排列返回
func (r *TemplateformController) Index(ctx http.Context) http.Response {
	template_id := ctx.Request().RouteInt("id")
	template_forms := []models.TemplateForm{}
	// 查询该模板下所有表单字段，按排序号升序、ID降序排列
	facades.Orm().Query().Model(&models.TemplateForm{}).Where("template_id=?", template_id).
		Order("sort asc").Order("id desc").Find(&template_forms)
	return httpfacades.NewResult(ctx).Success("", template_forms)
}

// Show 查看单个模板表单字段的详细信息
// 根据路由参数中的表单字段 ID 查询并返回
func (r *TemplateformController) Show(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt("id")
	var templateform models.TemplateForm
	facades.Orm().Query().Model(&models.TemplateForm{}).Where("id=?", id).Find(&templateform)
	return httpfacades.NewResult(ctx).Success("", templateform)
}

// Store 创建新的模板表单字段
// 使用 TemplateformRequest 对请求参数进行校验，
// 校验通过后将字段配置信息写入数据库
func (r *TemplateformController) Store(ctx http.Context) http.Response {
	// 旧版验证逻辑（已注释，保留以备参考）
	// 使用框架内置的 Validator 进行字段校验，包含以下规则：
	// - template_id: 必填，关联的模板ID
	// - field: 必填且只允许字母和规则格式，字段标识名
	// - field_name: 必填，字段显示标题
	// - field_type: 必填，字段类型（如 text、select、date 等）
	//validator, _ := facades.Validation().Make(map[string]any{
	//	"template_id":         ctx.Request().InputInt("template_id"),
	//	"field":               ctx.Request().Input("field"),
	//	"field_name":          ctx.Request().Input("field_name"),
	//	"field_type":          ctx.Request().Input("field_type"),
	//	"field_value":         ctx.Request().Input("field_value"),
	//	"field_default_value": ctx.Request().Input("field_default_value"),
	//	"field_rules":         ctx.Request().Input("field_rules"),
	//	"sort":                ctx.Request().InputInt("sort"),
	//}, map[string]string{
	//	"template_id": "required",
	//	"field":       "required|alpha_rule",
	//	"field_name":  "required",
	//	"field_type":  "required",
	//}, validation.Messages(map[string]string{
	//	"template_id.required": "模板不能为空",
	//	"field.required":       "字段名称不能为空",
	//	"field_name.required":  "字段标题不能为空",
	//	"field_type.required":  "字段类型不能为空",
	//}))
	//if validator.Fails() {
	//	return httpfacades.NewResult(ctx).ValidError("参数错误", validator.Errors().All())
	//}

	// 使用 Request 结构体进行参数绑定和校验
	var templateformRequest requests.TemplateformRequest
	errors, err := ctx.Request().ValidateRequest(&templateformRequest)
	if errors != nil || err != nil {
		return httpfacades.NewResult(ctx).ValidError("参数错误", errors.All())
	}

	// 将请求参数映射到模型结构体
	tpform := models.TemplateForm{}
	tpform.TemplateID = templateformRequest.TemplateID // 关联模板ID
	tpform.Field = templateformRequest.Field           // 字段标识名
	tpform.FieldName = templateformRequest.FieldName   // 字段显示标题
	tpform.FieldType = templateformRequest.FieldType   // 字段类型
	tpform.FieldValue = templateformRequest.FieldValue            // 字段可选值列表
	tpform.FieldDefaultValue = templateformRequest.FieldDefaultValue // 字段默认值
	tpform.Sort = templateformRequest.Sort                             // 排序号
	tpform.FieldRules = templateformRequest.FieldRules                 // 字段校验规则
	if err != nil {
		return httpfacades.NewResult(ctx).ValidError("", errors.All())
	}

	// 写入数据库
	facades.Orm().Query().Model(&models.TemplateForm{}).Create(&tpform)
	return httpfacades.NewResult(ctx).Success("创建成功", nil)
}

// Update 更新模板表单字段
// 根据路由参数中的字段 ID 查找已有记录，
// 校验请求参数后更新字段配置并保存
func (r *TemplateformController) Update(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt("id")

	// 参数校验
	var templateformRequest requests.TemplateformRequest
	errors, err := ctx.Request().ValidateRequest(&templateformRequest)
	if errors != nil || err != nil {
		return httpfacades.NewResult(ctx).ValidError("参数错误", errors.All())
	}

	// 查找已有记录
	existTpform := models.TemplateForm{}
	facades.Orm().Query().Model(&models.TemplateForm{}).Where("id=?", id).Find(&existTpform)

	// 将请求参数更新到已有模型
	existTpform.TemplateID = templateformRequest.TemplateID // 关联模板ID
	existTpform.Field = templateformRequest.Field           // 字段标识名
	existTpform.FieldName = templateformRequest.FieldName   // 字段显示标题
	existTpform.FieldType = templateformRequest.FieldType   // 字段类型
	existTpform.FieldValue = templateformRequest.FieldValue            // 字段可选值列表
	existTpform.FieldDefaultValue = templateformRequest.FieldDefaultValue // 字段默认值
	existTpform.Sort = templateformRequest.Sort                             // 排序号
	existTpform.FieldRules = templateformRequest.FieldRules                 // 字段校验规则
	if err != nil {
		return httpfacades.NewResult(ctx).Error(500, "参数错误", map[string]any{})
	}

	// 保存更新到数据库
	facades.Orm().Query().Save(&existTpform)
	return httpfacades.NewResult(ctx).Success("修改成功", nil)
}

// Destroy 删除模板表单字段
// 根据路由参数中的字段 ID 删除对应的模板表单字段记录
func (r *TemplateformController) Destroy(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt("id")
	facades.Orm().Query().Model(&models.TemplateForm{}).Where("id=?", id).Delete(&models.TemplateForm{})
	return httpfacades.NewResult(ctx).Success("删除成功", nil)
}

// FlowTemplateForm 根据流程ID获取其关联模板的表单字段列表
// 通过 flow_id 查找对应的流程（Flow），再通过流程关联的模板（Template）
// 查询该模板下的所有表单字段配置，用于前端动态渲染发起流程时的表单
func (r *TemplateformController) FlowTemplateForm(ctx http.Context) http.Response {
	flow_id := ctx.Request().InputInt("flow_id")

	// 查询流程及其关联的模板
	var flow models.Flow
	facades.Orm().Query().Model(&models.Flow{}).Where("id=?", flow_id).With("Template").Find(&flow)

	// 获取模板ID，查询该模板下的所有表单字段
	template_id := flow.TemplateID
	var template_forms []models.TemplateForm
	facades.Orm().Query().Model(&models.TemplateForm{}).Where("template_id=?", template_id).Find(&template_forms)
	return httpfacades.NewResult(ctx).Success("", template_forms)
}
