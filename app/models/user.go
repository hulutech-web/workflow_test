package models

import (
	"goravel/database/factories"
	"goravel/packages/goravel-workflow/services/workflow"

	"github.com/goravel/framework/contracts/database/factory"
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
)

type User struct {
	orm.Model
	Username string `gorm:"column:username" form:"username" json:"username"`
	Sex      string `gorm:"sex" form:"sex" json:"sex"`

	Phone string `gorm:"phone" form:"phone" json:"phone"`

	Openid       string           `gorm:"column:openid" form:"openid" json:"openid"`
	Unionid      string           `gorm:"column:unionid" form:"unionid" json:"unionid"`
	Password     string           `gorm:"password" form:"password"  json:"password"`
	Avatar       string           `gorm:"avatar" form:"avatar" json:"avatar"`
	Realname     string           `gorm:"realname" form:"realname" json:"realname"`
	IDCardNumber string           `gorm:"id_card_number" form:"id_card_number" json:"id_card_number"`
	Remark       string           `gorm:"remark" form:"remark" json:"remark"`
	LastLogin    *carbon.DateTime `gorm:"column:last_login" form:"last_login" json:"last_login"`
	Status       string           `gorm:"column:status;default:null" form:"status" json:"status"`
	Roles        []Role           `gorm:"many2many:user_roles;"`

	Workflow *workflow.Workflow
}

// 通知发起人，在被驳回时调用，或者整个流程结束时调用。
func (u *User) NotifySendOne(id uint) error {

	//logrus.WithFields(logrus.Fields{
	//	"id": id,
	//}).Info("NotifySendOne通知来咯")
	facades.Log().Infof("NotifySendOne通知来咯%d", id)

	return nil
}

// 通知下一个审批人，当当前环节的审批人通过时，触发。
func (u *User) NotifyNextAuditor(id uint) error {
	//logrus.WithFields(logrus.Fields{
	//	"id": id,
	//}).Info("NotifyNextAuditor通知来咯")
	facades.Log().Infof("NotifyNextAuditor通知来咯%d", id)
	return nil
}
func (u *User) Factory() factory.Factory {
	return &factories.UserFactory{}
}
func (u *User) IsAdmin() bool {
	return u.ID < 3
}

func (u *User) GetRoles() []string {
	roles := []string{}
	facades.Orm().Query().Model(u).With("Roles").Find(&u)
	for _, role := range u.Roles {
		roles = append(roles, role.Name)
	}
	return roles
}

// // GetPermissions 获取权限
func (u *User) GetPermissions() []string {
	permissions := []string{}
	uniquePermissions := make(map[string]struct{}) // 用于去重的 map
	facades.Orm().Query().Model(&u).With("Roles").Find(&u)
	for _, role := range u.Roles {
		facades.Orm().Query().With("Permissions").Find(&role)
		// 使用 map 检查权限是否存在，避免重复
		for _, permission := range role.Permissions {
			// 使用 map 检查权限是否存在，避免重复
			if _, exists := uniquePermissions[permission.Code]; !exists {
				uniquePermissions[permission.Code] = struct{}{}
				permissions = append(permissions, permission.Code)
			}
		}
	}
	return permissions
}

/*SyncRole 同步角色*/
func (u *User) SyncRole(role *Role) error {
	return facades.Orm().Query().Model(&u).With("Roles").Association("Roles").Replace(u.Roles, role)
}
