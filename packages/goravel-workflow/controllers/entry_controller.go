package controllers

import (
	"reflect"
	"strings"

	"goravel/packages/goravel-workflow/controllers/common"
	"goravel/packages/goravel-workflow/models"
	"goravel/packages/goravel-workflow/services/workflow"
	"goravel/packages/goravel-workflow/services/workflow/official_plugins"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/validation"
	httpfacades "github.com/hulutech-web/http_result"
	"github.com/spf13/cast"
)

type EntryController struct {
	workflow         *workflow.Workflow
	dynamicValidator *common.DynamicValidator
}

func NewEntryController() *EntryController {
	return &EntryController{
		workflow:         workflow.NewBaseWorkflow(),
		dynamicValidator: common.NewDynamicValidator(),
	}
}

func (r *EntryController) Create(ctx http.Context) http.Response {
	flow_id := ctx.Request().RouteInt("id")
	var flow models.Flow
	facades.Orm().Query().Model(&models.Flow{}).Where("id", flow_id).
		With("Template.TemplateForms").Find(&flow)
	return httpfacades.NewResult(ctx).Success("", flow)
}

func (r *EntryController) Index(ctx http.Context) http.Response {
	return nil
}

func (r *EntryController) Show(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt("id")
	var entry models.Entry
	facades.Orm().Query().Model(&models.Entry{}).With("EntryDatas").With("Flow.Template.TemplateForms").Where("id", id).Find(&entry)
	return httpfacades.NewResult(ctx).Success("", entry)
}

func (r *EntryController) EntryData(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt("id")
	var entrydata []models.EntryData
	var entry models.Entry
	query := facades.Orm().Query()
	query.Model(&models.Entry{}).Where("id=?", id).Find(&entry)
	//当时子流程时，需要查找当前流程的父流程
	query.Model(&models.EntryData{}).Where("entry_id=?", id).OrWhere("entry_id=?", entry.Pid).Find(&entrydata)

	last_flowlink := models.Flowlink{}
	query.Model(&models.Flowlink{}).Where("next_process_id=?", entry.ProcessID).
		Where("type=?", "Condition").Find(&last_flowlink)
	plugin_configs := official_plugins.PluginConfig{}
	//找上一个process
	query.Model(&official_plugins.PluginConfig{}).Where("process_id=?", last_flowlink.ProcessID).Find(&plugin_configs)
	return httpfacades.NewResult(ctx).Success("", http.Json{
		"entry":          entry,
		"entrydata":      entrydata,
		"plugin_configs": plugin_configs,
	})
}

func (r *EntryController) Store(ctx http.Context) http.Response {
	//添加发起节点
	flow_id := ctx.Request().InputInt("flow_id")
	var user models.Emp
	facades.Auth(ctx).User(&user)

	flowlink := models.Flowlink{}
	// 查找第一步（position=0）的非Condition类型flowlink，用于初始化审批任务
	var firstProcessId uint
	facades.Orm().Query().Model(&models.Process{}).Where("flow_id=? AND position=?", flow_id, 0).Pluck("id", &firstProcessId)
	if firstProcessId == 0 {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "未找到第一步流程配置", "")
	}
	var result []models.Flowlink
	facades.Orm().Query().Model(&models.Flowlink{}).
		Where("flow_id=? AND process_id=?", flow_id, firstProcessId).
		Where("(type!=? OR type IS NULL)", "Condition").
		Order("sort ASC").Find(&result)
	if len(result) == 0 {
		// 如果第一步只有 Condition 类型的连线，尝试找所有从第一步出发的 flowlink
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
	var withFlowlink models.Flowlink
	facades.Orm().Query().Model(&models.Flowlink{}).Where("id=?", flowlink.ID).
		With("Process").With("NextProcess").First(&withFlowlink)
	//校验提交的数据
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
	entry.FlowID = cast.ToUint(flow_id)
	entry.EmpID = user.ID
	entry.Circle = 1
	entry.Status = models.EntryStatusPending
	err = query.Model(&models.Entry{}).Create(&entry)

	var withEntry models.Entry
	query.Model(&models.Entry{}).Where("id=?", entry.ID).With("Flow").With("Emp.Dept").With("Procs").With("EnterProcess").
		Find(&withEntry)
	//进程初始化
	//第一步看是否指定审核人

	//向entrydata中插入数据 — must be done before SetFirstProcessAuditor for condition evaluation
	for key, val := range all {
		if key == "flow_id" || key == "id" || key == "entry_id" {
			continue
		} else {
			//判断val的类型，如果是[]string,则转换为解析为字符串

			if reflect.TypeOf(val).Kind() == reflect.Slice {
				var sliceStr []string
				//将val解析为sliceStr
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
				var entryData models.EntryData
				entryData.FlowID = cast.ToInt(flow_id)
				entryData.EntryID = cast.ToInt(entry.ID)
				entryData.FieldName = key
				entryData.FieldValue = cast.ToString(val)
				query.Model(&models.EntryData{}).Create(&entryData)
			}
		}
	}

	err = r.workflow.SetFirstProcessAuditor(withEntry, withFlowlink)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, err.Error(), "")
	}
	//流程表单数据插入，需要goravel的验证规则
	return httpfacades.NewResult(ctx).Success("发起成功", entry)
}

func (r *EntryController) Update(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt("id")
	query := facades.Orm().Query()

	var entry models.Entry
	query.Model(&models.Entry{}).Where("id=?", id).First(&entry)
	if entry.ID == 0 {
		return httpfacades.NewResult(ctx).Error(http.StatusBadRequest, "流程不存在", "")
	}
	if entry.Status != models.EntryStatusRejected && entry.Status != models.EntryStatusRevoked {
		return httpfacades.NewResult(ctx).Error(http.StatusBadRequest, "当前状态不允许修改", "")
	}

	all := ctx.Request().All()
	if title := cast.ToString(all["title"]); title != "" {
		entry.Title = title
		query.Model(&models.Entry{}).Where("id=?", entry.ID).Save(&entry)
	}

	// Update entrydatas
	for key, val := range all {
		if key == "flow_id" || key == "id" || key == "entry_id" {
			continue
		}
		fieldValue := ""
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
			query.Model(&models.EntryData{}).Where("id=?", existing.ID).Update("field_value", fieldValue)
		} else {
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
	var entryResend models.Entry
	query.Model(&models.Entry{}).Where("id=?", id).Where("status IN (?, ?)", models.EntryStatusRejected, models.EntryStatusRevoked).
		With("Flow").With("Emp.Dept").With("Procs").With("EnterProcess").Find(&entryResend)
	if entryResend.ID == 0 {
		return httpfacades.NewResult(ctx).Error(http.StatusBadRequest, "当前状态不允许重发", "")
	}

	flow := models.Flow{}
	query.Model(&models.Flow{}).Where("id=?", entryResend.FlowID).Where("is_publish=?", true).Find(&flow)
	if flow.ID == 0 {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "流程未发布", "请检查")
	}
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

	var map_entry = make(map[string]interface{})
	map_entry["circle"] = entryResend.Circle + 1
	map_entry["child"] = 0
	map_entry["status"] = models.EntryStatusPending
	query.Model(&models.Entry{}).Where("id=?", entryResend.ID).Update(map_entry)
	newEntry := models.Entry{}
	query.Model(&models.Entry{}).Where("id=?", entryResend.ID).With("Flow").With("Emp.Dept").With("Procs").With("EnterProcess").Find(&newEntry)

	err := r.workflow.SetFirstProcessAuditor(newEntry, withFlowlink)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, err.Error(), "")
	}
	return httpfacades.NewResult(ctx).Success("修改并重发成功", entryResend)
}

func (r *EntryController) Destroy(ctx http.Context) http.Response {
	return nil
}

// 重发
func (r *EntryController) Resend(ctx http.Context) http.Response {
	entry_id := ctx.Request().Input("entry_id")
	entry := models.Entry{}
	query := facades.Orm().Query()
	query.Model(&models.Entry{}).Where("id=?", entry_id).Where("status=?", models.EntryStatusRejected).With("Flow").With("Emp.Dept").With("Procs").With("EnterProcess").
		Find(&entry)

	flow := models.Flow{}

	query.Model(&models.Flow{}).Where("id=?", entry.FlowID).Where("is_publish=?", true).Find(&flow)
	if flow.ID == 0 {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "流程未发布，请检查", "")
	}
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
	var map_entry = make(map[string]interface{})
	map_entry["circle"] = entry.Circle + 1
	map_entry["child"] = 0
	map_entry["status"] = models.EntryStatusPending
	query.Model(&models.Entry{}).Where("id=?", entry.ID).Update(map_entry)
	newEntry := models.Entry{}
	query.Model(&models.Entry{}).Where("id=?", entry.ID).With("Flow").With("Emp.Dept").With("Procs").With("EnterProcess").Find(&newEntry)

	err := r.workflow.SetFirstProcessAuditor(newEntry, withFlowlink)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "系统错误，请检查", "")
	}
	return httpfacades.NewResult(ctx).Success("重发成功", entry)
}

// Revoke 撤回流程
func (r *EntryController) Revoke(ctx http.Context) http.Response {
	var user models.Emp
	facades.Auth(ctx).User(&user)
	entry_id := ctx.Request().InputInt("entry_id")
	err := r.workflow.Revoke(uint(entry_id), user)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(500, "撤回失败", err.Error())
	}
	return httpfacades.NewResult(ctx).Success("撤回成功", nil)
}
