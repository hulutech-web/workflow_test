package models

import "fmt"

// =============================================================================
// Entry State Machine
// =============================================================================
//
// Valid transitions:
//
//	                    +-- Resend ----+
//	                    v              |
//	Pending(0) ──┬──> Completed(9)    |
//	             ├──> Rejected(-1) ──>+
//	             └──> Revoked(-2) ───>+
//
//	Child workflows: when a child completes with ChildAfter=2,
//	the parent Entry stays at Pending(0) and its ProcessID advances.

type EntryTransition struct {
	From int
	To   int
}

var validEntryTransitions = map[int]map[int]bool{
	EntryStatusPending: {
		EntryStatusCompleted: true, // handleLastStep
		EntryStatusRejected:  true, // handleRejectEntry, UnPass, timeout
		EntryStatusRevoked:   true, // Revoke
	},
	EntryStatusRejected: {
		EntryStatusPending: true, // Resend
	},
	EntryStatusRevoked: {
		EntryStatusPending: true, // Resend
	},
}

func IsValidEntryTransition(from, to int) bool {
	targets, ok := validEntryTransitions[from]
	if !ok {
		return false
	}
	return targets[to]
}

func ValidateEntryTransition(from, to int) error {
	if IsValidEntryTransition(from, to) {
		return nil
	}
	return fmt.Errorf("invalid entry status transition: %d -> %d", from, to)
}

// =============================================================================
// Proc State Machine
// =============================================================================
//
// Valid transitions:
//
//	                          +-- Revoke -------------------------+
//	                          v                                   |
//	Pending(0) ──┬──> Approved(1)    [Pass/Transfer]              |
//	             ├──> Rejected(-1)   [UnPass, timeout]            |
//	             ├──> Revoked(-2)    [Revoke]                     |
//	             ├──> Transferred(3) [TransferProc]               |
//	             ├──> Skipped(4)     [or-sign skip by colleague]  |
//	             └──> Consensus(9)   [auto-approve first/end step]|
//
//	Note: Consensus(9) is not "restartable" once set.

var validProcTransitions = map[int]map[int]bool{
	ProcStatusPending: {
		ProcStatusApproved:    true, // Pass/Transfer
		ProcStatusRejected:    true, // UnPass, timeout
		ProcStatusRevoked:     true, // Revoke
		ProcStatusTransferred: true, // TransferProc
		ProcStatusSkipped:     true, // or-sign skip
		ProcStatusConsensus:   true, // auto-approve (first step, end node)
	},
}

func IsValidProcTransition(from, to int) bool {
	targets, ok := validProcTransitions[from]
	if !ok {
		return false
	}
	return targets[to]
}

func ValidateProcTransition(from, to int) error {
	if IsValidProcTransition(from, to) {
		return nil
	}
	return fmt.Errorf("invalid proc status transition: %d -> %d", from, to)
}

// =============================================================================
// Terminal state checks
// =============================================================================

// IsEntryActive returns true if the entry is still in progress and can be acted upon.
func IsEntryActive(status int) bool {
	return status == EntryStatusPending
}

// IsEntryFinished returns true if the entry has reached a terminal state.
func IsEntryFinished(status int) bool {
	return status == EntryStatusCompleted || status == EntryStatusRejected || status == EntryStatusRevoked
}

// IsProcPending returns true if the proc is waiting for action.
func IsProcPending(status int) bool {
	return status == ProcStatusPending
}

// IsProcFinished returns true if the proc has been decided upon.
func IsProcFinished(status int) bool {
	return status == ProcStatusApproved || status == ProcStatusRejected ||
		status == ProcStatusTransferred || status == ProcStatusSkipped || status == ProcStatusConsensus
}

// =============================================================================
// String representations for debugging/logging
// =============================================================================

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
