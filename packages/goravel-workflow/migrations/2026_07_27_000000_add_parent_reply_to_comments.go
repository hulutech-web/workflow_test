package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260727000000AddParentReplyToComments struct {
}

func (r *M20260727000000AddParentReplyToComments) Signature() string {
	return "2026_07_27_000000_add_parent_reply_to_comments"
}

func (r *M20260727000000AddParentReplyToComments) Up() error {
	if err := facades.Schema().Table("proc_comments", func(table schema.Blueprint) {
		if !facades.Schema().HasColumn("proc_comments", "parent_id") {
			table.BigInteger("parent_id").Default(0).Comment("父评论ID，0表示顶级评论")
		}
		if !facades.Schema().HasColumn("proc_comments", "reply_to_emp_id") {
			table.BigInteger("reply_to_emp_id").Default(0).Comment("被回复的员工ID")
		}
		if !facades.Schema().HasColumn("proc_comments", "reply_to_emp_name") {
			table.String("reply_to_emp_name").Default("").Comment("被回复的员工名称")
		}
	}); err != nil {
		return err
	}
	return nil
}

func (r *M20260727000000AddParentReplyToComments) Down() error {
	if err := facades.Schema().Table("proc_comments", func(table schema.Blueprint) {
		if facades.Schema().HasColumn("proc_comments", "parent_id") {
			table.DropColumn("parent_id")
		}
		if facades.Schema().HasColumn("proc_comments", "reply_to_emp_id") {
			table.DropColumn("reply_to_emp_id")
		}
		if facades.Schema().HasColumn("proc_comments", "reply_to_emp_name") {
			table.DropColumn("reply_to_emp_name")
		}
	}); err != nil {
		return err
	}
	return nil
}
