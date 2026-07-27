package controllers

import (
	"reflect"
	"strings"

	"goravel/packages/goravel-workflow/controllers/common"
	"goravel/packages/goravel-workflow/models"
	"goravel/packages/goravel-workflow/services/workflow"
	"goravel/packages/goravel-workflow/services/workflow/official_plugins"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/validation"
	httpfacades "github.com/hulutech-web/http_result"
	"github.com/spf13/cast"
)

// EntryController 流程实例控制器，负责流程发起、查看、编辑重发、撤回等操作
type EntryController struct {
	workflow         *workflow.Workflow         // 工作流引擎实例
	dynamicValidator *common.DynamicValidator    // 动态表单验证器
}

// NewEntryController 创建 EntryController 实例，初始化工作流引擎和动态验证器
func NewEntryController() *EntryController {
	return &EntryController{
		workflow:         workflow.NewBaseWorkflow(),
		dynamicValidator: common.NewDynamicValidator(),
	}
}

// Create 获取发起流程页面数据，根据流程 ID 加载流程定义及关联的表单模板
func (r *EntryController) Create(ctx http.Context) http.Response {
	flow_id := ctx.Request().RouteInt("id")
	var flow models.Flow
	// 查询流程及其关联的模板和模板表单字段
	facades.Orm().Query().Model(&models.Flow{}).Where("id", flow_id).
		With("Template.TemplateForms").Find(&flow)
	return httpfacades.NewResult(ctx).Success("", flow)
}

// Index 获取流程实例列表，支持分页查询，并关联加载流程、发起人、当前步骤等信息
func (r *EntryController) Index(ctx http.Context) http.Response {
	entries := []models.Entry{}
	queries := ctx.Request().Queries()
	// 使用分页查询，并预加载关联的 Flow、Emp、Process 数据
	result, _ := httpfacades.NewResult(ctx).SearchByParams(queries, nil).ResultPagination(&entries, []httpfacades.WithConfig{
		{
			Relation: "Flow",
			Callback: nil,
		},
		{
			Relation: "Emp",
			Callback: nil,
		},
		{
			Relation: "Process",
			Callback: nil,
		},
	})
	return result
}

// Show 查看流程实例详情，包含实例数据、审批任务、评论、子流程及抄送记录
func (r *EntryController) Show(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt("id")
	var entry models.Entry
	// 预加载: 表单数据、流程模板、发起人及部门、审批任务（按ID升序）、当前步骤、进入步骤
	facades.Orm().Query().Model(&models.Entry{}).
		With("EntryDatas").
		With("Flow.Template.TemplateForms").
		With("Emp.Dept").
		With("Procs", func(q orm.Query) orm.Query {
			return q.Order("id asc")
		}).
		With("Process").
		With("EnterProcess").
		Where("id", id).Find(&entry)

	var comments []models.ProcComment
	comments, _ = workflow.NewBaseWorkflow().GetComments(uint(id))

	// Load CC records for this entry
	// 加载该流程实例的抄送记录
	var ccRecords []models.CcRecord
	facades.Orm().Query().Model(&models.CcRecord{}).Where("entry_id=?", id).Order("id asc").Find(&ccRecords)

	// Mark current user's CC records as read
	// 将当前登录用户的抄送记录标记为已读
	var emp models.Emp
	facades.Auth(ctx).User(&emp)
	if emp.ID > 0 {
		facades.Orm().Query().Model(&models.CcRecord{}).
			Where("entry_id=? AND emp_id=? AND status=?", id, emp.ID, 0).
			Update("status", 1)
		// Update in-memory slice to reflect read status
		// 更新内存中的切片数据以反映已读状态
		for i := range ccRecords {
			if uint(ccRecords[i].EmpID) == emp.ID {
				ccRecords[i].Status = 1
			}
		}
	}

	// Recursively load child entries with their procs and comments
	// 递归加载子流程实例及其审批任务和评论
	children := r.loadChildrenRecursive(id)

	return httpfacades.NewResult(ctx).Success("", http.Json{
		"entry":      entry,
		"comments":   comments,
		"children":   children,
		"cc_records": ccRecords,
	})
}

// loadChildrenRecursive recursively loads all descendant entries for a parent entry
// loadChildrenRecursive 递归加载某个父流程实例下的所有子流程实例
func (r *EntryController) loadChildrenRecursive(parentID int) []http.Json {
	var children []models.Entry
	// 查询子流程实例，预加载审批任务、步骤、流程、发起人及部门、表单数据
	facades.Orm().Query().Model(&models.Entry{}).
		With("Procs", func(q orm.Query) orm.Query {
			return q.Order("id asc")
		}).
		With("Process").
		With("EnterProcess").
		With("Flow").
		With("Emp.Dept").
		With("EntryDatas").
		Where("pid = ?", parentID).
		Find(&children)

	var result []http.Json
	for _, child := range children {
		comments, _ := workflow.NewBaseWorkflow().GetComments(uint(child.ID))
		// 递归加载更深层级的子流程
		item := http.Json{
			"entry":    child,
			"comments": comments,
			"children": r.loadChildrenRecursive(int(child.ID)),
		}
		result = append(result, item)
	}
	return result
}

// EntryData 获取流程实例的表单数据，如果是子流程则同时查询父流程的表单数据，以及关联的插件配置
func (r *EntryController) EntryData(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt("id")
	var entrydata []models.EntryData
	var entry models.Entry
	query := facades.Orm().Query()
	query.Model(&models.Entry{}).Where("id=?", id).Find(&entry)
	//当时子流程时，需要查找当前流程的父流程
	// 如果是子流程，则同时查询父流程的表单数据，确保数据完整性
	query.Model(&models.EntryData{}).Where("entry_id=?", id).OrWhere("entry_id=?", entry.Pid).Find(&entrydata)

	last_flowlink := models.Flowlink{}
	// 查找进入当前步骤的条件类型连线，用于获取插件配置
	query.Model(&models.Flowlink{}).Where("next_process_id=?", entry.ProcessID).
		Where("type=?", "Condition").Find(&last_flowlink)
	plugin_configs := official_plugins.PluginConfig{}
	//找上一个process
	// 查找上一个步骤的插件配置
	query.Model(&official_plugins.PluginConfig{}).Where("process_id=?", last_flowlink.ProcessID).Find(&plugin_configs)
	return httpfacades.NewResult(ctx).Success("", http.Json{
		"entry":          entry,
		"entrydata":      entrydata,
		"plugin_configs": plugin_configs,
	})
}

// Store 发起流程，保存流程实例。核心步骤：
// 1. 查找第一步流程连线 2. 动态校验表单数据 3. 创建 Entry 及 EntryData
// 4. 初始化第一个审批任务（SetFirstProcessAuditor）
func (r *EntryController) Store(ctx http.Context) http.Response {
	//添加发起节点
	// 获取发起参数
	flow_id := ctx.Request().InputInt("flow_id")
	var user models.Emp
	facades.Auth(ctx).User(&user)

	flowlink := models.Flowlink{}
	// 查找第一步（position=0）的非Condition类型flowlink，用于初始化审批任务
	// 查找流程的第一步（position=0），即发起后的第一个审批节点
	var firstProcessId uint
	facades.Orm().Query().Model(&models.Process{}).Where("flow_id=? AND position=?", flow_id, 0).Pluck("id", &firstProcessId)
	if firstProcessId == 0 {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "未找到第一步流程配置", "")
	}
	var result []models.Flowlink
	// 优先查找非条件类型的连线（即实际的审批人连线）
	facades.Orm().Query().Model(&models.Flowlink{}).
		Where("flow_id=? AND process_id=?", flow_id, firstProcessId).
		Where("(type!=? OR type IS NULL)", "Condition").
		Order("sort ASC").Find(&result)
	if len(result) == 0 {
		// 如果第一步只有 Condition 类型的连线，尝试找所有从第一步出发的 flowlink
		// 如果没有非条件连线，则查询所有从第一步出发的连线
		facades.Orm().Query().Model(&models.Flowlink{}).
			Where("flow_id=? AND process_id=?", flow_id, firstProcessId).
			Order("sort ASC").Find(&result)
	}
	if len(result) == 0 {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "未找到第一步流程配置", "")
	}
	flowlink = result[0]
	if flowlink.ID == 0 {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "未找到第一步流程配置", "")
	}
	// 加载连线的关联数据：步骤和下一步骤
	var withFlowlink models.Flowlink
	facades.Orm().Query().Model(&models.Flowlink{}).Where("id=?", flowlink.ID).
		With("Process").With("NextProcess").First(&withFlowlink)
	//校验提交的数据
	// 根据流程模板动态生成验证规则并校验提交的表单数据
	validRule, validMsg := r.dynamicValidator.DynamicValidate(flow_id)
	msgMap := make(map[string]string)
	for k, v := range validMsg {
		if s, ok := v.(string); ok {
			msgMap[k] = s
		}
	}
	validator, err := facades.Validation().Make(ctx, r.dynamicValidator.DynamicValidateField(ctx), validRule, validation.Messages(msgMap))
	if err != nil {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, err.Error(), "")
	}
	if validator.Fails() {
		return httpfacades.NewResult(ctx).ValidError("", validator.Errors().All())
	}
	query := facades.Orm().Query()
	var entry models.Entry

	all := ctx.Request().All()
	entry.Title = cast.ToString(all["title"])
	// 如果没有 title，取第一个表单字段的值作为标题
	// 如果没有提供标题，自动取第一个表单字段的值作为标题
	if entry.Title == "" {
		for key, val := range all {
			if key != "flow_id" && key != "id" && key != "entry_id" {
				entry.Title = cast.ToString(val)
				break
			}
		}
	}
	entry.FlowID = cast.ToUint(flow_id)
	entry.EmpID = user.ID
	entry.Circle = 1
	entry.Status = models.EntryStatusPending
	err = query.Model(&models.Entry{}).Create(&entry)

	// 重新加载新建的 Entry 及其关联数据
	var withEntry models.Entry
	query.Model(&models.Entry{}).Where("id=?", entry.ID).With("Flow").With("Emp.Dept").With("Procs").With("EnterProcess").
		Find(&withEntry)
	//进程初始化
	//第一步看是否指定审核人
	// 流程初始化：下一步将调用 SetFirstProcessAuditor 确定第一个审批人

	//向entrydata中插入数据 — must be done before SetFirstProcessAuditor for condition evaluation
	// 向 entrydata 中插入表单数据 — 必须在 SetFirstProcessAuditor 之前完成，以便条件分支能够正确评估
	for key, val := range all {
		if key == "flow_id" || key == "id" || key == "entry_id" {
			continue
		} else {
			//判断val的类型，如果是[]string,则转换为解析为字符串
			// 判断字段值类型：如果是数组类型（如多选框），则拼接为逗号分隔的字符串

			if reflect.TypeOf(val).Kind() == reflect.Slice {
				var sliceStr []string
				//将val解析为sliceStr
				// 将 interface{} 切片转换为字符串切片
				for _, v := range val.([]interface{}) {
					sliceStr = append(sliceStr, cast.ToString(v))
				}
				var newVal string
				newVal = strings.Join(sliceStr, ",")
				var entryData models.EntryData
				entryData.FlowID = cast.ToInt(flow_id)
				entryData.EntryID = cast.ToInt(entry.ID)
				entryData.FieldName = key
				entryData.FieldValue = newVal
				query.Model(&models.EntryData{}).Create(&entryData)
			} else {
				// 普通字段值直接存储
				var entryData models.EntryData
				entryData.FlowID = cast.ToInt(flow_id)
				entryData.EntryID = cast.ToInt(entry.ID)
				entryData.FieldName = key
				entryData.FieldValue = cast.ToString(val)
				query.Model(&models.EntryData{}).Create(&entryData)
			}
		}
	}

	// 设置第一步审批人，创建初始审批任务（Proc）
	err = r.workflow.SetFirstProcessAuditor(withEntry, withFlowlink)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, err.Error(), "")
	}
	//流程表单数据插入，需要goravel的验证规则
	// 发起成功，返回新建的流程实例
	return httpfacades.NewResult(ctx).Success("发起成功", entry)
}

// Update 修改被驳回或已撤回的流程实例并重新发起。
// 仅允许 status 为已驳回(-1)或已撤回(-2)的实例进行修改重发。
func (r *EntryController) Update(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt("id")
	query := facades.Orm().Query()

	var entry models.Entry
	query.Model(&models.Entry{}).Where("id=?", id).First(&entry)
	if entry.ID == 0 {
		return httpfacades.NewResult(ctx).Error(http.StatusBadRequest, "流程不存在", "")
	}
	// 仅允许已驳回或已撤回状态的流程进行修改
	if entry.Status != models.EntryStatusRejected && entry.Status != models.EntryStatusRevoked {
		return httpfacades.NewResult(ctx).Error(http.StatusBadRequest, "当前状态不允许修改", "")
	}

	all := ctx.Request().All()
	// 更新标题（如有提供）
	if title := cast.ToString(all["title"]); title != "" {
		entry.Title = title
		query.Model(&models.Entry{}).Where("id=?", entry.ID).Save(&entry)
	}

	// Update entrydatas
	// 更新表单数据：如果数据已存在则更新，否则新增
	for key, val := range all {
		if key == "flow_id" || key == "id" || key == "entry_id" {
			continue
		}
		fieldValue := ""
		// 处理数组类型字段，拼接为逗号分隔字符串
		if reflect.TypeOf(val).Kind() == reflect.Slice {
			var sliceStr []string
			for _, v := range val.([]interface{}) {
				sliceStr = append(sliceStr, cast.ToString(v))
			}
			fieldValue = strings.Join(sliceStr, ",")
		} else {
			fieldValue = cast.ToString(val)
		}
		var existing models.EntryData
		query.Model(&models.EntryData{}).
			Where("entry_id=? AND field_name=?", id, key).
			First(&existing)
		if existing.ID > 0 {
			// 字段已存在，更新值
			query.Model(&models.EntryData{}).Where("id=?", existing.ID).Update("field_value", fieldValue)
		} else {
			// 字段不存在，新增记录
			entryData := models.EntryData{
				FlowID:     cast.ToInt(entry.FlowID),
				EntryID:    cast.ToInt(id),
				FieldName:  key,
				FieldValue: fieldValue,
			}
			query.Model(&models.EntryData{}).Create(&entryData)
		}
	}

	// Resend
	// 重新发起流程：查询符合条件的 Entry 并重新设置审批链
	var entryResend models.Entry
	query.Model(&models.Entry{}).Where("id=?", id).Where("status IN (?, ?)", models.EntryStatusRejected, models.EntryStatusRevoked).
		With("Flow").With("Emp.Dept").With("Procs").With("EnterProcess").Find(&entryResend)
	if entryResend.ID == 0 {
		return httpfacades.NewResult(ctx).Error(http.StatusBadRequest, "当前状态不允许重发", "")
	}

	// 校验流程是否已发布
	flow := models.Flow{}
	query.Model(&models.Flow{}).Where("id=?", entryResend.FlowID).Where("is_publish=?", true).Find(&flow)
	if flow.ID == 0 {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "流程未发布", "请检查")
	}
	// 查找第一步的 flowlink，重新初始化审批链
	var flowlink models.Flowlink
	var firstProcessId uint
	query.Model(&models.Process{}).Where("flow_id=? AND position=?", entryResend.FlowID, 0).Pluck("id", &firstProcessId)
	if firstProcessId == 0 {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "未找到第一步流程配置", "")
	}
	query.Model(&models.Flowlink{}).
		Where("flow_id=? AND type!=? AND process_id=?", entryResend.FlowID, "Condition", firstProcessId).
		Order("sort ASC").First(&flowlink)
	if flowlink.ID == 0 {
		query.Model(&models.Flowlink{}).
			Where("flow_id=? AND process_id=?", entryResend.FlowID, firstProcessId).
			Order("sort ASC").First(&flowlink)
	}
	if flowlink.ID == 0 {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "节点关系错误", "请检查")
	}
	var withFlowlink models.Flowlink
	query.Model(&models.Flowlink{}).Where("id=?", flowlink.ID).With("Process").With("NextProcess").Find(&withFlowlink)

	// 更新 Entry 状态：轮次 +1，重置为待审批状态
	var map_entry = make(map[string]interface{})
	map_entry["circle"] = entryResend.Circle + 1
	map_entry["child"] = 0
	map_entry["status"] = models.EntryStatusPending
	query.Model(&models.Entry{}).Where("id=?", entryResend.ID).Update(map_entry)
	newEntry := models.Entry{}
	query.Model(&models.Entry{}).Where("id=?", entryResend.ID).With("Flow").With("Emp.Dept").With("Procs").With("EnterProcess").Find(&newEntry)

	// 重新初始化第一步审批人
	err := r.workflow.SetFirstProcessAuditor(newEntry, withFlowlink)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, err.Error(), "")
	}
	return httpfacades.NewResult(ctx).Success("修改并重发成功", entryResend)
}

// Destroy 删除流程实例（暂未实现）
func (r *EntryController) Destroy(ctx http.Context) http.Response {
	return nil
}

// 重发
// Resend 重发已驳回的流程实例。与 Update 类似，但不修改表单数据，仅重新发起审批链。
func (r *EntryController) Resend(ctx http.Context) http.Response {
	entry_id := ctx.Request().Input("entry_id")
	entry := models.Entry{}
	query := facades.Orm().Query()
	// 查询已驳回状态的 Entry，并预加载关联数据
	query.Model(&models.Entry{}).Where("id=?", entry_id).Where("status=?", models.EntryStatusRejected).With("Flow").With("Emp.Dept").With("Procs").With("EnterProcess").
		Find(&entry)

	// 校验流程是否已发布
	flow := models.Flow{}

	query.Model(&models.Flow{}).Where("id=?", entry.FlowID).Where("is_publish=?", true).Find(&flow)
	if flow.ID == 0 {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "流程未发布，请检查", "")
	}
	// 查找第一步的 flowlink
	var flowlink models.Flowlink
	var firstProcessId uint
	facades.Orm().Query().Model(&models.Process{}).Where("flow_id=? AND position=?", entry.FlowID, 0).Pluck("id", &firstProcessId)
	if firstProcessId == 0 {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "未找到第一步流程配置", "")
	}
	facades.Orm().Query().Model(&models.Flowlink{}).
		Where("flow_id=? AND type!=? AND process_id=?", entry.FlowID, "Condition", firstProcessId).
		Order("sort ASC").First(&flowlink)
	if flowlink.ID == 0 {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "节点关系错误，请检查", "")
	}
	var withFlowlink models.Flowlink
	facades.Orm().Query().Model(&models.Flowlink{}).Where("id=?", flowlink.ID).
		With("Process").With("NextProcess").Find(&withFlowlink)
	//零值更新
	// 重置 Entry：轮次 +1，子流程标记清零，状态恢复为待审批
	var map_entry = make(map[string]interface{})
	map_entry["circle"] = entry.Circle + 1
	map_entry["child"] = 0
	map_entry["status"] = models.EntryStatusPending
	query.Model(&models.Entry{}).Where("id=?", entry.ID).Update(map_entry)
	newEntry := models.Entry{}
	query.Model(&models.Entry{}).Where("id=?", entry.ID).With("Flow").With("Emp.Dept").With("Procs").With("EnterProcess").Find(&newEntry)

	// 初始化第一步审批人
	err := r.workflow.SetFirstProcessAuditor(newEntry, withFlowlink)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "系统错误，请检查", "")
	}
	return httpfacades.NewResult(ctx).Success("重发成功", entry)
}

// Revoke 撤回流程
// Revoke 撤回流程实例。仅允许发起人撤回收到的审批任务尚未被处理的流程。
func (r *EntryController) Revoke(ctx http.Context) http.Response {
	var user models.Emp
	facades.Auth(ctx).User(&user)
	entry_id := ctx.Request().InputInt("entry_id")
	// 调用工作流引擎的撤回方法，引擎内部会校验权限和状态
	err := r.workflow.Revoke(uint(entry_id), user)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(500, "撤回失败", err.Error())
	}
	return httpfacades.NewResult(ctx).Success("撤回成功", nil)
}
