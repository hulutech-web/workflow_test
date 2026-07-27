package workflow

// Hook 定义工作流通知钩子接口，由宿主应用的用户模型实现。
// Hook defines the workflow notification hook interface, implemented by the host application's user model.
type Hook interface {
	// NotifySendOne 通知流程发起人（当流程被驳回或完成时调用）。
	// NotifySendOne notifies the workflow initiator (called when the entry is rejected or completed).
	NotifySendOne(id uint) error

	// NotifyNextAuditor 通知下一审批人（当审批流转到新节点时调用）。
	// NotifyNextAuditor notifies the next approver (called when the workflow advances to a new step).
	NotifyNextAuditor(id uint) error
}
