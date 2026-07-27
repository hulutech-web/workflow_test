package controllers

import (
	"goravel/packages/goravel-workflow/models"
	workflow "goravel/packages/goravel-workflow/services/workflow"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	httpfacades "github.com/hulutech-web/http_result"
	"github.com/spf13/cast"
)

// ProcController 审批任务控制器，处理审批流程中的各项操作
type ProcController struct {
	workflow *workflow.Workflow
}

// NewProcController 创建审批任务控制器实例，初始化工作流引擎
func NewProcController() *ProcController {
	return &ProcController{
		workflow: workflow.NewBaseWorkflow(),
	}
}

// Index 获取指定流程实例下的所有审批任务列表
// 根据 entry_id 查询关联的审批任务（Proc），并预加载关联的流程实例（Entry）及发起人（Emp）信息
func (r *ProcController) Index(ctx http.Context) http.Response {
	entry_id := ctx.Request().InputInt("entry_id") // 从请求中获取流程实例 ID
	var procs []models.Proc
	// 查询指定 entry_id 下的所有审批任务，预加载关联数据
	facades.Orm().Query().Model(&models.Proc{}).Where("entry_id=?", entry_id).With("Entry.Emp").Find(&procs)
	return httpfacades.NewResult(ctx).Success("", procs)
}

// Pass 审批通过操作
// 当前审批人对指定的审批步骤进行通过操作，推动流程向前流转
func (r *ProcController) Pass(ctx http.Context) http.Response {
	var user models.Emp
	facades.Auth(ctx).User(&user) // 从 JWT 认证中获取当前登录用户
	process_id := ctx.Request().InputInt("process_id") // 步骤 ID
	content := ctx.Request().Input("content")           // 审批意见
	formData := ctx.Request().All()                     // 表单数据（用于条件分支判断）
	err := r.workflow.Pass(process_id, user, content, formData)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(500, "审批失败", err.Error())
	}
	return httpfacades.NewResult(ctx).Success("审批成功", nil)
}

// UnPass 驳回操作
// 支持普通驳回（驳回到上一步）和驳回到指定节点（unpassTo）两种模式
func (r *ProcController) UnPass(ctx http.Context) http.Response {
	var user models.Emp
	facades.Auth(ctx).User(&user) // 从 JWT 认证中获取当前登录用户
	withUser := models.Emp{}
	// 查询当前用户及其所属部门信息，用于驳回时判断审批人所在部门
	facades.Orm().Query().Model(&models.Emp{}).Where("id=?", user.ID).With("Dept").Find(&withUser)
	proc_id := ctx.Request().InputInt("proc_id")                         // 当前审批任务 ID
	content := ctx.Request().Input("content")                            // 驳回意见
	target_process_id := ctx.Request().InputInt("target_process_id")     // 目标步骤 ID（驳回到指定节点时使用）

	var err error
	if target_process_id > 0 {
		// 驳回到指定节点：将流程回退到指定的历史步骤
		err = r.workflow.UnPassTo(proc_id, withUser, content, target_process_id)
	} else {
		// 普通驳回：将流程驳回到上一步，标记整个流程实例为已驳回
		err = r.workflow.UnPass(proc_id, withUser, content)
	}
	if err != nil {
		return httpfacades.NewResult(ctx).Error(500, "驳回失败", err.Error())
	}
	return httpfacades.NewResult(ctx).Success("驳回成功", nil)
}

// Revoke 撤回功能 / 撤回功能：发起人撤回自己发起的流程实例
// 只有流程发起人才能撤回，且流程必须处于待审批状态（尚无人审批通过）
func (r *ProcController) Revoke(ctx http.Context) http.Response {
	var user models.Emp
	facades.Auth(ctx).User(&user) // 从 JWT 认证中获取当前登录用户
	entry_id := ctx.Request().InputInt("entry_id") // 流程实例 ID
	err := r.workflow.Revoke(uint(entry_id), user)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(500, "撤回失败", err.Error())
	}
	return httpfacades.NewResult(ctx).Success("撤回成功", nil)
}

// AddSign 加签功能 / 加签功能：在当前审批步骤中额外添加审批人
// sign_type 可选 "before"（前加签）或 "after"（后加签）
func (r *ProcController) AddSign(ctx http.Context) http.Response {
	var user models.Emp
	facades.Auth(ctx).User(&user) // 从 JWT 认证中获取当前登录用户
	entry_id := ctx.Request().InputInt("entry_id")         // 流程实例 ID
	process_id := ctx.Request().InputInt("process_id")     // 当前步骤 ID
	sign_emp_id := ctx.Request().InputInt("sign_emp_id")   // 被加签的员工 ID
	sign_type := ctx.Request().Input("sign_type")          // 加签类型: "before" 前加签 或 "after" 后加签

	err := r.workflow.AddSign(uint(entry_id), process_id, sign_emp_id, sign_type, user)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(500, "加签失败", err.Error())
	}
	return httpfacades.NewResult(ctx).Success("加签成功", nil)
}

// TransferProc 转交功能 / 转交功能：将当前审批人的审批任务转交给其他员工处理
// 原审批任务标记为已转交（status=3），创建新的待审批任务给目标员工
func (r *ProcController) TransferProc(ctx http.Context) http.Response {
	var user models.Emp
	facades.Auth(ctx).User(&user) // 从 JWT 认证中获取当前登录用户
	entry_id := ctx.Request().InputInt("entry_id")             // 流程实例 ID
	proc_id := ctx.Request().InputInt("proc_id")               // 当前审批任务 ID
	target_emp_id := ctx.Request().InputInt("target_emp_id")    // 目标接收员工 ID

	err := r.workflow.TransferProc(uint(entry_id), uint(proc_id), target_emp_id, user)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(500, "转交失败", err.Error())
	}
	return httpfacades.NewResult(ctx).Success("转交成功", nil)
}

// AddComment 评论功能 / 评论功能：在审批流程中添加评论，支持回复（通过 parent_id 指定父评论）
func (r *ProcController) AddComment(ctx http.Context) http.Response {
	var user models.Emp
	facades.Auth(ctx).User(&user) // 从 JWT 认证中获取当前登录用户
	var emp models.Emp
	// 通过 user_id 关联查询对应的员工记录
	facades.Orm().Query().Model(&models.Emp{}).Where("user_id=?", user.ID).Find(&emp)

	entry_id := ctx.Request().InputInt("entry_id")                   // 流程实例 ID
	proc_id := ctx.Request().InputInt("proc_id")                     // 审批任务 ID
	content := ctx.Request().Input("content")                        // 评论内容
	parent_id := ctx.Request().InputInt("parent_id")                 // 父评论 ID（用于回复场景）
	reply_to_emp_id := ctx.Request().InputInt("reply_to_emp_id")      // 回复目标员工 ID
	reply_to_emp_name := ctx.Request().Input("reply_to_emp_name")     // 回复目标员工姓名

	// 创建评论记录，支持嵌套回复结构
	err := r.workflow.AddComment(uint(entry_id), uint(proc_id), cast.ToInt(emp.ID), emp.Name, content,
		uint(parent_id), reply_to_emp_id, reply_to_emp_name)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(500, "评论失败", err.Error())
	}
	return httpfacades.NewResult(ctx).Success("评论成功", nil)
}

// GetComments 获取评论列表 / 获取评论列表：根据流程实例 ID 查询所有评论
// 评论按 ID 升序排列，支持树形嵌套结构（父子评论）
func (r *ProcController) GetComments(ctx http.Context) http.Response {
	entry_id := ctx.Request().RouteInt("entry_id") // 从路由参数中获取流程实例 ID
	comments, err := r.workflow.GetComments(uint(entry_id))
	if err != nil {
		return httpfacades.NewResult(ctx).Error(500, "获取评论失败", err.Error())
	}
	return httpfacades.NewResult(ctx).Success("", comments)
}

// RejectableProcesses 获取可驳回的节点列表 / 获取可驳回的节点列表：返回流程实例中可以驳回到的历史步骤
// 筛选逻辑：查找当前步骤之前已经通过（status=1 已审批）的步骤，排除待处理、已跳过、已驳回的状态
// 同时确保流程起始步骤（position=0）始终包含在可驳回列表中
func (r *ProcController) RejectableProcesses(ctx http.Context) http.Response {
	entry_id := ctx.Request().RouteInt("entry_id") // 从路由参数中获取流程实例 ID
	var entry models.Entry
	// 查询流程实例信息
	facades.Orm().Query().Model(&models.Entry{}).Where("id=?", entry_id).Find(&entry)
	if entry.ID == 0 {
		return httpfacades.NewResult(ctx).Error(500, "流程实例不存在", nil)
	}
	currentProcessID := int(entry.ProcessID) // 当前所在步骤 ID

	// Query all procs for this entry/circle, dedup in Go
	// 查询该流程实例当前轮次的所有审批任务，在 Go 代码中去重
	var allProcs []models.Proc
	facades.Orm().Query().Model(&models.Proc{}).
		Where("entry_id=?", entry_id).
		Where("circle=?", entry.Circle).
		Find(&allProcs)

	seen := make(map[int]bool)
	var pastProcessIDs []int
	for _, p := range allProcs {
		if seen[p.ProcessID] {
			continue // 已处理的步骤 ID，跳过去重
		}
		if p.ProcessID == currentProcessID {
			continue // 排除当前步骤本身
		}
		if p.Status == models.ProcStatusPending || p.Status == models.ProcStatusSkipped || p.Status == models.ProcStatusRejected {
			continue // 排除待处理、已跳过、已驳回状态的任务
		}
		seen[p.ProcessID] = true
		pastProcessIDs = append(pastProcessIDs, p.ProcessID)
	}

	// Ensure the flow's starting process (position=0) is always included
	// 确保流程的起始步骤（position=0）始终包含在可驳回列表中
	if !seen[int(currentProcessID)] {
		var startProcess models.Process
		facades.Orm().Query().Model(&models.Process{}).
			Where("flow_id=?", entry.FlowID).
			Where("position=?", 0).
			Find(&startProcess)
		if startProcess.ID > 0 && !seen[int(startProcess.ID)] && int(startProcess.ID) != currentProcessID {
			// 将起始步骤插入到列表最前面
			pastProcessIDs = append([]int{int(startProcess.ID)}, pastProcessIDs...)
		}
	}

	// 根据收集到的步骤 ID 列表查询对应的流程步骤信息
	var processes []models.Process
	facades.Orm().Query().Model(&models.Process{}).
		Where("flow_id=?", entry.FlowID).
		Where("id IN ?", pastProcessIDs).
		Where("position != ?", 9). // 排除 position=9 的特殊步骤（已完成的最后步骤标记）
		Order("position ASC").      // 按步骤位置升序排列
		Find(&processes)

	// 组装返回结果，只返回必要的字段
	var result []map[string]interface{}
	for _, p := range processes {
		result = append(result, map[string]interface{}{
			"id":           p.ID,
			"process_name": p.ProcessName,
			"position":     p.Position,
		})
	}
	return httpfacades.NewResult(ctx).Success("", result)
}
