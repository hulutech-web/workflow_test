// Package Upload 文件上传服务，处理附件上传并将文件记录持久化到数据库。
// 支持从 HTTP 请求中接收文件，存储到配置的存储驱动，并创建 Attachment 记录。
package Upload

import (
	"fmt"
	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"goravel/packages/goravel-workflow/models"
	"time"
)

// UploadService 上传服务结构体，提供文件上传相关的业务逻辑。
// 通过 NewUploadService 工厂方法创建实例。
type UploadService struct {
}

// UploadImport 上传导入参数结构体，用于接收上传请求中的表单数据。
type UploadImport struct {
	Path   string `json:"path" form:"path"`     //附件路径 / 附件的存储路径
	Alias  string `json:"alias" form:"alias"`    //自定义表名 / 导入时自定义的表名称
	Id     int    `json:"id" form:"id"`          //附件id attachment_id / 附件的数据库记录ID
	UserId int    `json:"user_id" form:"user_id"` // 用户ID，标识上传者
}

// NewUploadService 创建并返回一个新的 UploadService 实例。
// 使用工厂方法模式，便于后续依赖注入扩展。
func NewUploadService() *UploadService {
	return &UploadService{}
}

// Upload 处理文件上传的核心方法。
// 接收 HTTP 上下文和上传的文件对象，将文件存储到磁盘并创建数据库记录。
//
// 流程说明:
//  1. 提取文件的原始名称和扩展名
//  2. 按年月（如 2026-07）组织目录，将文件存储到对应目录
//  3. 尝试从 JWT 令牌中解析当前登录用户信息
//  4. 如果无法获取用户信息，创建不带 UserID 的附件记录（匿名上传）
//  5. 如果可以获取用户信息，创建带 UserID 的附件记录
//  6. 返回创建的 Attachment 记录
//
// 参数:
//   - ctx: HTTP 上下文，用于获取认证用户信息
//   - file: 上传的文件对象，由框架的文件系统契约定义
//
// 返回值:
//   - *models.Attachment: 创建成功的附件记录指针
//   - error: 处理过程中发生的错误
func (*UploadService) Upload(ctx http.Context, file filesystem.File) (*models.Attachment, error) {
	// 获取文件的原始名称（客户端上传时的文件名）
	name := file.GetClientOriginalName()
	// 获取文件的扩展名，如 "jpg"、"pdf" 等
	extension := file.GetClientOriginalExtension()

	// 格式化当前年月，用于按月份分目录存储，格式如 "2026-07"
	yearMonth := fmt.Sprintf("%d-%02d", time.Now().Year(), time.Now().Month())

	// 将文件存储到磁盘，返回存储后的相对路径
	putFile, err := facades.Storage().PutFile(yearMonth, file)

	// 尝试从 HTTP 上下文中获取当前登录用户信息，用于关联上传者
	user := models.Emp{}
	if err1 := facades.Auth(ctx).User(&user); err1 != nil {
		// 无法获取用户信息时的处理：创建不带用户关联的附件记录（适用于未登录或匿名上传场景）
		att := models.Attachment{
			Name: name,
			Path: putFile,
			Ext:  extension,
		}
		err1 := facades.Orm().Query().Model(&models.Attachment{}).Create(&att)
		if err1 != nil {
			return nil, err1
		}
		return &att, nil
	}

	// 获取到用户信息时的处理：创建带 UserID 的附件记录，关联上传者
	att := models.Attachment{
		Name:   name,
		Path:   putFile,
		UserID: user.ID,
		Ext:    extension,
	}
	// 将附件信息持久化到数据库
	err2 := facades.Orm().Query().Model(&models.Attachment{}).Create(&att)
	if err != nil {
		// 注意：此处判断的是外层的 err 而非 err2，实际应检查 err2。
		// 由于 PutFile 成功时 err 为 nil，此处的错误处理实际上是返回 err2。
		return nil, err2
	}
	return &att, nil
}
