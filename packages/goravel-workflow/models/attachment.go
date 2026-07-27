package models

import (
	"github.com/goravel/framework/database/orm"
)

// Attachment represents a file uploaded to a workflow entry 附件模型，用于存储上传的文件信息
type Attachment struct {
	orm.Model
	// UserID is the ID of the user who uploaded the file 上传文件的用户ID
	UserID uint `gorm:"column:user_id;type:int(11)" form:"user_id" json:"user_id"`
	// Name is the original name of the uploaded file 文件的原始名称
	Name string `gorm:"column:name;type:varchar(255);not null" json:"name" form:"name"`
	// Path is the storage path of the file on the server 文件在服务器上的存储路径
	Path string `gorm:"column:path;type:varchar(255);not null" json:"path" form:"path"`
	// Ext is the file extension (e.g., .pdf, .docx) 文件扩展名（如 .pdf、.docx）
	Ext string `gorm:"column:ext;type:varchar(255);not null" json:"ext" form:"ext"`
}
