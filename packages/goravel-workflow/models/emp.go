package models

import (
	"fmt"
	"github.com/goravel/framework/database/orm"
)

// Emp 员工模型，对应 workflow 审批系统中的员工/用户表
// Emp represents an employee/user in the workflow approval system
type Emp struct {
	orm.Model
	// Name 员工姓名
	// Name is the employee's display name
	Name string `gorm:"column:name;not null" json:"name" form:"name"`
	// Email 员工邮箱地址，全局唯一
	// Email is the employee's email address, globally unique
	Email string `gorm:"column:email;not null;unique_index:users_email_unique" json:"email" form:"email"`
	// Password 员工登录密码，JSON 序列化时排除 // Exclude password from JSON response
	// Password is the employee login password, excluded from JSON serialization
	Password string `gorm:"column:password;not null" json:"password" form:"password"`
	// WorkNo 工号，全局唯一
	// WorkNo is the employee's work number, globally unique
	WorkNo string `gorm:"column:workno;not null;unique_index:users_workno_unique" json:"workno" form:"workno"`
	// DeptID 所属部门 ID
	// DeptID is the ID of the department this employee belongs to
	DeptID int `gorm:"column:dept_id;not null;default:0" json:"dept_id" form:"dept_id"`
	// Leave 离职状态：0=在职，1=已离职
	// Leave indicates employment status: 0=active, 1=left
	Leave int `gorm:"column:leave;not null;default:0" json:"leave" form:"leave"`
	// UserID 关联的系统用户 ID
	// UserID is the associated system user ID
	UserID uint `gorm:"column:user_id;" json:"user_id" form:"user_id"`
	// Dept 关联的部门信息（用于查询部门主管/经理等）
	// Dept is the associated department, used to query director/manager info
	Dept Dept `json:"Dept"`
}

// Passhook 审批通过时的钩子方法，默认实现仅打印日志。
// 宿主应用可在自己的 Emp/User 模型中覆盖此方法以实现自定义逻辑。
// Passhook is a hook method called when an approval passes; default implementation only prints a log.
// Host applications can override this method on their own Emp/User model for custom logic.
func (e *Emp) Passhook() {
	fmt.Println("Emp Passhook called.")
}

// UnPasshook 审批驳回时的钩子方法，默认实现仅打印日志。
// 宿主应用可在自己的 Emp/User 模型中覆盖此方法以实现自定义逻辑。
// UnPasshook is a hook method called when an approval is rejected; default implementation only prints a log.
// Host applications can override this method on their own Emp/User model for custom logic.
func (e *Emp) UnPasshook() {
	fmt.Println("Emp UnPasshook called.")
}

// Register 返回模型注册名称，供 workflow 引擎通过反射查找对应的模型
// Register returns the model registration name used by the workflow engine to locate this model via reflection
func (u *Emp) Register() string {
	return "Emp"
}

// Action 返回一个基于字符串指令执行操作的函数，当前无具体实现
// Action returns a function that executes an operation based on a string command; currently no implementation
func (u *Emp) Action() func(string) error {
	return nil
}
