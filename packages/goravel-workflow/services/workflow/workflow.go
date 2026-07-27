package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"reflect"

	"goravel/packages/goravel-workflow/controllers/common"
	"goravel/packages/goravel-workflow/models"
	"goravel/packages/goravel-workflow/services/workflow/official_plugins"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"github.com/spf13/cast"
)

// Workflow 并发类型常量 / Workflow concurrency type constants
const (
	// ConcTypeSequential 依次审批（默认）：按顺序逐个审批，前一人通过后下一人才开始
	ConcTypeSequential = 0 // 依次审批（默认）
	// ConcTypeConsensus 会签：所有人通过才进入下一步，任意一人驳回则整体驳回
	ConcTypeConsensus = 1 // 会签：所有人通过才进入下一步
	// ConcTypeAny 或签：一人通过即进入下一步，其余审批人自动跳过
	ConcTypeAny = 2 // 或签：一人通过即进入下一步，其余跳过
)

// Flowlink 审批人特殊值常量 / Flowlink auditor special values
const (
	// AuditorInitiator 发起人：审批人为流程发起人自己
	AuditorInitiator = -1000 // 发起人
	// AuditorDirector 部门主管：审批人为发起人所属部门的负责人
	AuditorDirector = -1001 // 部门主管
	// AuditorManager 部门经理：审批人为发起人所属部门的经理
	AuditorManager = -1002 // 部门经理
	// AuditorFormField 从表单字段读取审批人：从 EntryData 中根据 ApproverRule 指定的字段名读取审批人 ID
	AuditorFormField = -1003 // 从表单字段读取审批人
	// AuditorDynamicExpr 动态表达式计算审批人：根据 ApproverRule 映射键动态计算审批人
	AuditorDynamicExpr = -1004 // 动态表达式计算审批人
)

// Workflow 是工作流引擎的核心结构体，包含钩子注册表和并发安全的读写锁。
// Workflow is the core workflow engine struct, holding hooks and a concurrency-safe mutex.
type Workflow struct {
	hooks map[string][]reflect.Value // 钩子映射表，键为钩子名称，值为注册的回调方法列表 / Hook registry, keyed by hook name, valued by registered callback methods
	mutex sync.RWMutex               // 读写互斥锁，保证高并发场景下钩子的安全注册与调用 / Read-write mutex for safe hook registration and invocation
}

// Singleton 是 Workflow 的单例实例 / Singleton is the Workflow singleton instance
var (
	baseWorkflowInstance *Workflow // 工作流引擎全局单例 / Global workflow engine singleton
	once                 sync.Once // sync.Once 确保单例只初始化一次 / sync.Once ensures singleton is initialized exactly once
)

// NewBaseWorkflow 返回 Workflow 的单例实例。使用 sync.Once 保证全局唯一。
// NewBaseWorkflow returns the singleton Workflow instance.
func NewBaseWorkflow() *Workflow {
	once.Do(func() {
		baseWorkflowInstance = &Workflow{
			hooks: make(map[string][]reflect.Value),
		}
	})
	return baseWorkflowInstance
}

// RegisterHook 通过名称注册一个钩子函数。已注册的钩子会在对应事件发生时被调用。
// RegisterHook registers a hook function by name.
func (w *Workflow) RegisterHook(name string, method reflect.Value) {
	if w.hooks == nil {
		w.hooks = make(map[string][]reflect.Value)
	}

	// 加写锁以安全地修改钩子映射表 / Acquire write lock to safely modify the hooks map
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.hooks[name] = append(w.hooks[name], method)
}

// NotifySendOne 调用 NotifySendOneHook 钩子，用于通知流程发起人（如流程完成、驳回等场景）。
// NotifySendOne calls the NotifySendOne hook.
func (w *Workflow) NotifySendOne(id uint) error {
	if w == nil {
		return fmt.Errorf("workflow instance is nil")
	}
	facades.Log().Infof("NotifySendOne called with id=%d", id)
	w.invokeHooks("NotifySendOneHook", id)
	return nil
}

// NotifyNextAuditor 调用 NotifyNextAuditorHook 钩子，用于通知下一个审批人（如有新任务分配）。
// NotifyNextAuditor calls the NotifyNextAuditor hook.
func (w *Workflow) NotifyNextAuditor(id uint) error {
	if w == nil {
		return fmt.Errorf("workflow instance is nil")
	}
	facades.Log().Infof("NotifyNextAuditor called with id=%d", id)
	w.invokeHooks("NotifyNextAuditorHook", id)
	return nil
}

// invokeHooks 调用指定名称的所有已注册钩子函数。
// 使用读写锁的读锁进行并发安全的钩子查找，允许多个 goroutine 同时读取钩子列表。
// invokeHooks calls all registered hooks for the given name.
// Uses RWMutex for thread-safe concurrent reads.
func (w *Workflow) invokeHooks(hookName string, id uint) {
	// 加读锁查找钩子列表 / Acquire read lock to look up hooks
	w.mutex.RLock()
	hooks, ok := w.hooks[hookName]
	w.mutex.RUnlock()

	if !ok {
		facades.Log().Info("Hook not found: " + hookName)
		return
	}

	// 遍历所有已注册的钩子函数并逐个调用 / Iterate all registered hooks and invoke each one
	for _, hook := range hooks {
		methodType := hook.Type()
		// 验证方法签名：必须接受一个 uint 类型参数 / Validate method signature: must accept one uint parameter
		if methodType.NumIn() == 1 && methodType.In(0).Kind() == reflect.Uint {
			facades.Log().Info("Calling hook: " + hookName)
			hook.Call([]reflect.Value{reflect.ValueOf(id)})
			facades.Log().Info("Hook completed: " + hookName)
		} else {
			facades.Log().Info("Method signature mismatch or invalid hook for " + hookName)
		}
	}
}

// skipRemainingConcurrentProcs 在或签模式下，将同一流程同一步骤同一圈次的所有剩余待处理审批任务标记为已跳过。
// skipRemainingConcurrentProcs marks all remaining pending procs as skipped in an any-sign step
func (w *Workflow) skipRemainingConcurrentProcs(query orm.Query, entryID uint, processID int, circle int) error {
	var pendingProcs []models.Proc
	// 查询所有待处理的审批任务 / Query all pending approval tasks
	if err := query.Model(&models.Proc{}).
		Where("entry_id=?", entryID).
		Where("process_id=?", processID).
		Where("circle=?", circle).
		Where("status=?", models.ProcStatusPending).
		Find(&pendingProcs); err != nil {
		return err
	}
	// 批量标记为已跳过 / Batch mark as skipped
	for i := range pendingProcs {
		pendingProcs[i].Status = models.ProcStatusSkipped
		if err := query.Model(&models.Proc{}).Where("id=?", pendingProcs[i].ID).Save(&pendingProcs[i]); err != nil {
			return err
		}
	}
	return nil
}

// checkConsensusComplete 检查会签模式下所有审批人是否已完成（通过或驳回）。
// 返回三个值：是否全部完成、是否存在驳回、错误信息。
// checkConsensusComplete checks if all consensus approvers have finished (approved or rejected)
func (w *Workflow) checkConsensusComplete(query orm.Query, entryID uint, processID int, circle int, currentProcID uint) (allDone bool, hasRejection bool, err error) {
	// 统计总审批任务数 / Count total procs
	totalProcs, err := query.Model(&models.Proc{}).Where("entry_id=?", entryID).Where("process_id=?", processID).Where("circle=?", circle).Count()
	if err != nil {
		return false, false, err
	}

	// 统计已通过的审批任务数 / Count approved procs
	approvedProcs, err := query.Model(&models.Proc{}).Where("entry_id=?", entryID).Where("process_id=?", processID).Where("circle=?", circle).Where("status=?", models.ProcStatusApproved).Count()
	if err != nil {
		return false, false, err
	}

	// 统计已驳回的审批任务数 / Count rejected procs
	rejectedProcs, err := query.Model(&models.Proc{}).Where("entry_id=?", entryID).Where("process_id=?", processID).Where("circle=?", circle).Where("status=?", models.ProcStatusRejected).Count()
	if err != nil {
		return false, false, err
	}

	// 如果传入了当前审批任务 ID 且仍为待处理状态，将其计入已通过（用于预判当前操作后的状态）
	// If current proc ID is provided and it's still pending, count it as done for this check
	if currentProcID > 0 {
		currentProc := models.Proc{}
		if err := query.Model(&models.Proc{}).Where("id=?", currentProcID).First(&currentProc); err == nil && currentProc.Status == models.ProcStatusPending {
			approvedProcs++
		}
	}

	hasRejection = rejectedProcs > 0
	allDone = (approvedProcs + rejectedProcs) >= totalProcs
	return allDone, hasRejection, nil
}

// createProcsForProcess 为指定流程步骤的所有审批人创建审批任务（Proc）记录。
// 每创建一个审批任务后，通过 NotifyNextAuditor 钩子通知对应审批人。
// createProcsForProcess creates Proc records for all auditors at a given process step
func (w *Workflow) createProcsForProcess(query orm.Query, entry *models.Entry, processID int, processName string, auditorIDs []int) error {
	if len(auditorIDs) == 0 {
		return errors.New("未找到审批人")
	}

	now := carbon.NewDateTime(carbon.Now())
	// 遍历每个审批人，创建对应的 Proc 记录 / Iterate each auditor and create their Proc record
	for _, empID := range auditorIDs {
		var emp models.Emp
		// 加载员工及其部门信息 / Load employee with department info
		if err := query.Model(&models.Emp{}).Where("id=?", empID).With("Dept").First(&emp); err != nil {
			continue
		}
		proc := models.Proc{
			EntryID:     entry.ID,
			FlowID:      cast.ToInt(entry.FlowID),
			ProcessID:   processID,
			ProcessName: processName,
			EmpID:       cast.ToInt(emp.ID),
			EmpName:     emp.Name,
			DeptName:    emp.Dept.DeptName,
			Circle:      entry.Circle,
			Status:      models.ProcStatusPending, // 初始状态为待处理 / Initial status is pending
			IsRead:      0,                        // 未读 / Unread
			Concurrence: now,                      // 并发时间戳，用于区分同一圈次的审批批次 / Concurrency timestamp for batch identification
		}
		if err := query.Model(&models.Proc{}).Create(&proc); err != nil {
			return err
		}
		// 通知下一审批人 / Notify the next auditor
		w.NotifyNextAuditor(emp.ID)
	}
	return nil
}

// SetFirstProcessAuditor 初始化流程的第一步审批任务。
// 处理发起人步骤自动通过、条件分支匹配、审批人计算等逻辑。
// SetFirstProcessAuditor initializes approval tasks for the first process step
func (w *Workflow) SetFirstProcessAuditor(entry models.Entry, flowlink models.Flowlink) error {
	return facades.Orm().Transaction(func(tx orm.Query) error {
		var myFlowlink models.Flowlink
		var auditor_ids []int

		// 查询当前步骤的非条件类型流转关系 / Query non-Condition flowlink for this process step
		err := tx.Model(&models.Flowlink{}).Where("type != ?", "Condition").
			Where("process_id=?", flowlink.ProcessID).First(&myFlowlink)
		if err != nil {
			return err
		}

		var process_id int
		var process_name string
		// 自动通过发起人步骤（position=0）仅适用于根流程，不适用于子流程。
		// 子流程的第一步有真实的审批人，需要创建待处理任务。
		// Auto-approve the initiator step (position=0) only for the root entry, NOT child flows.
		// Child flows' first step has real approvers that should get pending procs.
		isFirstStep := flowlink.Process.Position == 0 && entry.Pid == 0
		if myFlowlink.ID == 0 || isFirstStep {
			// 第一步没有指定审批人或为发起人步骤，自动通过并进入下一步
			// First step has no designated approver or is the initiator step, auto-approve and move to next
			proc := models.Proc{
				EntryID:     entry.ID,
				FlowID:      cast.ToInt(entry.FlowID),
				ProcessID:   cast.ToInt(flowlink.ProcessID),
				ProcessName: flowlink.Process.ProcessName,
				EmpID:       cast.ToInt(entry.EmpID),
				EmpName:     entry.Emp.Name,
				DeptName:    entry.Emp.Dept.DeptName,
				AuditorID:   cast.ToInt(entry.EmpID),
				AuditorName: entry.Emp.Name,
				AuditorDept: entry.Emp.Dept.DeptName,
				Status:      models.ProcStatusConsensus, // 会签通过状态（自动通过） / Consensus-approved (auto-approved)
				Circle:      entry.Circle,
				Concurrence: carbon.NewDateTime(carbon.Now()),
			}
			if err := tx.Model(&models.Proc{}).Create(&proc); err != nil {
				return err
			}

			// 创建时评估条件分支以路由到正确步骤
			// Evaluate condition branches at creation time to route to correct step
			var nextProcID int
			var proc_name string
			var evalErr error
			auditor_ids, nextProcID, proc_name, evalErr = w.evalConditionsAtCreate(tx, &entry, cast.ToInt(flowlink.ProcessID))
			if evalErr != nil {
				return evalErr
			}
			// nextProcID == -1: 流程已直接到达结束节点，已完成并归档
			// nextProcID == -1: entry reached end node directly, already completed + archived
			if nextProcID == -1 {
				return nil
			}
			process_id = nextProcID
			process_name = proc_name
			entry.ProcessID = cast.ToUint(nextProcID)
		} else {
			// 非发起人步骤：根据步骤配置计算审批人
			auditor_ids = w.GetProcessAuditorIds(entry, cast.ToInt(flowlink.ProcessID))
			process_id = cast.ToInt(flowlink.ProcessID)
			process_name = flowlink.Process.ProcessName
			entry.ProcessID = cast.ToUint(flowlink.ProcessID)
		}

		// 使用外层事务查询审批人，避免嵌套事务 / Query auditors using the outer transaction instead of creating a nested one
		var auditors_emps []models.Emp
		if err := tx.Model(&models.Emp{}).Where("id IN (?)", auditor_ids).With("Dept").Find(&auditors_emps); err != nil {
			return err
		}
		if len(auditors_emps) < 1 {
			return errors.New("下一步骤未找到审批人")
		}

		// 为每个审批人创建 Proc 记录 / Create Proc for each auditor
		for _, emp := range auditors_emps {
			proc2 := models.Proc{
				EntryID:     entry.ID,
				FlowID:      cast.ToInt(entry.FlowID),
				ProcessID:   process_id,
				ProcessName: process_name,
				EmpID:       cast.ToInt(emp.ID),
				EmpName:     emp.Name,
				DeptName:    emp.Dept.DeptName,
				Circle:      entry.Circle,
				Status:      models.ProcStatusPending,
				IsRead:      0,
				Concurrence: carbon.NewDateTime(carbon.Now()),
			}
			if err := tx.Model(&models.Proc{}).Create(&proc2); err != nil {
				return err
			}
			w.NotifyNextAuditor(emp.ID)
		}

		// 更新当前步骤 / Update current process step on entry
		_, err = tx.Model(models.Entry{}).Where("id=?", entry.ID).Update("process_id", entry.ProcessID)
		return err
	})
}

// evalConditionsAtCreate 在流程创建时（position=0 自动通过）评估条件分支。
// 返回值：(审批人ID列表, 下一步骤ID, 步骤名称, 错误)。
// 当没有配置条件分支时，回退到普通（非 Condition 类型）的流转关系。
// evalConditionsAtCreate evaluates condition flowlinks at entry creation time (position=0 auto-approve).
// Returns (auditor_ids, next_process_id, process_name, error).
// Falls back to normal (non-Condition) flowlink when no conditions are configured.
func (w *Workflow) evalConditionsAtCreate(tx orm.Query, entry *models.Entry, procProcessID int) ([]int, int, string, error) {
	flowlinks := []models.Flowlink{}
	tx.Model(&models.Flowlink{}).Where("process_id=?", procProcessID).Where("type=?", "Condition").Order("sort ASC").Find(&flowlinks)

	// 没有条件流转关系：回退到普通流转关系（简单线性流转）
	// No condition flowlinks — fall back to normal flowlink (simple linear flow)
	if len(flowlinks) == 0 {
		var defaultLink models.Flowlink
		if err := tx.Model(&models.Flowlink{}).Where("process_id=?", procProcessID).Where("type!=?", "Condition").
			With("NextProcess").First(&defaultLink); err != nil {
			return nil, 0, "", errors.New("未找到下一步骤的流程连线")
		}
		// 如果下一步是结束节点（position=9 或 NextProcessID=-1），
		// 直接完成流程 — 结束节点不配置审批人。
		// If the next step is the end node (position=9 or NextProcessID=-1),
		// complete the entry directly — end nodes have no approvers configured.
		if defaultLink.NextProcessID == -1 || defaultLink.NextProcess.Position == 9 {
			entry.Status = models.EntryStatusCompleted
			entry.ProcessID = cast.ToUint(defaultLink.NextProcessID)
			if err := tx.Model(&models.Entry{}).Where("id=?", entry.ID).Save(entry); err != nil {
				return nil, 0, "", err
			}
			w.archiveEntry(entry.ID, models.ProcStatusConsensus)
			return nil, -1, "", nil // -1 信号：流程已完成 / -1 signals: entry completed
		}
		nextProcID := cast.ToInt(defaultLink.NextProcessID)
		auditor_ids := w.GetProcessAuditorIds(*entry, nextProcID)
		if len(auditor_ids) == 0 {
			return nil, 0, "", errors.New("下一步骤未找到审批人")
		}
		return auditor_ids, nextProcID, defaultLink.NextProcess.ProcessName, nil
	}

	// 查询条件变量配置 / Query process variable config
	pvar := models.ProcessVar{}
	if err := tx.Model(&models.ProcessVar{}).Where("process_id=?", procProcessID).First(&pvar); err != nil {
		return nil, 0, "", err
	}

	var matchedFlowlink models.Flowlink
	field := pvar.ExpressionField // 条件字段名 / Condition field name

	// 支持的操作符白名单 / Supported operator whitelist
	validOperators := map[string]bool{
		"=": true, "!=": true, ">": true, "<": true, ">=": true, "<=": true,
		"like": true, "in": true, "not in": true, "between": true,
	}

	// 遍历条件分支，找到第一个满足条件的流转关系
	// Iterate condition branches, find the first matching flowlink
	for _, m := range flowlinks {
		if m.Expression == "" {
			return nil, 0, "", errors.New("未设置流转条件，无法流转")
		}
		// Expression="1" 表示无条件匹配 / Expression="1" means unconditional match
		if m.Expression == "1" {
			matchedFlowlink = m
			break
		}

		// 解析 JSON 条件表达式 / Parse JSON condition expression
		processConditions := []common.ProcessCondition{}
		if err := json.Unmarshal([]byte(UnescapeExpressionJSON(m.Expression)), &processConditions); err != nil {
			continue
		}
		if len(processConditions) == 0 {
			continue
		}

		// 验证所有条件引用同一个字段 / Validate all conditions reference the same field
		for _, cond := range processConditions {
			if cond.Field != field {
				return nil, 0, "", errors.New("没有该条件字段，请检查")
			}
			if !validOperators[strings.ToLower(cond.Operator)] {
				return nil, 0, "", errors.New("不支持的操作符")
			}
			// 转义单引号防止 SQL 注入 / Escape single quotes to prevent SQL injection
			cond.Value = strings.ReplaceAll(cond.Value, "'", "\\'")
		}

		// 构建参数化 SQL / Build parameterized SQL
		var conditionSqlParts []string
		for _, cond := range processConditions {
			extraPart := cond.Extra
			// between 操作符：使用 extra_value 作为上限 / BETWEEN operator: use ExtraValue as upper bound
			if strings.ToLower(cond.Operator) == "between" && cond.ExtraValue != "" {
				extraPart = fmt.Sprintf(" AND `field_value` >= '%s' AND `field_value` <= '%s'", cond.Value, cond.ExtraValue)
			} else {
				condExtra := strings.ReplaceAll(cond.Extra, "'", "\\'")
				extraPart = fmt.Sprintf(" `field_value` %s '%s' %s", cond.Operator, cond.Value, condExtra)
			}
			conditionSqlParts = append(conditionSqlParts, extraPart)
		}
		combined := strings.Join(conditionSqlParts, " ")

		// 正则验证字段名，防止 SQL 注入 / Regex-validate field name to prevent SQL injection
		if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(field) {
			return nil, 0, "", errors.New("无效的字段名")
		}

		// 通过原生 SQL 查询 EntryData 评估条件
		// 使用 CAST 将 field_value 转为数值类型进行数值比较
		// Evaluate condition by querying EntryData directly using sql.DB
		// Use CAST for numeric comparisons since field_value is stored as string
		numericCombined := strings.ReplaceAll(combined, "`field_value`", "CAST(`field_value` AS SIGNED)")
		sql := fmt.Sprintf("SELECT count(*) FROM entrydatas WHERE entry_id=%d AND flow_id=%d AND (%s) AND (`field_name`='%s')",
			entry.ID, entry.FlowID, numericCombined, field)
		dbHandle, err := tx.DB()
		if err != nil {
			continue
		}
		var count int
		if err := dbHandle.QueryRow(sql).Scan(&count); err != nil {
			continue
		}
		if count > 0 {
			matchedFlowlink = m
			break
		}
	}

	if matchedFlowlink.ID == 0 {
		return nil, 0, "", errors.New("未找到符合条件的流转条件，无法流转")
	}

	// 预加载匹配的流转关系的下一步信息 / Preload NextProcess for the matched flowlink
	var withFlowlink models.Flowlink
	tx.Model(&models.Flowlink{}).With("NextProcess").Where("id=?", matchedFlowlink.ID).First(&withFlowlink)
	matchedFlowlink = withFlowlink

	nextProcID := cast.ToInt(matchedFlowlink.NextProcessID)
	auditor_ids := w.GetProcessAuditorIds(*entry, nextProcID)
	if len(auditor_ids) == 0 {
		return nil, 0, "", errors.New("下一步骤未找到审批人")
	}

	return auditor_ids, nextProcID, matchedFlowlink.NextProcess.ProcessName, nil
}

// GetProcessAuditorIds 计算指定流程步骤的审批人 ID 列表。
// 按照优先级查询：Sys 类型（系统自动指定）> Emp 类型（指定员工）> Dept 类型（指定部门）。
// 最终返回去重后的审批人 ID 数组。
// GetProcessAuditorIds calculates approver IDs for a given process step
func (w *Workflow) GetProcessAuditorIds(entry models.Entry, next_process_id int) []int {
	var auditor_ids []int
	query := facades.Orm().Query()

	// 优先检查 Sys 类型（特殊审批规则：发起人/主管/经理等）
	// Check Sys type first (special approver rules like initiator/director/manager)
	var sysFlowlink models.Flowlink
	query.Model(&models.Flowlink{}).Where("type = ?", "Sys").Where("process_id=?", next_process_id).First(&sysFlowlink)

	if sysFlowlink.ID > 0 {
		switch sysFlowlink.Auditor {
		case "-1000":
			// 发起人自己 / Initiator themselves
			auditor_ids = append(auditor_ids, cast.ToInt(entry.EmpID))
		case "-1001":
			// 部门主管 / Department director
			if entry.Emp.Dept.ID > 0 {
				auditor_ids = append(auditor_ids, cast.ToInt(entry.Emp.Dept.DirectorID))
			}
		case "-1002":
			// 部门经理 / Department manager
			if entry.Emp.Dept.ID > 0 {
				auditor_ids = append(auditor_ids, cast.ToInt(entry.Emp.Dept.ManagerID))
			}
		case "-1003":
			// 从表单字段读取审批人 ID：ApproverRule 存储字段名
			// Read approver ID from form field specified in approver_rule
			if sysFlowlink.ApproverRule != "" {
				var fieldValue string
				query.Model(&models.EntryData{}).Select("field_value").
					Where("entry_id=?", entry.ID).
					Where("field_name=?", sysFlowlink.ApproverRule).
					Pluck("field_value", &fieldValue)
				if fieldValue != "" {
					if id := cast.ToInt(fieldValue); id > 0 {
						auditor_ids = append(auditor_ids, id)
					}
				}
			}
		case "-1004":
			// 动态表达式：ApproverRule 作为映射键，支持 director/manager/数字ID
			// Dynamic expression: use approver_rule as a mapping key
			if entry.Emp.Dept.ID > 0 {
				switch sysFlowlink.ApproverRule {
				case "director":
					auditor_ids = append(auditor_ids, cast.ToInt(entry.Emp.Dept.DirectorID))
				case "manager":
					auditor_ids = append(auditor_ids, cast.ToInt(entry.Emp.Dept.ManagerID))
				default:
					if id := cast.ToInt(sysFlowlink.ApproverRule); id > 0 {
						auditor_ids = append(auditor_ids, id)
					}
				}
			}
		default:
			// 直接指定审批人 ID / Direct auditor ID
			if id := cast.ToInt(sysFlowlink.Auditor); id > 0 {
				auditor_ids = append(auditor_ids, id)
			}
		}
	} else {
		// 非 Sys 类型：基于 Emp 或 Dept 的审批人分配
		// Non-Sys type: Emp or Dept based assignment

		// Emp 类型：逗号分隔的员工 ID 列表 / Emp type: comma-separated employee IDs
		var empFlowlink models.Flowlink
		query.Model(&models.Flowlink{}).Where("type = ?", "Emp").Where("process_id=?", next_process_id).First(&empFlowlink)
		if empFlowlink.ID > 0 && empFlowlink.Auditor != "" {
			for _, idStr := range strings.Split(empFlowlink.Auditor, ",") {
				if id := cast.ToInt(strings.TrimSpace(idStr)); id > 0 {
					auditor_ids = append(auditor_ids, id)
				}
			}
		}

		// Dept 类型：逗号分隔的部门 ID 列表，取各部门的负责人
		// Dept type: comma-separated department IDs, fetch each department's director
		var deptFlowlink models.Flowlink
		query.Model(&models.Flowlink{}).Where("type = ?", "Dept").Where("process_id=?", next_process_id).First(&deptFlowlink)
		if deptFlowlink.ID > 0 && deptFlowlink.Auditor != "" {
			var deptIDs []int
			for _, idStr := range strings.Split(deptFlowlink.Auditor, ",") {
				if id := cast.ToInt(strings.TrimSpace(idStr)); id > 0 {
					deptIDs = append(deptIDs, id)
				}
			}
			if len(deptIDs) > 0 {
				var directorIDs []int
				query.Model(&models.Dept{}).Select("director_id").Where("id IN (?)", deptIDs).Pluck("director_id", &directorIDs)
				auditor_ids = append(auditor_ids, directorIDs...)
			}
		}
	}

	// 去重并返回 / Deduplicate and return
	return uniqueSlice(auditor_ids)
}

// uniqueSlice 对 int 切片去重，同时保持原始顺序。
// uniqueSlice removes duplicates from an int slice while preserving order
func uniqueSlice(slice []int) []int {
	seen := make(map[int]bool)
	result := []int{}
	for _, value := range slice {
		if _, ok := seen[value]; !ok {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}

// Transfer 是工作流引擎的核心流转方法，处理审批通过后的路由逻辑。
// 包含会签/或签/依次审批、条件分支匹配、子流程创建、父子流程联动等完整逻辑。
// Transfer is the core workflow engine that handles approval routing
func (w *Workflow) Transfer(process_id int, user models.Emp, content string, formData map[string]any) error {
	query := facades.Orm().Query()

	// 将用户解析为员工 / Resolve user to emp
	var emp models.Emp
	if err := query.Model(&models.Emp{}).With("Dept").Where("user_id=?", user.ID).First(&emp); err != nil {
		return err
	}

	// 查找当前待处理的审批任务 / Find the current proc
	var proc models.Proc
	query.Model(&models.Proc{}).
		With("Entry.Emp.Dept").
		With("Entry.ParentEntry").
		Where("process_id=?", process_id).
		Where("emp_id=?", emp.ID).
		Where("status=?", models.ProcStatusPending).
		First(&proc)

	if proc.ID == 0 {
		return errors.New("未绑定员工，请设置员工绑定")
	}

	// 如果当前步骤是发起人节点（position=0），将"通过"操作视为重发：
	// 标记发起人审批任务为完成，并像新流程一样推进到下一步。
	// Position==0 仅在根流程中表示"发起人节点"。子流程的第一步有真实审批人，
	// 应使用正常的 Transfer，而非 transferFromInitiator。
	// If current step is the initiator node (position=0), treat "pass" as resend:
	// mark this initiator proc done and advance to the next step like a fresh entry
	var currentProcess models.Process
	query.Model(&models.Process{}).Where("id=?", proc.ProcessID).First(&currentProcess)
	// Position==0 only means "initiator node" for root entries. Child flows' first step
	// has real approvers and uses normal Transfer, not the resend-style transferFromInitiator.
	if currentProcess.Position == 0 && proc.Entry.Pid == 0 {
		return w.transferFromInitiator(query, &proc, content, emp, formData)
	}

	// 检查当前步骤的并发模式 / Check if this process step has concurrency mode
	var currentFlowlink models.Flowlink
	query.Model(&models.Flowlink{}).Where("process_id=?", proc.ProcessID).Where("type != ?", "Condition").First(&currentFlowlink)
	concurrencyType := currentFlowlink.ConcurrencyType

	// 会签模式：检查是否所有审批人都已完成
	// For consensus (会签): check if all approvers are done
	if concurrencyType == ConcTypeConsensus {
		allDone, hasRejection, err := w.checkConsensusComplete(query, proc.EntryID, proc.ProcessID, proc.Entry.Circle, proc.ID)
		if err != nil {
			return err
		}
		if allDone {
			if hasRejection {
				// 有人驳回 → 整个会签步骤被驳回 / One person rejected → reject the whole consensus step
				proc.Status = models.ProcStatusRejected
				proc.Content = content
				proc.AuditorID = cast.ToInt(emp.ID)
				proc.AuditorName = emp.Name
				query.Model(&models.Proc{}).Where("id=?", proc.ID).Save(&proc)
				return w.handleRejectEntry(query, &proc)
			}
			// 全部通过 → 继续后续流转（fall through）
			// All approved → proceed to next step (fall through)
		} else {
			// 还有其他审批人未完成 → 仅标记当前为已通过，等待其他人
			// Still waiting for other approvers — just mark this one approved
			proc.Status = models.ProcStatusApproved
			proc.Content = content
			proc.AuditorID = cast.ToInt(emp.ID)
			proc.AuditorName = emp.Name
			query.Model(&models.Proc{}).Where("id=?", proc.ID).Save(&proc)
			return nil
		}
	}

	// 或签模式：第一个审批人通过后，其余所有人的任务自动跳过
	// For any-sign (或签): first approver wins, skip others
	if concurrencyType == ConcTypeAny {
		if err := w.skipRemainingConcurrentProcs(query, proc.EntryID, proc.ProcessID, proc.Entry.Circle); err != nil {
			return err
		}
	}

	// --- 正常流转逻辑（依次审批或会签/或签决策后） ---
	// --- Normal transfer logic (sequential or after consensus/any-sign decision) ---

	// 检查条件分支 / Check for conditional branches
	fkcount, err := query.Model(&models.Flowlink{}).Where("process_id=?", proc.ProcessID).Where("type=?", "Condition").Count()
	if err != nil {
		return err
	}

	// 存在多个条件分支 → 条件路由 / Multiple condition branches → conditional routing
	if fkcount > 1 {
		return w.transferWithConditions(query, &proc, content, emp)
	}

	// 没有条件分支 → 查找下一步流转关系 / No conditions — find the next flowlink
	var fklink models.Flowlink
	query.Model(&models.Flowlink{}).With("Process").With("NextProcess").
		Where("process_id=?", proc.ProcessID).Where("type != ?", "Condition").First(&fklink)

	// 检查当前步骤和下一步骤的子流程标记
	// （子流程入口可能配置在当前步骤或下一步骤）
	// Check both current and next process for child workflow flag
	// (子流程入口可能配在当前步骤或下一步骤)
	childFlowID := fklink.Process.ChildFlowID
	if childFlowID == 0 {
		childFlowID = fklink.NextProcess.ChildFlowID
	}
	if childFlowID > 0 {
		return w.handleChildWorkflow(query, &proc, &fklink, content, emp)
	}

	// NextProcessID == -1 表示最后一步 / NextProcessID == -1 means last step
	if fklink.NextProcessID == -1 {
		return w.handleLastStep(query, &proc, &fklink, content, emp)
	}

	// 正常下一步 / Normal next step
	return w.handleNextStep(query, &proc, &fklink, content, emp)
}

// transferFromInitiator 处理发起人（position=0）在被驳回后点击"通过"的场景。
// 等效于重发，从发起人步骤推进到下一步。
// transferFromInitiator handles the case where the initiator (position=0) clicks "pass"
// after being rejected back — equivalent to resend, advancing to the next step.
func (w *Workflow) transferFromInitiator(query orm.Query, proc *models.Proc, content string, emp models.Emp, formData map[string]any) error {
	// 查找发起人步骤的条件类型流转关系用于路由 / Find the Condition flowlink from the initiator step for routing
	var fklink models.Flowlink
	query.Model(&models.Flowlink{}).With("NextProcess").
		Where("process_id=?", proc.ProcessID).
		Where("type=?", "Condition").
		Order("sort ASC").
		First(&fklink)
	if fklink.ID == 0 {
		return errors.New("发起人节点未配置流转关系")
	}

	// 用表单数据更新 EntryData / Update entrydatas from form data
	for key, val := range formData {
		// 跳过系统字段 / Skip system fields
		if key == "flow_id" || key == "id" || key == "entry_id" || key == "process_id" || key == "content" || key == "proc_id" {
			continue
		}
		fieldValue := cast.ToString(val)
		var existing models.EntryData
		query.Model(&models.EntryData{}).Where("entry_id=? AND field_name=?", proc.EntryID, key).First(&existing)
		if existing.ID > 0 {
			query.Model(&models.EntryData{}).Where("id=?", existing.ID).Update("field_value", fieldValue)
		}
	}

	// 将发起人审批任务标记为已通过 / Mark the initiator proc as approved with content
	proc.Status = models.ProcStatusApproved
	proc.Content = content
	proc.AuditorID = cast.ToInt(emp.ID)
	proc.AuditorName = emp.Name
	query.Model(&models.Proc{}).Where("id=?", proc.ID).Save(proc)

	// 推进到下一步，跳过已被驳回时跳过的步骤（rejectToNode 已标记为 skipped）
	// Advance to the next step, skipping already-passed steps (all should be skipped from rejectToNode)
	return w.handleNextStep(query, proc, &fklink, content, emp)
}

// transferWithConditions 处理条件分支路由。
// 解析条件表达式，通过白名单校验和参数化 SQL 查询匹配条件，找到目标流转关系。
// transferWithConditions handles conditional branch routing
func (w *Workflow) transferWithConditions(query orm.Query, proc *models.Proc, content string, emp models.Emp) error {
	// 查询条件变量配置 / Query process variable config
	pvar := models.ProcessVar{}
	if err := query.Model(&models.ProcessVar{}).Where("process_id=?", proc.ProcessID).First(&pvar); err != nil {
		return err
	}

	// 查询所有条件类型流转关系 / Query all condition-type flowlinks
	flowlinks := []models.Flowlink{}
	query.Model(&models.Flowlink{}).With("NextProcess").Where("process_id=?", proc.ProcessID).Where("type=?", "Condition").Order("sort ASC").Find(&flowlinks)

	var matchedFlowlink models.Flowlink
	field := pvar.ExpressionField

	// 遍历条件分支匹配 / Iterate conditions to find match
	for _, m := range flowlinks {
		if m.Expression == "" {
			return errors.New("未设置流转条件，无法流转")
		}
		// Expression="1" 表示无条件匹配 / Expression="1" means unconditional match
		if m.Expression == "1" {
			matchedFlowlink = m
			break
		}

		// 解析 JSON 条件 / Parse JSON conditions
		processConditions := []common.ProcessCondition{}
		if err := json.Unmarshal([]byte(UnescapeExpressionJSON(m.Expression)), &processConditions); err != nil {
			continue
		}
		if len(processConditions) == 0 {
			continue
		}

		// 安全校验：操作符白名单 / Security check: operator whitelist
		validOperators := map[string]bool{
			"=": true, "!=": true, ">": true, "<": true, ">=": true, "<=": true,
			"like": true, "in": true, "not in": true, "between": true,
		}
		for _, cond := range processConditions {
			if cond.Field != field {
				return errors.New("没有该条件字段，请检查")
			}
			if !validOperators[strings.ToLower(cond.Operator)] {
				return errors.New("不支持的操作符")
			}
			// 转义单引号防注入 / Escape single quotes against injection
			cond.Value = strings.ReplaceAll(cond.Value, "'", "\\'")
		}

		// 构造 SQL 条件表达式 / Build SQL condition expression
		conditionSql := ""
		for _, cond := range processConditions {
			extraPart := cond.Extra
			// between 操作符: 使用 extra_value 作为上限 / BETWEEN operator: use ExtraValue as upper bound
			if strings.ToLower(cond.Operator) == "between" && cond.ExtraValue != "" {
				extraPart = fmt.Sprintf(" AND `field_value` >= '%s' AND `field_value` <= '%s'", cond.Value, cond.ExtraValue)
				conditionSql += extraPart
			} else {
				condExtra := strings.ReplaceAll(cond.Extra, "'", "\\'")
				conditionSql += fmt.Sprintf(" `field_value` %s '%s' %s", cond.Operator, cond.Value, condExtra)
			}
		}

		// 正则验证字段名 / Regex-validate field name
		if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(field) {
			return errors.New("无效的字段名")
		}
		escapedField := strings.ReplaceAll(field, "'", "\\'")
		// 使用 CAST 进行数值比较 / Use CAST for numeric comparison
		numericConditionSql := strings.ReplaceAll(conditionSql, "`field_value`", "CAST(`field_value` AS SIGNED)")
		conditionSql = fmt.Sprintf("SELECT count(*) FROM entrydatas WHERE entry_id=%d AND flow_id=%d AND (%s) AND (`field_name`='%s')", proc.EntryID, proc.FlowID, numericConditionSql, escapedField)

		// 执行 SQL 评估条件 / Execute SQL to evaluate condition
		dbHandle, err := query.DB()
		if err != nil {
			return errors.New("条件语法错误，请检查")
		}
		var resultCount int
		if err := dbHandle.QueryRow(conditionSql).Scan(&resultCount); err != nil {
			return errors.New("条件语法错误，请检查")
		}
		if resultCount > 0 {
			matchedFlowlink = m
			break
		}
	}

	if matchedFlowlink.ID == 0 {
		// 构建详细错误信息，帮助排查条件匹配失败的原因 / Build a detailed error message
		var entryData models.EntryData
		query.Model(&models.EntryData{}).
			Where("entry_id=? AND flow_id=? AND field_name=?", proc.EntryID, proc.FlowID, field).
			First(&entryData)
		fieldValue := ""
		if entryData.ID > 0 {
			fieldValue = entryData.FieldValue
		}
		return errors.New(FormatConditionError(field, fieldValue, flowlinks))
	}

	// 预加载匹配流转关系的下一步信息 / Preload NextProcess for matched flowlink
	var withFlowlink models.Flowlink
	query.Model(&models.Flowlink{}).With("NextProcess").Where("id=?", matchedFlowlink.ID).First(&withFlowlink)

	// 计算下一步审批人 / Calculate next step auditors
	auditor_ids := w.GetProcessAuditorIds(proc.Entry, withFlowlink.NextProcessID)
	if len(auditor_ids) == 0 {
		return errors.New("未找到下一步骤审批人")
	}

	// 为下一步创建审批任务 / Create procs for next step
	if err := w.createProcsForProcess(query, &proc.Entry, withFlowlink.NextProcessID, withFlowlink.NextProcess.ProcessName, auditor_ids); err != nil {
		return err
	}

	// 更新 Entry 的当前步骤 / Update entry's current process
	query.Model(&models.Entry{}).Where("id=?", proc.EntryID).Update("process_id", cast.ToUint(withFlowlink.NextProcessID))

	// 如果是子流程，更新父流程的 child 字段 / Update parent entry child field if needed
	if proc.Entry.Pid > 0 {
		parentEntry := models.Entry{}
		query.Model(&models.Entry{}).Where("pid=?", proc.Entry.Pid).First(&parentEntry)
		if parentEntry.ID > 0 {
			query.Model(&models.Entry{}).Where("id=?", parentEntry.ID).Update("child", withFlowlink.NextProcessID)
		}
	}

	return w.finalizeProc(query, proc, content, emp, withFlowlink.NextProcessID)
}

// handleChildWorkflow 创建或恢复子流程。
// 如果子流程已存在则恢复，否则新建并初始化。
// handleChildWorkflow creates or resumes a child workflow
func (w *Workflow) handleChildWorkflow(query orm.Query, proc *models.Proc, fklink *models.Flowlink, content string, emp models.Emp) error {
	// 确定子流程 ID — 可能在当前步骤或下一步骤上配置
	// Determine the child flow ID — may be on current step or next step
	childFlowID := fklink.Process.ChildFlowID
	if childFlowID == 0 {
		childFlowID = fklink.NextProcess.ChildFlowID
	}
	if childFlowID == 0 {
		return errors.New("未配置子流程")
	}

	// 查找是否已存在子流程实例 / Look for existing child entry
	child_entry := models.Entry{}
	query.Model(&models.Entry{}).
		Where("pid=?", proc.Entry.ID).
		Where("circle=?", proc.Entry.Circle).
		First(&child_entry)

	if child_entry.ID == 0 {
		// 创建新的子流程实例 / Create new child entry
		// 确定入口步骤：如果子流程入口在当前步骤的 NextProcess 上，用 NextProcessID 做 enterProcessID
		enterProcessID := cast.ToInt(fklink.ProcessID)
		if fklink.Process.ChildFlowID == 0 && fklink.NextProcess.ChildFlowID > 0 {
			enterProcessID = fklink.NextProcessID
		}
		newChild := models.Entry{
			Title:          proc.Entry.Title,      // 继承父流程标题 / Inherit parent title
			FlowID:         cast.ToUint(childFlowID),
			EmpID:          cast.ToUint(proc.Entry.EmpID),
			Status:         models.EntryStatusPending,
			Pid:            cast.ToInt(proc.Entry.ID), // 关联父流程 / Link to parent entry
			Circle:         proc.Entry.Circle,
			EnterProcessID: enterProcessID,
			EnterProcID:    cast.ToInt(proc.ID),
		}
		query.Model(&models.Entry{}).Create(&newChild)

		// 重新加载完整的子流程信息 / Reload full child entry info
		query.Model(&models.Entry{}).Where("id=?", newChild.ID).
			With("Flow").With("Process").With("EnterProcess").With("Emp.Dept").First(&child_entry)
	} else {
		// 恢复已有的子流程 / Resume existing child entry
		query.Model(&models.Entry{}).Where("id=?", child_entry.ID).
			With("Flow").With("Process").With("EnterProcess").With("Emp.Dept").First(&child_entry)
	}

	// 查找子流程的第一步流转关系（position=0, 非 Condition 类型）
	// Find child flow's first step flowlink (position=0, non-Condition type)
	var childProcessID uint
	query.Model(&models.Process{}).Where("flow_id=? AND position=?", childFlowID, 0).Pluck("id", &childProcessID)
	if childProcessID == 0 {
		return errors.New("子流程未配置第一步")
	}
	child_flowlink := models.Flowlink{}
	query.Model(&models.Flowlink{}).
		Where("flow_id=? AND process_id=? AND type!=?", childFlowID, childProcessID, "Condition").
		Order("sort ASC").First(&child_flowlink)

	var resolvedFlowlink models.Flowlink
	query.Model(&models.Flowlink{}).Where("id=?", child_flowlink.ID).With("Process").With("NextProcess").First(&resolvedFlowlink)

	// 初始化子流程的第一步审批 / Initialize child's first process auditors
	if err := w.SetFirstProcessAuditor(child_entry, resolvedFlowlink); err != nil {
		return err
	}

	// 更新父流程的 child 字段，指向子流程当前步骤 / Update parent's child field to current child step
	query.Model(&models.Entry{}).Where("id=?", child_entry.Pid).Update("child", child_entry.ProcessID)

	// 如果子流程自动完成（第一步也是最后一步），立即处理父子联动。
	// If the child flow auto-completed (first step was also the last step),
	// handle parent-child linkage immediately.
	var childAfterEntry models.Entry
	query.Model(&models.Entry{}).Where("id=?", child_entry.ID).
		With("EnterProcess").First(&childAfterEntry)
	if childAfterEntry.Status == models.EntryStatusCompleted {
		return w.handleParentAfterChildComplete(query, &models.Proc{
			EntryID: childAfterEntry.ID,
			Entry:   childAfterEntry,
		}, content, emp)
	}

	return w.finalizeProc(query, proc, content, emp, cast.ToInt(fklink.ProcessID))
}

// handleLastStep 完成流程的最后一步：标记流程为已完成，处理父子流程联动。
// handleLastStep completes the entry and handles parent workflow linkage
func (w *Workflow) handleLastStep(query orm.Query, proc *models.Proc, fklink *models.Flowlink, content string, emp models.Emp) error {
	procEntry := models.Entry{}
	query.Model(&models.Entry{}).Where("id=?", proc.EntryID).First(&procEntry)
	procEntry.Status = 9 // 已完成 / Completed
	procEntry.ProcessID = fklink.ProcessID
	query.Model(&models.Entry{}).Where("id=?", procEntry.ID).Save(&procEntry)
	// 归档流程数据 / Archive entry data
	w.archiveEntry(proc.EntryID, models.ProcStatusConsensus)

	// 先完成当前审批任务（标记为已通过），再处理父子联动
	// Finalize current proc first (mark as approved) before handling parent linkage
	if err := w.finalizeProc(query, proc, content, emp, cast.ToInt(fklink.ProcessID)); err != nil {
		return err
	}

	// 如果有父流程，处理子流程完成后的父流程联动 / Handle parent linkage if this is a child entry
	if proc.Entry.Pid > 0 {
		return w.handleParentAfterChildComplete(query, proc, content, emp)
	}

	return nil
}

// handleParentAfterChildComplete 子流程完成后处理父流程的走向。
// 根据 ChildAfter 配置决定是同时结束父流程还是返回父流程继续。
// handleParentAfterChildComplete handles parent workflow when child completes
func (w *Workflow) handleParentAfterChildComplete(query orm.Query, proc *models.Proc, content string, emp models.Emp) error {
	parentEntry := models.Entry{}
	query.Model(&models.Entry{}).Where("id=?", proc.Entry.Pid).First(&parentEntry)

	// 查找父流程中配置了子流程入口的步骤 — child_flow_id 匹配子流程的步骤
	// Find the child workflow entry step on the parent flow — the process where child_flow_id matches
	var childEntryProcess models.Process
	query.Model(&models.Process{}).Where("flow_id=? AND child_flow_id=?", parentEntry.FlowID, proc.Entry.FlowID).First(&childEntryProcess)
	if childEntryProcess.ID == 0 {
		// 备选方案：用子流程的 enter_process_id 匹配 / Fallback: try matching by the child entry's enter_process_id
		query.Model(&models.Process{}).Where("id=?", proc.Entry.EnterProcessID).First(&childEntryProcess)
	}
	if childEntryProcess.ID == 0 {
		// 最后备选方案：直接结束父流程 / Last fallback: just end the parent
		parentEntry.Status = models.EntryStatusCompleted
		parentEntry.Child = 0
		query.Model(&models.Entry{}).Where("id=?", parentEntry.ID).Save(&parentEntry)
		w.NotifySendOne(parentEntry.EmpID)
		return nil
	}

	// ChildAfter == 1: 子流程完成后同时结束父流程
	// ChildAfter == 1: end the parent workflow as well
	if childEntryProcess.ChildAfter == 1 {
		parentEntry.Status = models.EntryStatusCompleted
		parentEntry.Child = 0
		query.Model(&models.Entry{}).Where("id=?", parentEntry.ID).Save(&parentEntry)
		w.archiveEntry(parentEntry.ID, models.EntryStatusCompleted)
		w.NotifySendOne(proc.Entry.ID)
	} else {
		// ChildAfter == 2: 子流程完成后返回父流程继续
		// ChildAfter == 2: return to parent workflow
		if childEntryProcess.ChildBackProcess > 0 {
			// 回到父流程的指定步骤 / Return to the specified step in parent workflow
			w.goToProcess(query, &parentEntry, childEntryProcess.ChildBackProcess)
		} else {
			// 进入父流程的下一步 / Go to parent's next step
			parentFlowlink := models.Flowlink{}
			query.Model(&models.Flowlink{}).Where("process_id=?", childEntryProcess.ID).Where("type != ?", "Condition").First(&parentFlowlink)
			if parentFlowlink.NextProcessID == -1 {
				// 父流程最后一步：结束 / Parent's last step: complete
				parentEntry.Status = models.EntryStatusCompleted
				parentEntry.Child = 0
				query.Model(&models.Entry{}).Where("id=?", parentEntry.ID).Save(&parentEntry)
				w.archiveEntry(parentEntry.ID, models.EntryStatusCompleted)
				w.NotifySendOne(parentEntry.EmpID)
			} else {
				// 进入父流程下一步 / Advance to parent's next step
				w.goToProcess(query, &parentEntry, parentFlowlink.NextProcessID)
				parentEntry.ProcessID = cast.ToUint(parentFlowlink.NextProcessID)
				parentEntry.Status = models.EntryStatusPending
				query.Model(&models.Entry{}).Where("id=?", parentEntry.ID).Save(&parentEntry)
				w.NotifySendOne(cast.ToUint(proc.AuditorID))
			}
		}
		parentEntry.Child = 0
		query.Model(&models.Entry{}).Where("id=?", parentEntry.ID).Save(&parentEntry)
	}

	return nil
}

// handleNextStep 为下一步骤创建审批任务并推进流程。
// 如果下一步没有指定审批人，则视为最后一步并完成流程。
// handleNextStep creates procs for the next process and continues the workflow
func (w *Workflow) handleNextStep(query orm.Query, proc *models.Proc, fklink *models.Flowlink, content string, emp models.Emp) error {
	// 计算下一步审批人 / Calculate auditors for next step
	auditor_ids := w.GetProcessAuditorIds(proc.Entry, fklink.NextProcessID)
	if len(auditor_ids) == 0 {
		// 下一步没有指定审批人 — 视为最后一步（完成流程）
		// Next process has no designated approvers — treat as last step (completion)
		return w.handleLastStep(query, proc, fklink, content, emp)
	}

	// 创建下一步的审批任务 / Create approval tasks for next step
	if err := w.createProcsForProcess(query, &proc.Entry, fklink.NextProcessID, fklink.NextProcess.ProcessName, auditor_ids); err != nil {
		return err
	}

	// 更新流程当前步骤 / Update entry's current process
	procEntry := models.Entry{}
	query.Model(&models.Entry{}).Where("id=?", proc.Entry.ID).First(&procEntry)
	procEntry.ProcessID = cast.ToUint(fklink.NextProcessID)
	query.Model(&models.Entry{}).Where("id=?", procEntry.ID).Save(&procEntry)

	// 如果存在父流程，更新父流程的 child 字段 / Update parent entry child field if needed
	var parentEntry models.Entry
	query.Model(&models.Entry{}).Where("id=?", proc.Entry.Pid).First(&parentEntry)
	if parentEntry.ID > 0 {
		parentEntry.Child = fklink.NextProcessID
		query.Model(&models.Entry{}).Where("id=?", parentEntry.ID).Save(&parentEntry)
	}

	return w.finalizeProc(query, proc, content, emp, fklink.NextProcessID)
}

// finalizeProc 标记当前审批任务为已通过，并触发插件执行和抄送。
// finalizeProc marks the current proc as approved and triggers plugins/CC
func (w *Workflow) finalizeProc(query orm.Query, proc *models.Proc, content string, emp models.Emp, nextProcessID int) error {
	// 仅在安装了 distribute 插件时查询插件配置 / Plugin configs only exist if the distribute plugin has been installed
	var pluginConfigs []official_plugins.PluginConfig
	if facades.Schema().HasTable("plugin_configs") {
		query.Model(official_plugins.PluginConfig{}).Where("process_id=?", proc.ProcessID).Find(&pluginConfigs)
	}
	pluginConfigJSON, _ := json.Marshal(pluginConfigs)

	// 将当前审批任务标记为已通过 / Mark current proc as approved
	var currentProc models.Proc
	query.Model(&models.Proc{}).
		Where("entry_id=?", proc.EntryID).
		Where("process_id=?", proc.ProcessID).
		Where("circle=?", proc.Entry.Circle).
		Where("status=?", models.ProcStatusPending).
		First(&currentProc)
	if currentProc.ID > 0 {
		currentProc.Status = models.ProcStatusApproved
		currentProc.AuditorID = cast.ToInt(emp.ID)
		currentProc.AuditorName = emp.Name
		currentProc.DeptName = emp.Dept.DeptName
		currentProc.Content = content
		currentProc.Beizhu = string(pluginConfigJSON)     // 存储插件配置 JSON 作为备注 / Store plugin configs as notes
		currentProc.Concurrence = carbon.NewDateTime(carbon.Now())
		query.Model(&models.Proc{}).Where("id=?", currentProc.ID).Save(&currentProc)
	}

	// 执行分发插件 / Execute distribute plugin
	w.ExecPluginMethod("DistributePlugin", cast.ToUint(proc.FlowID), cast.ToUint(proc.ProcessID))
	// 触发抄送 / Trigger CC
	w.triggerCC(proc.EntryID, cast.ToUint(proc.FlowID), cast.ToUint(proc.ProcessID), proc.ID)

	return nil
}

// handleRejectEntry 处理流程级别的驳回：将 Entry 标记为已驳回，并通知发起人。
// handleRejectEntry handles entry-level rejection (mark entry as rejected, notify initiator)
func (w *Workflow) handleRejectEntry(query orm.Query, proc *models.Proc) error {
	procEntry := models.Entry{}
	query.Model(&models.Entry{}).Where("id=?", proc.EntryID).First(&procEntry)
	procEntry.Status = models.ProcStatusRejected
	query.Model(&models.Entry{}).Where("id=?", procEntry.ID).Save(&procEntry)

	// 如果是子流程，同步更新父流程状态 / Sync parent entry status if this is a child entry
	if proc.Entry.Pid > 0 {
		parentEntry := models.Entry{}
		query.Model(&models.Entry{}).Where("id=?", proc.Entry.Pid).First(&parentEntry)
		parentEntry.Child = proc.ProcessID
		parentEntry.Status = models.ProcStatusRejected
		query.Model(&models.Entry{}).Where("id=?", parentEntry.ID).Save(&parentEntry)
	}

	// 通知流程发起人 / Notify the entry initiator
	w.NotifySendOne(proc.Entry.EmpID)
	return nil
}

// goToProcess 在指定流程步骤创建审批任务。
// 用于子流程返回父流程指定步骤、驳回到指定节点等场景。
// goToProcess creates approval tasks at a specified process step
func (w *Workflow) goToProcess(query orm.Query, entry *models.Entry, processID int) error {
	auditor_ids := w.GetProcessAuditorIds(*entry, processID)
	if len(auditor_ids) == 0 {
		return errors.New("未找到审批人")
	}

	var processName string
	query.Model(&models.Process{}).Where("id=?", processID).Select("process_name").First(&processName)

	return w.createProcsForProcess(query, entry, processID, processName, auditor_ids)
}

// Pass 是 Transfer 的别名，用于审批通过操作。
// Pass is an alias for Transfer
func (w *Workflow) Pass(process_id int, user models.Emp, content string, formData map[string]any) error {
	return w.Transfer(process_id, user, content, formData)
}

// UnPass 驳回当前审批任务，回退到上一步。
// UnPass rejects the current approval task and sends it back to the previous step
func (w *Workflow) UnPass(proc_id int, user models.Emp, content string) error {
	// 默认驳回到上一步（targetProcessID=0） / Default: reject to previous step (targetProcessID=0)
	return w.UnPassTo(proc_id, user, content, 0)
}

// UnPassTo 驳回到指定的目标步骤（0 表示上一步）。
// UnPassTo rejects to a specific target process step (0 = previous step)
func (w *Workflow) UnPassTo(proc_id int, user models.Emp, content string, targetProcessID int) error {
	query := facades.Orm().Query()
	var emp models.Emp
	query.Model(&models.Emp{}).Where("user_id=?", user.ID).First(&emp)

	// 查找审批任务 / Find the proc
	var proc models.Proc
	query.Model(&models.Proc{}).Where("id=?", proc_id).With("Entry").First(&proc)
	if proc.ID == 0 {
		return errors.New("审批任务不存在")
	}

	// 如果指定了目标步骤，驳回到指定节点 / If targetProcessID specified, reject to specific node
	if targetProcessID > 0 {
		return w.rejectToNode(query, &proc, emp, content, targetProcessID)
	}

	// 否则驳回到上一步 / Otherwise reject to previous step
	return w.rejectPreviousStep(query, &proc, emp, content)
}

// rejectPreviousStep 实现原始的驳回回上一步逻辑。
// rejectPreviousStep implements the original reject-to-previous-step logic
func (w *Workflow) rejectPreviousStep(query orm.Query, proc *models.Proc, emp models.Emp, content string) error {
	// 查找当前步骤待处理的审批任务 / Find the pending proc at current step
	var todoProc models.Proc
	query.Model(&models.Proc{}).
		Where("entry_id=?", proc.EntryID).
		Where("process_id=?", proc.ProcessID).
		Where("circle=?", proc.Entry.Circle).
		Where("status=?", models.ProcStatusPending).
		First(&todoProc)

	// 如果没有待处理任务，直接驳回整个流程 / If no pending task, reject the entire entry
	if todoProc.ID == 0 {
		return w.handleRejectEntry(query, proc)
	}

	// 标记待处理任务为已驳回 / Mark the pending task as rejected
	todoProc.Status = models.ProcStatusRejected
	todoProc.AuditorID = cast.ToInt(emp.ID)
	todoProc.AuditorName = emp.Name
	todoProc.AuditorDept = emp.Dept.DeptName
	todoProc.Content = content
	todoProc.IsRead = 1
	todoProc.Concurrence = carbon.NewDateTime(carbon.Now())
	query.Model(&models.Proc{}).Where("id=?", todoProc.ID).Save(&todoProc)

	return w.handleRejectEntry(query, proc)
}

// rejectToNode 实现驳回到任意指定节点的逻辑。
// 将目标节点之前的审批任务标记为跳过，重置目标节点为待处理状态。
// rejectToNode implements reject-to-arbitrary-node logic
func (w *Workflow) rejectToNode(query orm.Query, proc *models.Proc, emp models.Emp, content string, targetProcessID int) error {
	// 查找当前流程的所有待处理任务 / Find all pending procs for this entry
	var allPendingProcs []models.Proc
	query.Model(&models.Proc{}).
		Where("entry_id=?", proc.EntryID).
		Where("circle=?", proc.Entry.Circle).
		Where("status=?", models.ProcStatusPending).
		Find(&allPendingProcs)

	// 在待处理列表中查找目标步骤 / Find the target proc among pending ones
	var targetProc models.Proc
	found := false
	for i := range allPendingProcs {
		if allPendingProcs[i].ProcessID == targetProcessID {
			targetProc = allPendingProcs[i]
			found = true
			break
		}
	}

	if !found {
		// 目标步骤不在待处理列表中 — 在所有历史任务中搜索
		// Target process doesn't have a pending proc — search all procs
		query.Model(&models.Proc{}).
			Where("entry_id=?", proc.EntryID).
			Where("circle=?", proc.Entry.Circle).
			Where("process_id=?", targetProcessID).
			Order("id DESC").
			First(&targetProc)

		if targetProc.ID == 0 {
			return errors.New("目标审批节点不存在")
		}

		// 将当前驳回的审批任务标记为跳过 / Save content on the rejecting proc before marking as skipped
		proc.Status = models.ProcStatusSkipped
		proc.Content = content
		proc.AuditorID = cast.ToInt(emp.ID)
		proc.AuditorName = emp.Name
		query.Model(&models.Proc{}).Where("id=?", proc.ID).Save(proc)

		// 将目标节点和当前节点之间的所有待处理/已通过任务标记为跳过
		// Mark all remaining pending/approved procs as skipped (between target and current)
		var remainingProcs []models.Proc
		query.Model(&models.Proc{}).
			Where("entry_id=?", proc.EntryID).
			Where("circle=?", proc.Entry.Circle).
			Where("status IN (?, ?)", models.ProcStatusPending, models.ProcStatusApproved).
			Find(&remainingProcs)
		for i := range remainingProcs {
			remainingProcs[i].Status = models.ProcStatusSkipped
			query.Model(&models.Proc{}).Where("id=?", remainingProcs[i].ID).Save(&remainingProcs[i])
		}

		// 重置目标审批任务为待处理 / Reset target proc to pending
		targetProc.Status = models.ProcStatusPending
		targetProc.IsRead = 0
		targetProc.Concurrence = carbon.NewDateTime(carbon.Now())
		query.Model(&models.Proc{}).Where("id=?", targetProc.ID).Save(&targetProc)

		// 更新流程当前步骤指向目标 / Update entry to point to target process
		procEntry := models.Entry{}
		query.Model(&models.Entry{}).Where("id=?", proc.EntryID).First(&procEntry)
		procEntry.Status = models.ProcStatusPending
		procEntry.ProcessID = cast.ToUint(targetProcessID)
		query.Model(&models.Entry{}).Where("id=?", procEntry.ID).Save(&procEntry)

		// 通知目标审批人和发起人 / Notify target auditor and initiator
		w.NotifyNextAuditor(uint(targetProc.EmpID))
		w.NotifySendOne(proc.Entry.EmpID)
		w.archiveEntry(proc.EntryID, models.ProcStatusRejected)
		return nil
	}

	// 目标任务存在且为待处理状态 — 将目标之前的所有未完成任务标记为跳过
	// Target proc exists and is pending — mark current procs as skipped
	var currentProcs []models.Proc
	query.Model(&models.Proc{}).
		Where("entry_id=?", proc.EntryID).
		Where("circle=?", proc.Entry.Circle).
		Where("status IN (?, ?)", models.ProcStatusPending, models.ProcStatusApproved).
		Order("id ASC").
		Find(&currentProcs)

	skipped := false
	for i := range currentProcs {
		// 从目标节点之后的任务全部跳过 / Skip all procs after the target
		if skipped {
			currentProcs[i].Status = models.ProcStatusSkipped
			query.Model(&models.Proc{}).Where("id=?", currentProcs[i].ID).Save(&currentProcs[i])
		}
		if currentProcs[i].ProcessID == targetProcessID {
			skipped = true
		}
	}

	// 将驳回操作的审批任务标记为跳过 / Mark the rejecting proc as skipped
	proc.Status = models.ProcStatusSkipped
	proc.Content = content
	proc.AuditorID = cast.ToInt(emp.ID)
	proc.AuditorName = emp.Name
	query.Model(&models.Proc{}).Where("id=?", proc.ID).Save(proc)

	// 通知发起人 / Notify the initiator
	w.NotifySendOne(proc.Entry.EmpID)
	return nil
}

// ExecPluginMethod 执行指定插件。
// ExecPluginMethod executes a plugin
func (w *Workflow) ExecPluginMethod(plugin_name string, flowID uint, processID uint) error {
	ctor := GetCollectorIns()
	return ctor.DoPluginsExec(plugin_name, flowID, processID)
}

// Revoke 允许发起人撤回自己的待审批流程。
// 只有流程状态为"进行中"且尚未被任何人处理时才能撤回。
// Revoke allows the initiator to withdraw a pending entry
func (w *Workflow) Revoke(entryID uint, user models.Emp) error {
	return facades.Orm().Transaction(func(tx orm.Query) error {
		var entry models.Entry
		tx.Model(&models.Entry{}).Where("id=?", entryID).With("Emp").First(&entry)
		if entry.ID == 0 {
			return errors.New("流程不存在")
		}
		// 校验：只有发起人才能撤回 / Verify: only initiator can revoke
		if cast.ToInt(entry.EmpID) != cast.ToInt(user.ID) {
			return errors.New("只有发起人才能撤回流程")
		}
		// 校验：只有进行中的流程才能撤回 / Verify: only pending entries can be revoked
		if entry.Status != models.EntryStatusPending {
			return errors.New("当前流程状态不允许撤回")
		}

		// 检查是否有审批人已处理（auditor_id 不为 0 表示已被处理）
		// Verify: no proc has been handled yet (auditor_id != 0 means it's been processed)
		var pendingProcs []models.Proc
		tx.Model(&models.Proc{}).Where("entry_id=?", entryID).Where("status=?", models.ProcStatusPending).Find(&pendingProcs)
		for _, p := range pendingProcs {
			if p.AuditorID != 0 {
				return errors.New("流程已被处理，无法撤回")
			}
		}

		// 更新流程状态为已撤回 / Update entry status to revoked
		entry.Status = models.ProcStatusRevoked
		tx.Model(&models.Entry{}).Where("id=?", entryID).Save(&entry)
		w.archiveEntry(entryID, models.ProcStatusRevoked)

		// 批量将待处理审批任务标记为已撤回 / Batch mark pending procs as revoked
		for _, p := range pendingProcs {
			p.Status = models.ProcStatusRevoked
			p.AuditorID = cast.ToInt(user.ID)
			p.AuditorName = user.Name
			tx.Model(&models.Proc{}).Where("id=?", p.ID).Save(&p)
		}
		return nil
	})
}

// AddSign 为当前审批任务添加额外的审批人（加签）。
// signType: "before"=前加签（在被加签人之前审批）, "after"=后加签（在被加签人之后审批）。
// AddSign adds an additional approver to the current approval task
func (w *Workflow) AddSign(entryID uint, processID int, signEmpID int, signType string, currentUser models.Emp) error {
	return facades.Orm().Transaction(func(tx orm.Query) error {
		// 校验流程存在且状态允许加签 / Verify entry exists and allows add-sign
		var entry models.Entry
		tx.Model(&models.Entry{}).Where("id=?", entryID).First(&entry)
		if entry.ID == 0 {
			return errors.New("流程不存在")
		}
		if entry.Status != models.EntryStatusPending {
			return errors.New("流程状态不允许加签")
		}

		// 查找当前用户的待处理审批任务 / Find current user's pending proc
		var targetProc models.Proc
		tx.Model(&models.Proc{}).
			Where("entry_id=?", entryID).
			Where("process_id=?", processID).
			Where("emp_id=?", currentUser.ID).
			Where("status=?", models.ProcStatusPending).
			First(&targetProc)
		if targetProc.ID == 0 {
			return errors.New("未找到当前审批任务")
		}

		// 验证被加签员工存在 / Verify the sign employee exists
		var signEmp models.Emp
		tx.Model(&models.Emp{}).Where("id=?", signEmpID).With("Dept").First(&signEmp)
		if signEmp.ID == 0 {
			return errors.New("被加签员工不存在")
		}

		// 创建加签记录 / Create add-sign record
		sign := models.ProcAddSign{
			EntryID:     entryID,
			ProcID:      targetProc.ID,
			SignType:    signType,     // 加签类型：before / after
			SignEmpID:   signEmpID,
			SignEmpName: signEmp.Name,
			Status:      models.ProcStatusPending,
		}
		tx.Model(&models.ProcAddSign{}).Create(&sign)

		// 为被加签人创建新的审批任务 / Create new proc for the added signer
		newProc := models.Proc{
			EntryID:     entry.ID,
			FlowID:      cast.ToInt(entry.FlowID),
			ProcessID:   targetProc.ProcessID,
			ProcessName: targetProc.ProcessName,
			EmpID:       cast.ToInt(signEmp.ID),
			EmpName:     signEmp.Name,
			DeptName:    signEmp.Dept.DeptName,
			Circle:      targetProc.Circle,
			Status:      models.ProcStatusPending,
			IsRead:      0,
			Concurrence: carbon.NewDateTime(carbon.Now()),
		}
		tx.Model(&models.Proc{}).Create(&newProc)

		// 通知新加签的审批人 / Notify the newly added signer
		w.NotifyNextAuditor(uint(signEmp.ID))

		return nil
	})
}

// TransferProc 将当前审批任务转交给另一个员工。
// 原任务标记为"已转交"，创建新任务给目标员工。
// TransferProc transfers the current approval task to another employee
func (w *Workflow) TransferProc(entryID uint, procID uint, targetEmpID int, currentUser models.Emp) error {
	return facades.Orm().Transaction(func(tx orm.Query) error {
		// 校验审批任务存在、匹配且未被处理 / Verify proc exists, matches entry, and is still pending
		var targetProc models.Proc
		tx.Model(&models.Proc{}).Where("id=?", procID).First(&targetProc)
		if targetProc.ID == 0 {
			return errors.New("审批任务不存在")
		}
		if cast.ToInt(targetProc.EntryID) != cast.ToInt(entryID) {
			return errors.New("审批任务与流程不匹配")
		}
		if targetProc.Status != models.ProcStatusPending {
			return errors.New("审批任务已处理，无法转交")
		}

		// 校验流程状态 / Verify entry status allows transfer
		var entry models.Entry
		tx.Model(&models.Entry{}).Where("id=?", entryID).First(&entry)
		if entry.Status != models.EntryStatusPending {
			return errors.New("流程状态不允许转交")
		}

		// 验证被转交员工存在 / Verify target employee exists
		var targetEmp models.Emp
		tx.Model(&models.Emp{}).Where("id=?", targetEmpID).With("Dept").First(&targetEmp)
		if targetEmp.ID == 0 {
			return errors.New("被转交员工不存在")
		}

		// 为目标员工创建新的审批任务 / Create new proc for the target employee
		newProc := models.Proc{
			EntryID:     entry.ID,
			FlowID:      cast.ToInt(entry.FlowID),
			ProcessID:   targetProc.ProcessID,
			ProcessName: targetProc.ProcessName,
			EmpID:       cast.ToInt(targetEmp.ID),
			EmpName:     targetEmp.Name,
			DeptName:    targetEmp.Dept.DeptName,
			Circle:      targetProc.Circle,
			Status:      models.ProcStatusPending,
			IsRead:      0,
			Concurrence: carbon.NewDateTime(carbon.Now()),
		}
		tx.Model(&models.Proc{}).Create(&newProc)

		// 标记原任务为已转交 / Mark original proc as transferred
		targetProc.Status = models.ProcStatusTransferred
		targetProc.AuditorID = cast.ToInt(currentUser.ID)
		targetProc.AuditorName = currentUser.Name
		targetProc.Content = "已转交给" + targetEmp.Name
		tx.Model(&models.Proc{}).Where("id=?", procID).Save(&targetProc)

		return nil
	})
}

// AddComment 为流程添加评论。支持通过 parentID/replyTo 实现回复嵌套。
// AddComment adds a comment to an entry. Supports reply threading via parentID/replyTo.
func (w *Workflow) AddComment(entryID uint, procID uint, empID int, empName string,
	content string, parentID uint, replyToEmpID int, replyToEmpName string) error {
	comment := models.ProcComment{
		EntryID:        entryID,
		ProcID:          procID,
		EmpID:          empID,
		EmpName:        empName,
		Content:        content,
		Status:         1, // 有效 / Active
		ParentID:       parentID,       // 父评论 ID（用于嵌套回复） / Parent comment ID (for threaded replies)
		ReplyToEmpID:   replyToEmpID,   // 回复目标员工 ID / Reply target employee ID
		ReplyToEmpName: replyToEmpName, // 回复目标员工名称 / Reply target employee name
	}
	return facades.Orm().Query().Model(&models.ProcComment{}).Create(&comment)
}

// GetComments 获取一个流程的所有评论，以树形结构返回（根评论包含子回复）。
// GetComments retrieves all comments for an entry as a tree structure.
func (w *Workflow) GetComments(entryID uint) ([]models.ProcComment, error) {
	// 查询所有有效评论 / Query all active comments
	var flat []models.ProcComment
	err := facades.Orm().Query().
		Model(&models.ProcComment{}).
		Where("entry_id=? AND status=?", entryID, 1).
		Order("id asc").
		Find(&flat)
	if err != nil {
		return nil, err
	}

	// 构建树形结构：先将子评论附加到父评论，再收集根评论
	// Build tree: first attach children, then collect roots
	commentMap := make(map[uint]*models.ProcComment)
	for i := range flat {
		flat[i].Children = []models.ProcComment{}
		commentMap[flat[i].ID] = &flat[i]
	}
	// 将子评论挂载到父评论的 Children 字段 / Attach children to their parent's Children slice
	for i := range flat {
		if flat[i].ParentID > 0 {
			if parent, ok := commentMap[flat[i].ParentID]; ok {
				parent.Children = append(parent.Children, flat[i])
			}
		}
	}
	// 收集根评论（ParentID == 0）/ Now collect roots — after all children have been attached
	var roots []models.ProcComment
	for i := range flat {
		if flat[i].ParentID == 0 {
			roots = append(roots, flat[i])
		}
	}
	return roots, nil
}

// triggerCC 在审批任务通过后创建抄送记录。
// 根据 Process 的 cc_emp_ids 配置找到抄送人并批量创建 CcRecord。
// triggerCC creates CC records after a proc is approved
func (w *Workflow) triggerCC(entryID, flowID, processID, procID uint) {
	// 读取步骤的抄送人配置 / Read CC employee IDs from process config
	var ccEmpIDs []string
	facades.Orm().Query().
		Model(&models.Process{}).
		Where("id=?", processID).
		Pluck("cc_emp_ids", &ccEmpIDs)

	if len(ccEmpIDs) == 0 || ccEmpIDs[0] == "" {
		return
	}

	// 解析逗号分隔的员工 ID / Parse comma-separated employee IDs
	var empIDs []int
	for _, idStr := range strings.Split(ccEmpIDs[0], ",") {
		if id := cast.ToInt(strings.TrimSpace(idStr)); id > 0 {
			empIDs = append(empIDs, id)
		}
	}
	if len(empIDs) == 0 {
		return
	}

	// 查询抄送人信息 / Query employee info for CC
	var emps []models.Emp
	facades.Orm().Query().Model(&models.Emp{}).Where("id IN (?)", empIDs).Find(&emps)

	// 为每个抄送人创建抄送记录 / Create CC record for each employee
	for _, emp := range emps {
		record := models.CcRecord{
			EntryID:   entryID,
			FlowID:    flowID,
			ProcessID: processID,
			ProcID:    procID,
			EmpID:     cast.ToInt(emp.ID),
			EmpName:   emp.Name,
			Status:    0, // 未读 / Unread
		}
		facades.Orm().Query().Model(&models.CcRecord{}).Create(&record)
	}
}

// archiveEntry 在流程完成（通过/驳回/撤回）时创建一个完整的流程快照。
// 所有动态数据被序列化为 JSON 存储，确保即使员工离职后记录仍可读。
// archiveEntry creates a complete snapshot of the entry when it finishes
// (approved/rejected/revoked). All dynamic data is serialized to JSON so the
// record remains readable even after employees leave the organization.
func (w *Workflow) archiveEntry(entryID uint, finalStatus int) {
	query := facades.Orm().Query()

	// 1. 加载完整的流程实例及其所有关联数据 / Load full entry with all related data
	var entry models.Entry
	query.Model(&models.Entry{}).
		Where("id = ?", entryID).
		With("Emp.Dept").
		With("Flow").
		With("Process").
		With("EntryDatas").
		With("Procs").
		First(&entry)
	if entry.ID == 0 {
		return
	}

	// 2. 加载评论记录 / Load comments
	var comments []models.ProcComment
	query.Model(&models.ProcComment{}).Where("entry_id = ? AND status = 1", entryID).Find(&comments)

	// 3. 加载抄送记录 / Load cc records
	var ccRecords []models.CcRecord
	query.Model(&models.CcRecord{}).Where("entry_id = ?", entryID).Find(&ccRecords)

	// 4. 序列化为 JSON / Serialize to JSON
	flowJSON, _ := json.Marshal(entry.Flow)
	entryJSON, _ := json.Marshal(entry)
	formDataJSON, _ := json.Marshal(entry.EntryDatas)
	procsJSON, _ := json.Marshal(entry.Procs)
	commentsJSON, _ := json.Marshal(comments)
	ccJSON, _ := json.Marshal(ccRecords)

	// 5. Upsert 操作 — 如果该流程已有归档，则替换旧记录
	// Upsert — replace existing archive for this entry if resubmitted
	archive := models.EntryArchive{
		EntryID:          entry.ID,
		FlowID:           entry.FlowID,
		Status:           finalStatus,
		Title:            entry.Title,
		FlowSnapshot:     string(flowJSON),     // 流程定义快照 / Flow definition snapshot
		EntrySnapshot:    string(entryJSON),    // 流程实例快照 / Entry instance snapshot
		FormDataSnapshot: string(formDataJSON), // 表单数据快照 / Form data snapshot
		ProcsSnapshot:    string(procsJSON),    // 审批任务快照 / Approval tasks snapshot
		CommentsSnapshot: string(commentsJSON), // 评论快照 / Comments snapshot
		CCSnapshot:       string(ccJSON),       // 抄送记录快照 / CC records snapshot
	}

	var existing models.EntryArchive
	query.Model(&models.EntryArchive{}).Where("entry_id = ?", entryID).First(&existing)
	if existing.ID > 0 {
		query.Model(&models.EntryArchive{}).Where("id = ?", existing.ID).Save(&archive)
	} else {
		query.Model(&models.EntryArchive{}).Create(&archive)
	}
}
