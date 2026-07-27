package controllers

import (
	"encoding/json"
	"fmt"
	"strings"

	"goravel/packages/goravel-workflow/controllers/common"
	models "goravel/packages/goravel-workflow/models"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	httpfacades "github.com/hulutech-web/http_result"
	"github.com/spf13/cast"
)

// ProcessController 流程步骤控制器，负责工作流中各个审批步骤（Process）的增删改查。
// 包括步骤的创建、更新、删除、属性查看以及条件路由的设置。
type ProcessController struct {
	//Dependent services
	// 依赖的服务
}

// NewProcessController 创建 ProcessController 实例，用于注入依赖服务。
func NewProcessController() *ProcessController {
	return &ProcessController{
		//Inject services
		// 注入服务
	}
}

// Index 步骤列表首页（预留扩展，当前返回 nil）。
func (r *ProcessController) Index(ctx http.Context) http.Response {
	return nil
}

// List 根据 flow_id 查询某个流程下的所有步骤（Process），包含关联的 Flow 信息。
func (r *ProcessController) List(ctx http.Context) http.Response {
	flow_id := ctx.Request().InputInt("flow_id")
	processes := []models.Process{}
	facades.Orm().Query().Model(&models.Process{}).
		Where("flow_id=?", flow_id).With("Flow").Model(&models.Process{}).Find(&processes)
	return httpfacades.NewResult(ctx).Success("", processes)
}

// Show 步骤详情（预留扩展，当前返回 nil）。
func (r *ProcessController) Show(ctx http.Context) http.Response {
	return nil
}

// Store 在流程画布上创建新的步骤节点（Process），并同步更新流程的 jsPlumb 拓扑数据。
// 分为两个步骤：
//   步骤一：在数据库中创建 Process 记录，包含位置、样式等默认值
//   步骤二：将该节点信息写入 Flow.Jsplumb JSON 数据，保持前后端状态一致
func (r *ProcessController) Store(ctx http.Context) http.Response {
	flow_id := ctx.Request().InputInt("flow_id")
	left := ctx.Request().Input("left")
	top := ctx.Request().Input("top")
	// 开启数据库事务，确保步骤创建与流程数据更新原子执行
	tx, _ := facades.Orm().Query().Begin()
	var process models.Process
	var flow models.Flow
	tx.Model(&models.Flow{}).Where("id=?", flow_id).Find(&flow)

	//步骤一：创建新的流程步骤，设置默认名称、样式尺寸及画布位置
	process.FlowID = flow_id
	process.ProcessName = "新建流程"
	process.StyleWidth = 200
	process.StyleHeight = 48
	process.Style = fmt.Sprintf("width:200px;height:48px;line-height:30px;color:#66CDAA;left:%s;top:%s;", left, top)
	process.PositionLeft = left
	process.PositionTop = top
	if err := tx.Model(&models.Process{}).Create(&process); err != nil {
		tx.Rollback()
	}

	//步骤二：将新节点注册到流程的 jsPlumb 拓扑结构中
	jsMap := common.Plumb{}
	if flow.Jsplumb == "" {
		// 流程尚无节点拓扑数据：初始化 Jsplumb JSON，total 置 1，放入第一个节点
		//添加属性
		jsMap.Total = 1
		jsMap.List = map[string]common.Node{}
		listMap := map[string]common.Node{}
		node := common.Node{
			ID:          cast.ToInt(process.ID),
			FlowId:      process.FlowID,
			ProcessName: process.ProcessName,
			ProcessTo:   "",
			Icon:        "",
			Style:       process.Style,
		}
		listMap[cast.ToString(process.ID)] = node
		jsMap.List = listMap
		strByte, _ := json.Marshal(jsMap)
		tx.Model(&models.Flow{}).Where("id=?", flow_id).Update("jsplumb", strByte)
		flow.IsPublish = false
		tx.Model(&models.Flow{}).Where("id=?", flow_id).Update(&flow)
		tx.Commit()
		return httpfacades.NewResult(ctx).Success("", http.Json{
			"id":           process.ID,
			"flow_id":      process.FlowID,
			"process_name": process.ProcessName,
			"process_to":   "",
			"icon":         "",
			"style":        process.Style,
		})
	} else {
		// 流程已有节点拓扑数据：将已有 JSON 反序列化，追加新节点后再序列化保存
		//jsMap的list属性为二维数组
		var jsMapTemp common.Plumb
		jsMapTemp.List = map[string]common.Node{}
		//将flow中的Jsplumb转换为jsMapTemp
		json.Unmarshal([]byte(flow.Jsplumb), &jsMapTemp)
		node := common.Node{
			ID:          cast.ToInt(process.ID),
			FlowId:      process.FlowID,
			ProcessName: process.ProcessName,
			ProcessTo:   "",
			Icon:        "",
			Style:       process.Style,
		}
		jsMapTemp.List[cast.ToString(process.ID)] = node
		jsMap = jsMapTemp
		//转换jsMap为json
		strByte, _ := json.Marshal(jsMap)
		flow.Jsplumb = string(strByte)
		flow.IsPublish = false
		tx.Model(&models.Flow{}).Where("id=?", flow_id).Update(&flow)
		tx.Commit()
		return httpfacades.NewResult(ctx).Success("", http.Json{
			"id":           process.ID,
			"flow_id":      process.FlowID,
			"process_name": process.ProcessName,
			"process_to":   "",
			"icon":         "",
			"style":        process.Style,
		})
	}
}

// Update 更新流程步骤的完整配置，包括：
//   - 基础属性：名称、颜色、尺寸、图标、位置
//   - 步骤位置类型（起始/普通/子流程）
//   - 子流程配置（子流程ID、完成后行为、返回步骤）
//   - 时间限制与抄送人
//   - 同步更新 jsPlumb 拓扑 JSON 数据
//   - 审批人权限设置（系统自动/指定部门/指定员工）
//   - 条件分支表达式的分组写入
func (r *ProcessController) Update(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt("id")
	var process models.Process
	// 开启事务：步骤属性、流程拓扑、审批权限、条件表达式需整体原子更新
	tx, _ := facades.Orm().Query().Begin()
	err := tx.Model(&models.Process{}).Where("id=?", id).Find(&process)
	if err != nil {
		tx.Rollback()
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "流程不存在", nil)
	}
	var processRequest common.ProcessRequest
	if err := ctx.Request().Bind(&processRequest); err != nil {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "数据错误", nil)
	}

	// 若设置为结束/起始步骤（position=9），需校验：分支节点（>1 条转发）不允许设为结束步骤
	if processRequest.ProcessPosition == 9 {
		count, _ := tx.Model(&models.Flowlink{}).Where("process_id=?", id).Count()
		if count > 1 {
			return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "该节点是分支节点，不能设置为结束或起始步骤", nil)
		}
	}
	// 若设置为起始步骤（position=0），先将同流程内原有起始步骤降级为普通步骤
	if processRequest.ProcessPosition == 0 {
		_, err := tx.Model(&models.Process{}).Where("flow_id=?", process.FlowID).Where("position", 0).Update("position", 1)
		if err != nil {
			tx.Rollback()
			return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "数据错误", nil)
		}
	}
	// 更新步骤基础属性
	process.ProcessName = processRequest.ProcessName
	process.StyleColor = processRequest.StyleColor
	process.StyleHeight = processRequest.StyleHeight
	process.StyleWidth = processRequest.StyleWidth
	process.Style = fmt.Sprintf("width:%dpx;height:%dpx;line-height:30px;color:%s;left:%s;top:%s;",
		process.StyleWidth, process.StyleHeight, process.StyleColor, process.PositionLeft, process.PositionTop)
	process.Icon = processRequest.StyleIcon
	process.Position = processRequest.ProcessPosition
	process.ChildFlowID = processRequest.ChildFlowId
	process.ChildAfter = processRequest.ChildAfter
	process.ChildBackProcess = processRequest.ChildBackProcess
	process.LimitTime = processRequest.LimitTime
	process.CcEmpIDs = processRequest.CcEmpIDs
	if err := tx.Model(&models.Process{}).Where("id=?", id).Save(&process); err != nil {
		tx.Rollback()
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "数据错误", nil)
	}
	// 同步更新jsplumb json数据
	// 同步更新 jsPlumb JSON 数据：读取流程拓扑，找到当前节点并更新其属性
	var flow models.Flow
	err = tx.Model(&models.Flow{}).Where("id=?", process.FlowID).With("Template.TemplateForms").Find(&flow)
	if err != nil {
		tx.Rollback()
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "流程不存在", nil)
	}
	jsMap := common.Plumb{}
	//flow.Jsplum解析为jsMap
	// 将 flow.Jsplumb JSON 字符串解析为 jsMap 结构
	err = json.Unmarshal([]byte(flow.Jsplumb), &jsMap)
	if err != nil {
		tx.Rollback()
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "解析数据错误", nil)
	}

	//需要将jsMap读取出来，然后再写回去
	// 遍历 jsMap 节点列表，找到匹配当前 Process ID 的节点，用最新属性覆盖
	for key, val := range jsMap.List {
		if key == cast.ToString(process.ID) {
			jsMap.List[key] = common.Node{
				ID:          cast.ToInt(process.ID),
				FlowId:      process.FlowID,
				ProcessTo:   val.ProcessTo,
				ProcessName: processRequest.ProcessName,
				Icon:        processRequest.StyleIcon,
				Style: fmt.Sprintf("width:%dpx;height:%dpx;line-height:30px;color:%s;left:%s;top:%s;",
					processRequest.StyleWidth, processRequest.StyleHeight, processRequest.StyleColor, process.PositionLeft, process.PositionTop),
			}
		}
	}

	jsplumbByte, err := json.Marshal(jsMap)

	if err != nil {
		tx.Rollback()
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "解析数据错误", nil)
	}
	//更新流程图
	// 将更新后的拓扑 JSON 写回 Flow 并标记为未发布
	flow.Jsplumb = string(jsplumbByte)
	_, err = tx.Model(flow).Where("id=?", flow.ID).Update("jsplumb", flow.Jsplumb)
	_, err = tx.Model(flow).Where("id=?", flow.ID).Update("IsPublish", false)
	if err != nil {
		tx.Rollback()
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "数据错误", nil)
	}

	//更新步骤 流转条件 process_condition
	//根据ProcessCondition中的每一项分组，然后将每一组id相同的数据找出，将表达式合并为一个fmt.Sprintf("%s%s%s", condition.Field, condition.Operator, condition.Value)
	// 处理条件表达式：按 flowlink ID 分组，每组内的多个条件合并为一个 JSON 存储
	var conditionsMap map[int][]common.ProcessCondition
	if len(processRequest.ProcessCondition) > 0 {
		conditionsMap = groupConditionsById(processRequest.ProcessCondition)
	}
	//根据提交的conditionsMap更新process_var中的数据，如果有新数据，则新增，前提是只针对类型为int字段的数据，
	// 同步更新 process_var 表：当条件字段尚不存在时，自动新增一条记录
	for _, conditions := range conditionsMap {
		for _, condition := range conditions {
			if condition.Field != "" {
				exists_count, _ := facades.Orm().Query().Model(&models.ProcessVar{}).
					Where("flow_id=?", flow.ID).
					Where("process_id=?", id).
					Where("expression_field=?", condition.Field).Count()
				if exists_count == 0 {
					//新增一条
					// 新增一条 process_var 记录
					var newProcessVar models.ProcessVar
					newProcessVar.FlowID = cast.ToInt(flow.ID)
					newProcessVar.ProcessID = id
					newProcessVar.ExpressionField = condition.Field
					facades.Orm().Query().Model(&models.ProcessVar{}).Create(&newProcessVar)

				}
			}
		}
	}

	// 将每组条件序列化为 JSON 并写回对应的 flowlink 记录
	for key, conditions := range conditionsMap {
		jsonStr, _ := json.Marshal(conditions)
		// json.Marshal escapes < > as < > (HTML safety), which MySQL then
		// double-escapes to < >. Decode back to literal < > before storing
		// so the runtime condition evaluator can build valid SQL.
		// json.Marshal 会将 < > 转义为 HTML 实体（< >），MySQL 会再次转义，
		// 导致双重转义。此处将其还原为字面量 < >，确保运行时条件求值器能构建正确的 SQL。
		jsonStr = fixJSONUnicodeEscapes(jsonStr)
		tx.Model(&models.Flowlink{}).Where("id=?", key).Update(map[string]interface{}{
			"expression":     jsonStr,
			"condition_expr": jsonStr,
		})
	}

	//@改，如果当前的processRequest.AutoPerson=="0",更新当前的步骤
	// 如果 AutoPerson 为 "0"，更新当前步骤下所有 Condition 类型 flowlink 的审批人为 "0"
	if processRequest.AutoPerson == "0" {
		tx.Model(&models.Flowlink{}).Where("flow_id=?", flow.ID).Where("process_id=?", id).
			Where("type=?", "Condition").Update("auditor", processRequest.AutoPerson)
	}
	// 从 jsplumb JSON 中读取当前节点的 process_to，解析默认下一步骤
	// 从 jsPlumb 拓扑 JSON 中解析当前节点的默认下一步骤 ID
	defaultNextID := 0
	jsNode := jsMap.List[cast.ToString(process.ID)]
	if jsNode.ProcessTo != "" {
		parts := strings.Split(jsNode.ProcessTo, ",")
		if len(parts) > 0 && parts[0] != "" {
			if id, err := cast.ToIntE(parts[0]); err == nil && id > 0 {
				defaultNextID = id
			}
		}
	}

	// 多条件节点（>1 条 Condition flowlink）：不强制设置非 Condition flowlink 的默认路由
	// 每个分支独立路由，审批人设置不绑定固定的下一步骤
	// 多条件分支节点：每个分支独立路由，不强制绑定固定的下一步骤
	conditionCount, _ := tx.Model(&models.Flowlink{}).Where("flow_id=?", flow.ID).
		Where("type=?", "Condition").Where("process_id=?", id).Count()
	if conditionCount > 1 {
		defaultNextID = 0
	}

	//权限处理
	// 审批权限处理
	if processRequest.AutoPerson != "0" && processRequest.AutoPerson != "" {

		// 系统自动审批（Sys 类型）：更新或新建 Sys 类型的 flowlink
		var fk models.Flowlink
		tx.Model(&fk).Where("flow_id=?", flow.ID).Where("process_id=?", id).Where("type=?", "Sys").Find(&fk)
		if fk.ID != 0 {
			// 已有 Sys 记录：更新审批人、并发类型、审批规则和默认下一步骤
			fk.Auditor = cast.ToString(processRequest.AutoPerson)
			_, err := tx.Model(&models.Flowlink{}).Where("id=?", fk.ID).Update(map[string]interface{}{
				"auditor":          processRequest.AutoPerson,
				"concurrency_type": processRequest.ConcurrencyType,
				"approver_rule":    processRequest.ApproverRule,
				"next_process_id":  defaultNextID,
			})
			if err != nil {
				tx.Rollback()
				return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "数据错误", nil)
			}
		} else {
			// 无 Sys 记录：新建一条
			tx.Model(&models.Flowlink{}).Create(&models.Flowlink{
				FlowID:          flow.ID,
				Type:            "Sys",
				ProcessID:       cast.ToUint(id),
				Auditor:         cast.ToString(processRequest.AutoPerson),
				ConcurrencyType: processRequest.ConcurrencyType,
				ApproverRule:    processRequest.ApproverRule,
				NextProcessID:   defaultNextID,
				Sort:            100,
			})
		}
		//更新当前flowlink的Audiitor

		//删除其他权限
		// 删除非 Condition 且非 Sys 的其他审批人类型（清理由 AutoPerson 替换的旧权限）
		tx.Model(&models.Flowlink{}).Where("flow_id=?", flow.ID).Where("process_id=?", id).
			Where("type!=?", "Condition").Where("type!=?", "Sys").Delete(&models.Flowlink{})
	} else {
		// 手动指定审批人：按部门（Dept）和员工（Emp）分别处理
		//指定部门
		if len(processRequest.RangeDeptIds) > 0 {
			var fkdept models.Flowlink
			tx.Model(&fkdept).Where("flow_id=?", flow.ID).Where("process_id=?", id).Where("type=?", "Dept").Find(&fkdept)
			if fkdept.ID != 0 {
				//id组成的数组，然后转换为字符串
				// 已有部门记录：将部门 ID 数组拼接为逗号分隔的字符串
				auditor := ""
				for _, dept := range processRequest.RangeDeptIds {
					auditor += cast.ToString(dept) + ","
				}
				//取消最后一个,号
				auditor = strings.TrimSuffix(auditor, ",")
				fkdept.Auditor = auditor
				tx.Model(&models.Flowlink{}).Where("id=?", fkdept.ID).Update("auditor", fkdept.Auditor)
			} else {
				// 无部门记录：新建一条 Dept 类型的 flowlink
				auditor := ""
				for _, dept := range processRequest.RangeDeptIds {
					auditor += cast.ToString(dept) + ","
				}
				//去掉最后一个,号
				// 去掉末尾多余的逗号
				auditor = strings.TrimSuffix(auditor, ",")
				tx.Model(&models.Flowlink{}).Create(&models.Flowlink{FlowID: flow.ID, Type: "Dept", ProcessID: cast.ToUint(id), Auditor: auditor, NextProcessID: defaultNextID, Sort: 100, ConcurrencyType: processRequest.ConcurrencyType, ApproverRule: processRequest.ApproverRule})
			}
		} else {
			//删除部门权限
			// 未指定部门：删除该步骤原有的部门审批权限
			tx.Model(&models.Flowlink{}).Where("flow_id=?", flow.ID).Where("process_id=?", id).
				Where("type=?", "Dept").Delete(&models.Flowlink{})
		}
		//	指定员工
		if len(processRequest.RangeEmpIds) > 0 {
			var fkemp models.Flowlink
			tx.Model(&fkemp).Where("flow_id=?", flow.ID).Where("process_id=?", id).Where("type=?", "Emp").Find(&fkemp)
			if fkemp.ID != 0 {
				//id组成的数组，然后转换为字符串
				// 已有员工记录：将员工 ID 数组拼接为逗号分隔的字符串
				auditor := ""
				for _, emp := range processRequest.RangeEmpIds {
					auditor += cast.ToString(emp) + ","
				}
				auditor = strings.TrimSuffix(auditor, ",")
				fkemp.Auditor = auditor
				tx.Model(&models.Flowlink{}).Where("id=?", fkemp.ID).Update(map[string]interface{}{"auditor": fkemp.Auditor, "concurrency_type": processRequest.ConcurrencyType})
			} else {
				// 无员工记录：新建一条 Emp 类型的 flowlink
				auditor := ""
				for _, emp := range processRequest.RangeEmpIds {
					auditor += cast.ToString(emp) + ","
				}
				auditor = strings.TrimSuffix(auditor, ",")
				tx.Model(&models.Flowlink{}).Create(&models.Flowlink{FlowID: flow.ID, Type: "Emp", ProcessID: cast.ToUint(id), Auditor: auditor, NextProcessID: defaultNextID, Sort: 100, ConcurrencyType: processRequest.ConcurrencyType})
			}
		} else {
			//	删除
			// 未指定员工：删除该步骤原有的员工审批权限
			tx.Model(&models.Flowlink{}).Where("flow_id=?", flow.ID).Where("process_id=?", id).Where("type=?", "Emp").Delete(&models.Flowlink{})
		}
	}
	tx.Commit()
	return httpfacades.NewResult(ctx).Success("保存成功", nil)
}

// groupConditionsById 将前端提交的条件表达式数组按 flowlink ID 分组，
// 使同一 flowlink 下的多个条件合并存储为一组，便于后续条件求值。
func groupConditionsById(conditions []common.ProcessCondition) map[int][]common.ProcessCondition {
	grouped := make(map[int][]common.ProcessCondition)
	for _, condition := range conditions {
		grouped[condition.Id] = append(grouped[condition.Id], condition)
	}
	return grouped
}

// Destroy 删除指定的流程步骤（Process），同时：
//   - 删除与该步骤关联的所有 flowlink 转发记录
//   - 将其他步骤中指向该步骤的 next_process_id 置为 -1
//   - 从流程的 jsPlumb 拓扑 JSON 中移除该节点
//   - 更新 Flow 拓扑数据并提交事务
func (r *ProcessController) Destroy(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt("id")
	flow_id := ctx.Request().InputInt("flow_id")
	var flow models.Flow
	// 开启事务：删除步骤、清理关联 flowlink、更新拓扑数据需原子执行
	tx, _ := facades.Orm().Query().Begin()
	tx.Model(&flow).Where("id=?", flow_id).Find(&flow)
	// 删除以当前步骤为起点的所有 flowlink
	tx.Model(&models.Flowlink{}).Where("flow_id=?", id).Where("id=?", id).Delete(&models.Flowlink{})
	// 将其他步骤中指向当前步骤的 next_process_id 置为 -1（断开引用）
	tx.Model(&models.Flowlink{}).Where("flow_id=?", id).Where("next_process_id=?", id).Update("next_process_id", -1)
	// 删除步骤本身
	tx.Model(&models.Process{}).Where("id=?", id).Delete(&models.Process{})
	jsMap := common.Plumb{}
	//flow.Jsplum解析为jsMap
	// 将 flow.Jsplumb JSON 字符串解析为 jsMap 结构
	err := json.Unmarshal([]byte(flow.Jsplumb), &jsMap)
	if err != nil {
		tx.Rollback()
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "解析数据错误", nil)
	}

	//需要将jsMap读取出来，然后再写回去
	// 遍历 jsMap 节点列表，找到与当前 Process ID 匹配的节点并删除
	for key, _ := range jsMap.List {
		if key == cast.ToString(id) {
			//	删除
			// 从拓扑 JSON 中删除该节点
			delete(jsMap.List, key)
		}
	}

	jsplumbByte, err := json.Marshal(jsMap)

	if err != nil {
		tx.Rollback()
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "解析数据错误", nil)
	}
	//更新流程图
	// 将更新后的拓扑 JSON 写回 Flow
	flow.Jsplumb = string(jsplumbByte)
	tx.Model(&models.Flow{}).Where("id=?", flow.ID).Save(&flow)
	tx.Commit()
	return httpfacades.NewResult(ctx).Success("删除成功", nil)
}

// Attribute 获取流程步骤的完整属性信息，供前端属性面板渲染。返回内容包括：
//   1. 当前步骤的下一步操作（含默认路径和条件分支路径）
//   2. 可选步骤（beixuan）：未被当前步骤指向的其他流程步骤
//   3. 关联模板的表单字段
//   4. 已选定的员工和部门审批人
//   5. 系统自动审批设置（Sys 类型 flowlink）
//   6. 可用的子流程列表
//   7. 当前流程的所有步骤
//   8. 并发类型与审批规则
func (r *ProcessController) Attribute(ctx http.Context) http.Response {
	id := ctx.Request().QueryInt("id")
	process := models.Process{}
	tx := facades.Orm().Query()
	tx.Model(&models.Process{}).Where("id=?", id).Find(&process)

	//1- //当前步骤的下一步操作（包含默认非Condition路径 + Condition条件路径）
	// 1. 查询当前步骤的下一步操作：包含默认非 Condition 路径和 Condition 条件路径
	next_process := []models.Flowlink{}
	tx.Model(&models.Flowlink{}).Where("process_id=?", process.ID).
		Where("flow_id=?", process.FlowID).With("Process").
		With("NextProcess").Find(&next_process)
	// 过滤掉 position=0 的第一步（发起节点自身），只保留真正的"转出步骤"
	// 过滤：排除指向自身的 flowlink（自引用节点）
	filteredNext := []models.Flowlink{}
	for _, fl := range next_process {
		// 排除指向自身的 flowlink
		// 排除自引用：process_id == next_process_id 的记录
		if cast.ToInt(fl.ProcessID) == fl.NextProcessID {
			continue
		}
		filteredNext = append(filteredNext, fl)
	}
	next_process = filteredNext
	next_process_ids := []int{}
	tx.Model(&models.Flowlink{}).Where("process_id=?", process.ID).
		Where("flow_id=?", process.FlowID).With("Process").
		With("NextProcess").Pluck("next_process_id", &next_process_ids)
	beixuan_process := []models.Flowlink{}
	// beixuan: 所有非 Condition 的 flowlink 中，next_process_id 不在 next_process_ids 中的步骤
	// 备选步骤：查询流程中所有非 Condition 类型、且不在当前步骤转发目标中的其他步骤
	tx.Model(&models.Flowlink{}).Where("flow_id=?", process.FlowID).
		Where("type != ?", "Condition").
		Where("process_id not in (?)", next_process_ids).
		Where("next_process_id > 0").With("Process").With("NextProcess").Find(&beixuan_process)

	//	2-流程模板 表单字段
	// 2. 查询流程模板关联的表单字段
	flow := models.Flow{}

	fields := []models.TemplateForm{}
	tx.Model(&models.Flow{}).Where("id=?", process.FlowID).With("Template").Find(&flow)
	if flow.Template.ID != 0 {
		tfId := flow.Template.ID
		tx.Model(&models.TemplateForm{}).Where("template_id=?", tfId).Find(&fields)
	}

	//3-当前选择员工
	// 3. 查询当前步骤已选定的员工审批人
	select_emps := []models.Emp{}
	auditor_emp_flowlink := models.Flowlink{}
	tx.Model(&models.Flowlink{}).Where("process_id=?", process.ID).
		Where("type=?", "Emp").Select("auditor").Find(&auditor_emp_flowlink)
	//depts按照,拆分
	empsSlice := []string{}
	for _, emp := range strings.Split(auditor_emp_flowlink.Auditor, ",") {
		empsSlice = append(empsSlice, emp)
	}
	tx.Model(&models.Emp{}).Where("id in (?)", empsSlice).Find(&select_emps)
	//4 -flowlinks
	// 4. 查询系统自动审批（Sys 类型）的 flowlink 记录
	flowlink := models.Flowlink{}
	sys := "0"
	tx.Model(&models.Flowlink{}).Where("process_id = ?", process.ID).Where("flow_id=?", process.FlowID).
		Where("type=?", "Sys").Find(&flowlink)
	if flowlink.Auditor != "" {
		sys = flowlink.Auditor
	}

	// 5-部门
	// 5. 查询当前步骤已选定的部门审批人
	select_depts := []models.Dept{}
	auditor_dept_flowlink := models.Flowlink{}
	tx.Model(&models.Flowlink{}).Where("type=?", "Dept").Where("process_id=?", process.ID).
		Select("auditor").Find(&auditor_dept_flowlink)
	//depts按照,拆分
	deptsSlice := []string{}
	for _, dept := range strings.Split(auditor_dept_flowlink.Auditor, ",") {
		deptsSlice = append(deptsSlice, dept)
	}
	tx.Model(&models.Dept{}).Where("id in (?)", deptsSlice).Find(&select_depts)

	// 6-flow
	// 6. 查询可用的子流程列表（已发布且非当前流程）
	flows := []models.Flow{}
	tx.Model(&models.Flow{}).Where("is_publish=?", 1).Where("id!=?", process.FlowID).Find(&flows)

	// 7. 查询当前流程的所有步骤
	processes := []models.Process{}
	tx.Model(&models.Process{}).Where("flow_id=?", process.FlowID).Find(&processes)
	var can_child bool
	if process.Position != 0 { // 第一步不允许转入子流程
		// 第一步（起始步骤）不允许转入子流程，其他步骤可以
		can_child = true
	}

	// Get concurrency_type and approver_rule from Sys/Emp/Dept flowlink
	// 从 Sys/Emp/Dept 类型的 flowlink 中获取并发类型和审批规则
	concurrencyType := 0
	approverRule := ""
	if flowlink.ID != 0 {
		concurrencyType = flowlink.ConcurrencyType
		approverRule = flowlink.ApproverRule
	}
	// 优先级：Sys > Emp > Dept，若 Sys 未设置则依次向下查找
	if concurrencyType == 0 {
		empFlowlink := models.Flowlink{}
		tx.Model(&models.Flowlink{}).Where("process_id = ?", process.ID).Where("flow_id=?", process.FlowID).
			Where("type=?", "Emp").Find(&empFlowlink)
		if empFlowlink.ID != 0 {
			concurrencyType = empFlowlink.ConcurrencyType
			approverRule = empFlowlink.ApproverRule
		}
	}
	if concurrencyType == 0 {
		deptFlowlink := models.Flowlink{}
		tx.Model(&models.Flowlink{}).Where("process_id = ?", process.ID).Where("flow_id=?", process.FlowID).
			Where("type=?", "Dept").Find(&deptFlowlink)
		if deptFlowlink.ID != 0 {
			concurrencyType = deptFlowlink.ConcurrencyType
			approverRule = deptFlowlink.ApproverRule
		}
	}

	return httpfacades.NewResult(ctx).Success("", http.Json{
		"process":          process,
		"next_process":     next_process,
		"beixuan_process":  beixuan_process,
		"fields":           fields,
		"select_emps":      select_emps,
		"sys":              sys,
		"select_depts":     select_depts,
		"flows":            flows,
		"processes":        processes,
		"can_child":        can_child,
		"concurrency_type": concurrencyType,
		"approver_rule":    approverRule,
	})
}

// Condition 根据条件分支的 flowlink 记录，将表达式中使用的字段名（Field）替换为对应的
// 字段显示名称（FieldName），供前端条件预览面板展示可读的条件描述。
func (r *ProcessController) Condition(ctx http.Context) http.Response {
	flow_id := ctx.Request().InputInt("flow_id")
	process_id := ctx.Request().InputInt("process_id")
	next_process_id := ctx.Request().InputInt("next_process_id")
	//当前流程
	// 查询条件分支 flowlink 记录（Condition 类型）
	flowlink := models.Flowlink{}
	tx, _ := facades.Orm().Query().Begin()
	tx.Model(&models.Flowlink{}).Where("process_id=?", process_id).Where("next_process_id=?", next_process_id).
		Where("flow_id=?", flow_id).Where("type=?", "Condition").FindOrFail(&flowlink)
	flow := models.Flow{}
	tx.Model(&models.Flow{}).With("Template.TemplateForms").Where("id=?", flow_id).Find(&flow)
	// 将条件表达式中的字段名（Field）替换为显示名称（FieldName），
	// 例如：$day > 3 替换为 请假天数 > 3
	//$day > 3  AND
	// $sex == 女
	fieldsArr := []string{}
	for _, form := range flow.Template.TemplateForms {
		//form.field form.field_name
		// 去除表达式中的 $ 前缀
		cleanedExpression := strings.Replace(flowlink.Expression, "$", "", -1)

		if strings.Contains(cleanedExpression, form.Field) {
			// 新建一个replaceStr
			// 将字段名替换为字段显示名
			replaceStr := strings.Replace(cleanedExpression, form.Field, form.FieldName, -1)
			fieldsArr = append(fieldsArr, replaceStr)
		}
	}
	res := make(map[int]interface{})
	if len(fieldsArr) > 0 {
		res[flowlink.NextProcessID] = map[string]interface{}{
			"desc":   fieldsArr[0],
			"option": "",
		}
	} else {
		res[flowlink.NextProcessID] = map[string]interface{}{
			"desc":   []string{},
			"option": "",
		}
	}

	return httpfacades.NewResult(ctx).Success("", res)
}

// fixJSONUnicodeEscapes reverses Go's json.Marshal HTML escaping so that
// < > characters that were encoded as < > are restored to their
// literal form. This prevents MySQL from double-escaping them when the JSON
// is stored and later used for building SQL conditions at runtime.
// fixJSONUnicodeEscapes 反转 Go 的 json.Marshal HTML 转义，将 < 和 >
// 还原为字面量 < > 字符，避免 MySQL 存储时双重转义，确保运行时条件求值器能正确构建 SQL。
