package models

import "fmt"

// =============================================================================
// Entry State Machine 流程实例状态机
// =============================================================================
//
// Valid transitions 有效状态转换:
//
//	                    +-- Resend 重新发起 ----+
//	                    v                       |
//	Pending(0) ──┬──> Completed(9) 完成        |
//	             ├──> Rejected(-1) 驳回 ───────+
//	             └──> Revoked(-2) 撤回 ────────+
//
//	Child workflows 子流程: when a child completes with ChildAfter=2,
//	当子流程完成且 ChildAfter=2 时,
//	the parent Entry stays at Pending(0) and its ProcessID advances.
//	父流程实例保持 Pending(0) 状态并推进其 ProcessID。

// EntryTransition represents a valid state transition for an Entry.
// EntryTransition 表示流程实例的有效状态转换。
type EntryTransition struct {
	// From is the source status code 源状态码
	From int
	// To is the destination status code 目标状态码
	To   int
}

// validEntryTransitions defines all permitted state-to-state moves for Entry.
// validEntryTransitions 定义流程实例所有允许的状态到状态转换。
var validEntryTransitions = map[int]map[int]bool{
	EntryStatusPending: {
		EntryStatusCompleted: true, // handleLastStep 处理最后一步
		EntryStatusRejected:  true, // handleRejectEntry, UnPass, timeout 处理驳回入口/驳回/超时
		EntryStatusRevoked:   true, // Revoke 撤回
	},
	EntryStatusRejected: {
		EntryStatusPending: true, // Resend 重新发起
	},
	EntryStatusRevoked: {
		EntryStatusPending: true, // Resend 重新发起
	},
}

// IsValidEntryTransition checks whether a state transition is allowed for Entry.
// IsValidEntryTransition 检查流程实例的状态转换是否允许。
func IsValidEntryTransition(from, to int) bool {
	targets, ok := validEntryTransitions[from]
	if !ok {
		return false
	}
	return targets[to]
}

// ValidateEntryTransition returns nil if the transition is allowed, or an error.
// ValidateEntryTransition 如果转换允许则返回 nil，否则返回错误。
func ValidateEntryTransition(from, to int) error {
	if IsValidEntryTransition(from, to) {
		return nil
	}
	return fmt.Errorf("invalid entry status transition: %d -> %d", from, to)
}

// =============================================================================
// Proc State Machine 审批任务状态机
// =============================================================================
//
// Valid transitions 有效状态转换:
//
//	                          +-- Revoke 撤回 ----------------------------+
//	                          v                                           |
//	Pending(0) ──┬──> Approved(1) 已通过   [Pass/Transfer 通过/转交]      |
//	             ├──> Rejected(-1) 已驳回  [UnPass, timeout 驳回/超时]    |
//	             ├──> Revoked(-2) 已撤回   [Revoke 撤回]                  |
//	             ├──> Transferred(3) 已转交 [TransferProc 转交任务]       |
//	             ├──> Skipped(4) 已跳过    [or-sign skip by colleague 或签同事跳过] |
//	             └──> Consensus(9) 会签通过 [auto-approve first/end step 自动通过首步/末步] |
//
//	Note: Consensus(9) is not "restartable" once set.
//	注意: Consensus(9) 一旦设置不可重新启动。

// validProcTransitions defines all permitted state-to-state moves for Proc.
// validProcTransitions 定义审批任务所有允许的状态到状态转换。
var validProcTransitions = map[int]map[int]bool{
	ProcStatusPending: {
		ProcStatusApproved:    true, // Pass/Transfer 通过/转交
		ProcStatusRejected:    true, // UnPass, timeout 驳回/超时
		ProcStatusRevoked:     true, // Revoke 撤回
		ProcStatusTransferred: true, // TransferProc 转交任务
		ProcStatusSkipped:     true, // or-sign skip 或签跳过
		ProcStatusConsensus:   true, // auto-approve (first step, end node) 自动通过（首步/末步）
	},
}

// IsValidProcTransition checks whether a state transition is allowed for Proc.
// IsValidProcTransition 检查审批任务的状态转换是否允许。
func IsValidProcTransition(from, to int) bool {
	targets, ok := validProcTransitions[from]
	if !ok {
		return false
	}
	return targets[to]
}

// ValidateProcTransition returns nil if the transition is allowed, or an error.
// ValidateProcTransition 如果转换允许则返回 nil，否则返回错误。
func ValidateProcTransition(from, to int) error {
	if IsValidProcTransition(from, to) {
		return nil
	}
	return fmt.Errorf("invalid proc status transition: %d -> %d", from, to)
}

// =============================================================================
// Terminal state checks 终态检查
// =============================================================================

// IsEntryActive returns true if the entry is still in progress and can be acted upon.
// IsEntryActive 如果流程实例仍在进行中且可被操作则返回 true。
func IsEntryActive(status int) bool {
	return status == EntryStatusPending
}

// IsEntryFinished returns true if the entry has reached a terminal state.
// IsEntryFinished 如果流程实例已达到终态则返回 true。
func IsEntryFinished(status int) bool {
	return status == EntryStatusCompleted || status == EntryStatusRejected || status == EntryStatusRevoked
}

// IsProcPending returns true if the proc is waiting for action.
// IsProcPending 如果审批任务正在等待处理则返回 true。
func IsProcPending(status int) bool {
	return status == ProcStatusPending
}

// IsProcFinished returns true if the proc has been decided upon.
// IsProcFinished 如果审批任务已有明确结果则返回 true。
func IsProcFinished(status int) bool {
	return status == ProcStatusApproved || status == ProcStatusRejected ||
		status == ProcStatusTransferred || status == ProcStatusSkipped || status == ProcStatusConsensus
}

// =============================================================================
// String representations for debugging/logging 状态字符串表示（用于调试/日志）
// =============================================================================

// EntryStatusString returns a human-readable string for an Entry status code.
// EntryStatusString 返回流程实例状态码的可读字符串。
func EntryStatusString(status int) string {
	switch status {
	case EntryStatusPending:
		return "Pending"
	case EntryStatusCompleted:
		return "Completed"
	case EntryStatusRejected:
		return "Rejected"
	case EntryStatusRevoked:
		return "Revoked"
	default:
		return fmt.Sprintf("Unknown(%d)", status)
	}
}

// ProcStatusString returns a human-readable string for a Proc status code.
// ProcStatusString 返回审批任务状态码的可读字符串。
func ProcStatusString(status int) string {
	switch status {
	case ProcStatusPending:
		return "Pending"
	case ProcStatusApproved:
		return "Approved"
	case ProcStatusRejected:
		return "Rejected"
	case ProcStatusRevoked:
		return "Revoked"
	case ProcStatusTransferred:
		return "Transferred"
	case ProcStatusSkipped:
		return "Skipped"
	case ProcStatusConsensus:
		return "Consensus(Auto)"
	default:
		return fmt.Sprintf("Unknown(%d)", status)
	}
}
