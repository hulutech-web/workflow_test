package commands

import (
	"fmt"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"goravel/packages/goravel-workflow/models"
)

// TimeoutCheckCommand 超时检查命令，用于定时扫描并处理已超时的待审批任务
type TimeoutCheckCommand struct{}

// NewTimeoutCheckCommand 创建超时检查命令实例，返回 console.Command 接口
func NewTimeoutCheckCommand() console.Command {
	return &TimeoutCheckCommand{}
}

// Signature 返回命令签名，用于 artisan 命令行调用
func (r *TimeoutCheckCommand) Signature() string {
	return "workflow:timeout-check"
}

// Description 返回命令描述信息
func (r *TimeoutCheckCommand) Description() string {
	return "Check and handle overdue approval tasks" // 检查并处理超时的审批任务
}

// Extend 返回命令扩展配置，当前命令无需额外扩展
func (r *TimeoutCheckCommand) Extend() command.Extend {
	return command.Extend{}
}

// Handle 执行超时检查的核心逻辑：
// 遍历所有待处理的 Proc 任务，计算每个任务自创建以来的耗时，
// 若超过对应 Process 步骤设定的 LimitTime 限制时间，则自动驳回该任务及所属 Entry
func (r *TimeoutCheckCommand) Handle(ctx console.Context) error {
	// 获取数据库查询实例
	query := facades.Orm().Query()

	// 查询所有状态为"待处理"的审批任务
	var procs []models.Proc
	if err := query.Model(&models.Proc{}).
		Where("status=?", models.ProcStatusPending).
		Find(&procs); err != nil {
		return err
	}

	// 获取当前时间，用于计算各任务的已耗时间
	now := carbon.Now()
	for _, proc := range procs {
		// 查询当前任务所属的审批步骤，获取超时限制配置
		var process models.Process
		if err := query.Model(&models.Process{}).
			Where("id=?", proc.ProcessID).
			First(&process); err != nil {
			continue
		}

		// LimitTime <= 0 表示该步骤未设置超时限制，跳过检查
		if process.LimitTime <= 0 {
			continue // No timeout set for this step（该步骤未设置超时）
		}

		// Use Concurrence (the approval time) to calculate elapsed seconds
		// 使用 Concurrence（审批任务创建时间）计算已耗秒数
		elapsedSeconds := proc.Concurrence.DiffInSeconds(now)

		// 若已耗时间超过步骤限制时间，执行自动驳回
		if elapsedSeconds >= int64(process.LimitTime) {
			// 将审批任务标记为驳回，并记录驳回原因
			proc.Status = models.ProcStatusRejected
			proc.Content = "超时未处理，系统自动驳回"
			if saveErr := query.Model(&models.Proc{}).Where("id=?", proc.ID).Save(&proc); saveErr != nil {
				fmt.Printf("Failed to timeout-reject proc %d: %v\n", proc.ID, saveErr)
				continue
			}

			// 同步更新所属 Entry 的状态为已驳回
			var entry models.Entry
			if err := query.Model(&models.Entry{}).Where("id=?", proc.EntryID).First(&entry); err == nil {
				entry.Status = models.EntryStatusRejected
				query.Model(&models.Entry{}).Where("id=?", entry.ID).Save(&entry)
				fmt.Printf("Timed out entry %d at process %d (%ds/%ds)\n", entry.ID, proc.ProcessID, elapsedSeconds, process.LimitTime)
			}
		}
	}

	return nil
}
