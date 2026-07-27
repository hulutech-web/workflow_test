package controllers

import (
	// 数据模型：Employee、CcRecord（抄送记录）等
	"goravel/packages/goravel-workflow/models"
	"strings"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	httpfacades "github.com/hulutech-web/http_result"
	"github.com/spf13/cast"
)

// CcController 抄送控制器，处理审批流程中的抄送相关操作
type CcController struct{}

// NewCcController 创建抄送控制器实例
func NewCcController() *CcController {
	return &CcController{}
}

// Index 获取当前用户的抄送列表 / Get the current user's CC list
func (r *CcController) Index(ctx http.Context) http.Response {
	// 从 JWT 令牌中解析当前登录用户
	var emp models.Emp
	facades.Auth(ctx).User(&emp)

	// 查询该用户的所有抄送记录，支持分页与搜索参数
	var ccRecords []models.CcRecord
	queries := ctx.Request().Queries()
	res, _ := httpfacades.NewResult(ctx).SearchByParams(queries, map[string]interface{}{
		"emp_id": emp.ID,
	}).ResultPagination(&ccRecords)

	return res
}

// GetEntryCC 获取某个审批流程实例的抄送记录 / Get CC records for a specific entry
func (r *CcController) GetEntryCC(ctx http.Context) http.Response {
	// 从路由参数中提取流程实例 ID
	entry_id := ctx.Request().RouteInt("entry_id")
	var ccRecords []models.CcRecord
	// 按 ID 升序排列，确保抄送记录按创建时间顺序展示
	facades.Orm().Query().Model(&models.CcRecord{}).Where("entry_id=?", entry_id).Order("id asc").Find(&ccRecords)
	return httpfacades.NewResult(ctx).Success("", ccRecords)
}

// Store 手动添加抄送人 / Manually add CC recipients to an entry
func (r *CcController) Store(ctx http.Context) http.Response {
	// 获取流程实例 ID 和抄送员工 ID 列表（逗号分隔）
	entry_id := ctx.Request().InputInt("entry_id")
	emp_ids := ctx.Request().Input("emp_ids")

	// 解析并清洗员工 ID 列表：去除空格，过滤无效值
	var empIDs []int
	for _, s := range strings.Split(emp_ids, ",") {
		if id := cast.ToInt(strings.TrimSpace(s)); id > 0 {
			empIDs = append(empIDs, id)
		}
	}
	// 校验：至少需要一个有效的抄送人
	if len(empIDs) == 0 {
		return httpfacades.NewResult(ctx).Error(400, "请选择抄送人", "")
	}

	// 校验：流程实例必须存在
	var entry models.Entry
	facades.Orm().Query().Model(&models.Entry{}).Where("id=?", entry_id).First(&entry)
	if entry.ID == 0 {
		return httpfacades.NewResult(ctx).Error(400, "流程不存在", "")
	}

	// 逐个为有效的员工创建抄送记录
	for _, empID := range empIDs {
		var emp models.Emp
		facades.Orm().Query().Model(&models.Emp{}).Where("id=?", empID).First(&emp)
		// 跳过不存在的员工
		if emp.ID == 0 {
			continue
		}
		// 创建抄送记录，默认状态为 0（未读）
		facades.Orm().Query().Model(&models.CcRecord{}).Create(&models.CcRecord{
			EntryID:   entry.ID,
			FlowID:    entry.FlowID,
			ProcessID: entry.ProcessID,
			EmpID:     empID,
			EmpName:   emp.Name,
			Status:    0,
		})
	}
	return httpfacades.NewResult(ctx).Success("抄送成功", nil)
}
