package controllers

import (
	"fmt"

	"goravel/packages/goravel-workflow/models"
	workflow "goravel/packages/goravel-workflow/services/workflow"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/validation"
	httpfacades "github.com/hulutech-web/http_result"
)

// FlowController 流程控制器，负责流程定义（Flow）的 CRUD、设计、发布等操作
type FlowController struct {
	//Dependent services 依赖服务
}

// NewFlowController 创建流程控制器实例
func NewFlowController() *FlowController {
	return &FlowController{
		//Inject services 注入服务
	}
}

// Index 流程列表（分页），支持通过查询参数进行搜索和筛选
func (r *FlowController) Index(ctx http.Context) http.Response {
	flows := []models.Flow{}
	queries := ctx.Request().Queries()
	result, _ := httpfacades.NewResult(ctx).SearchByParams(queries, nil).ResultPagination(&flows)
	return result
}

// List 获取所有已发布的流程列表，用于下拉选择等场景
func (r *FlowController) List(ctx http.Context) http.Response {
	flows := []models.Flow{}
	facades.Orm().Query().Model(&models.Flow{}).Where("is_publish=?", 1).Find(&flows)
	return httpfacades.NewResult(ctx).Success("", flows)
}

// Create 返回创建流程所需的基础数据，包括可选模板列表和流程类型列表
func (r *FlowController) Create(ctx http.Context) http.Response {
	var templates []models.Template
	var flowtypes []models.Flowtype
	facades.Orm().Query().Model(&models.Template{}).Find(&templates)
	facades.Orm().Query().Model(&models.Flowtype{}).Find(&flowtypes)
	return httpfacades.NewResult(ctx).Success("", map[string]any{
		"templates": templates,
		"flowtypes": flowtypes,
	})
}

// Show 根据 ID 查看单个流程的详细信息
func (r *FlowController) Show(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt("id")
	var flow models.Flow
	facades.Orm().Query().Model(&models.Flow{}).Where("id=?", id).Find(&flow)
	return httpfacades.NewResult(ctx).Success("", flow)
}

// Store 保存新创建的流程定义，包含参数校验和数据库写入
func (r *FlowController) Store(ctx http.Context) http.Response {

	// 表单参数校验：流程编号、名称、模板ID、类型ID均为必填
	validator, err := facades.Validation().Make(ctx, map[string]any{
		"flow_no":     ctx.Request().Input("flow_no"),
		"flow_name":   ctx.Request().Input("flow_name"),
		"template_id": ctx.Request().InputInt("template_id"),
		"type_id":     ctx.Request().InputInt("type_id"),
	}, map[string]any{
		"flow_no":     "required",
		"flow_name":   "required",
		"template_id": "required",
		"type_id":     "required",
	}, validation.Messages(map[string]string{
		"flow_no.required":     "编号不能为空",
		"flow_name.required":   "名称不能为空",
		"template_id.required": "模板不能为空",
		"type_id.required":     "类型不能为空",
	}))
	if err != nil {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "参数校验失败", err)
	}
	if validator.Fails() {
		return httpfacades.NewResult(ctx).ValidError("参数错误", validator.Errors().All())
	}
	flow := models.Flow{}
	err = validator.Bind(&flow)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(500, "参数错误", map[string]any{})
	}
	// 校验通过后创建流程记录
	facades.Orm().Query().Model(&models.Flow{}).Create(&flow)
	return httpfacades.NewResult(ctx).Success("创建成功", flow)
}

// Update 根据 ID 更新流程定义信息
func (r *FlowController) Update(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt("id")
	var flow models.Flow
	ctx.Request().Bind(&flow)
	facades.Orm().Query().Model(&models.Flow{}).Where("id=?", id).Update(&flow)
	return httpfacades.NewResult(ctx).Success("保存成功", flow)

}

// Destroy 根据 ID 删除流程定义
func (r *FlowController) Destroy(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt("id")
	facades.Orm().Query().Where("id=?", id).Delete(&models.Flow{})
	return httpfacades.NewResult(ctx).Success("删除成功", nil)
}

// FlowDesign 获取流程设计数据，包含关联的步骤（Processes）和步骤变量（ProcessVars），用于流程图可视化编辑
func (r *FlowController) FlowDesign(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt("id")
	flow := models.Flow{}
	facades.Orm().Query().Model(&models.Flow{}).With("Processes").With("ProcessVars").Where("id=?", id).Find(&flow)
	return httpfacades.NewResult(ctx).Success("", flow)
}

// Publish 发布流程：对流程进行完整的合法性校验，校验通过后将流程标记为已发布状态
// 校验内容包括：开始步骤唯一性、步骤数量、条件分支表达式合法性、连线完整性等
// Publish 发布流程
func (r *FlowController) Publish(ctx http.Context) http.Response {
	flow_id := ctx.Request().InputInt("flow_id")
	flow := models.Flow{}
	facades.Orm().Query().Model(&models.Flow{}).Where("id=?", flow_id).Find(&flow)

	//如果设置了多个个开始步骤 如果设置了多个开始步骤，则发布失败
	// 流程设计要求有且仅有一个开始步骤（position=0）
	process_starts := []models.Process{}
	facades.Orm().Query().Model(&models.Process{}).Where("flow_id=?", flow_id).Where("position=?", 0).Find(&process_starts)
	if len(process_starts) > 1 {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "发布失败，只能设置一个开始步骤", nil)
	}
	// 检查至少有足够的步骤：统计所有流程步骤数量（flows表关联processes）
	// 以前要求 Condition flowlink >= 2，但简单线性流程不需要条件分支也能发布
	// 统计该流程下的步骤总数，至少需要两个步骤才能发布
	processCount, err := facades.Orm().Query().Model(&models.Process{}).Where("flow_id=?", flow_id).Count()
	if err != nil {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "查询失败", err)
	}
	if processCount < 2 {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "发布失败，至少需要两个步骤", nil)
	}

	// Validate condition flowlink expressions for logical coherence
	// 验证条件分支（Condition Flowlink）的表达式的逻辑一致性，防止发布包含无效条件表达式的流程
	{
		type condRow struct {
			ID         int
			ProcessID  int
			Expression string
			Sort       int
		}
		var condFlowlinks []condRow
		// 查询所有条件类型的流程连线
		facades.Orm().Query().Table("flowlinks").
			Join("left join processes on flowlinks.process_id=processes.id").
			Select("flowlinks.id, flowlinks.process_id, flowlinks.expression, flowlinks.sort").
			Where("flowlinks.flow_id=?", flow_id).
			Where("flowlinks.type=?", "Condition").
			Find(&condFlowlinks)

		type flInput = workflow.ConditionFlowlinkEntry
		// 按步骤 ID 分组整理条件分支
		byProcess := make(map[int][]flInput)
		for _, row := range condFlowlinks {
			byProcess[row.ProcessID] = append(byProcess[row.ProcessID], flInput{
				ID:         row.ID,
				Expression: row.Expression,
				Sort:       row.Sort,
			})
		}

		// 对每个步骤的条件分支进行逐一验证
		for processID, links := range byProcess {
			// 跳过只有一条空 expression 的 Condition flowlink（单出口节点残留，不是真正的条件分支）
			if len(links) == 1 && links[0].Expression == "" {
				continue
			}
			// 调用验证服务检查条件表达式
			if err := workflow.ValidateConditionFlowlinks(links); err != nil {
				var process models.Process
				facades.Orm().Query().Model(&models.Process{}).Where("id=?", processID).Find(&process)
				stepName := process.ProcessName
				if stepName == "" {
					stepName = fmt.Sprintf("步骤ID=%d", processID)
				}
				return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError,
					fmt.Sprintf("步骤\"%s\"的条件分支验证失败：%s", stepName, err.Error()), nil)
			}
		}
	}

	// 检查条件分支连线完整性：存在 next_process_id=-1 且 position!=9 的连线表示有步骤未创建完整连线
	fkCount2, err := facades.Orm().Query().Table("flowlinks").
		Join("left join processes on flowlinks.process_id=processes.id").
		Where("flowlinks.flow_id=?", flow_id).
		Where("flowlinks.type=?", "Condition").
		Where("flowlinks.next_process_id=?", -1).
		Where("processes.position != ?", 9).
		Count()
	if err != nil {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "查询失败", err)
	}
	if fkCount2 > 0 {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "发布失败，有步骤没有创建连线", nil)
	}
	type Countf struct {
		Fid uint `json:"fid"`
		Pid uint `json:"pid"`
	}

	// 检查开始步骤（position=0）是否存在流程连线，确保流程有起点
	flowlinkExists, err := facades.Orm().Query().Table("flowlinks").
		Join("left join processes on flowlinks.process_id=processes.id").
		Where("flowlinks.flow_id=?", flow_id).Where("processes.position=?", 0).Exists()
	if err != nil {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "查询失败", err)
	}
	if !flowlinkExists {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "发布失败，请配置从开始步骤出发的流程连线", nil)
	}
	// 查询所有非条件、非开始步骤的流程连线，检查每个步骤的审批权限是否已设置
	flowlinks := []models.Flowlink{}
	facades.Orm().Query().Table("flowlinks").Select("flowlinks.*").
		Join("join processes on flowlinks.process_id=processes.id").
		Where("flowlinks.flow_id=?", flow_id).
		Where("flowlinks.type !=?", "Condition").
		Where("processes.position !=?", 0).
		Find(&flowlinks)
	for _, flowlink := range flowlinks {
		// 检查每个步骤是否存在有效的非条件连线（即审批权限连线）
		cConditionMet, err := facades.Orm().Query().Table("flowlinks").
			Join("join processes on flowlinks.process_id=processes.id").
			Where("flowlinks.flow_id=?", flow_id).
			Where("flowlinks.process_id=?", flowlink.ProcessID).
			Where("flowlinks.type !=?", "Condition").
			Where("processes.position !=?", 0).Exists()
		if err != nil {
			return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "查询失败", err)
		}
		if !cConditionMet {
			return httpfacades.NewResult(ctx).
				Error(http.StatusInternalServerError, "发布失败，请给设置步骤审批权限", nil)
		}
	}

	// 所有校验通过，将流程标记为已发布
	flow.IsPublish = true
	facades.Orm().Query().Model(&models.Flow{}).Where("id=?", flow.ID).Update("is_publish", true)

	return httpfacades.NewResult(ctx).Success("发布成功", flow)
}
