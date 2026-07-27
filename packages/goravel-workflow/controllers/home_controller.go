package controllers

import (
	"goravel/packages/goravel-workflow/models"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	httpfacades "github.com/hulutech-web/http_result"
)

// HomeController 首页控制器，负责聚合展示首页仪表盘所需的所有数据
type HomeController struct {
	//Dependent services / 依赖的服务
}

// NewHomeController 创建 HomeController 实例的工厂函数
func NewHomeController() *HomeController {
	return &HomeController{
		//Inject services / 注入服务
	}
}

// Index 首页仪表盘接口，返回统计数据、我的申请、待办任务、工作流、已处理任务和抄送记录
func (r *HomeController) Index(ctx http.Context) http.Response {
	// 初始化所需的数据切片和用户模型
	entries := []models.Entry{}
	user := models.Emp{}
	facades.Auth(ctx).User(&user)
	emp := models.Emp{}
	facades.Orm().Query().Model(&models.Emp{}).Where("user_id=?", user.ID).Find(&emp)
	query := facades.Orm().Query()

	// 统计数据（Dept、Emp、Flow、Template、Entry 在 workflow 包中；User 由宿主应用提供，跳过）
	var deptCount, empCount, flowCount, templateCount, entryCount, pendingCount int64
	deptCount, _ = query.Model(&models.Dept{}).Count()
	empCount, _ = query.Model(&models.Emp{}).Count()
	flowCount, _ = query.Model(&models.Flow{}).Count()
	templateCount, _ = query.Model(&models.Template{}).Count()
	entryCount, _ = query.Model(&models.Entry{}).Count()
	pendingCount, _ = query.Model(&models.Proc{}).Where("status=?", models.ProcStatusPending).Count()

	//我的申请：查询当前用户发起的所有工作流申请，附带关联的员工、流程、最新审批步骤和工作流步骤信息
	query.Model(&models.Entry{}).With("Emp").With("Flow").With("Procs", func(q orm.Query) orm.Query {
		return q.Order("id desc").Limit(1)
	}).With("Process").Where("emp_id=?", user.ID).Where("pid=?", 0).Order("id desc").Find(&entries)

	//我的代办：查询分配给当前用户且状态为待处理的所有审批任务
	procs := []models.Proc{}
	query.Model(&models.Proc{}).With("Emp").With("Entry", func(query orm.Query) orm.Query {
		return query.With("Emp")
	}).Where("emp_id=?", user.ID).Where("status=?", models.ProcStatusPending).
		Order("id desc").Find(&procs)

	//工作流：查询已发布且可显示的工作流列表，供用户发起新申请
	flows := []models.Flow{}
	query.Model(&models.Flow{}).Where("is_publish=?", 1).Where("is_show=?", 1).Find(&flows)

	//待处理：查询当前用户已处理过（非待审状态）的审批任务，用于历史记录展示
	handle_procs := []models.Proc{}
	facades.Orm().Query().Model(&models.Proc{}).With("Emp").With("Entry").
		Where("emp_id=?", emp.ID).
		Where("status!=?", models.ProcStatusPending).
		Order("id desc").Find(&handle_procs)
	// 以下为已注释的备选查询方式：按 entry_id 分组去重
	//query.Model(&models.Proc{}).With("Emp").With("Entry", func(query orm.Query) {
	//	query.With("Emp")
	//}).Where("emp_id=?", user.ID).Where("status !=?", 0).Order("entry_id desc").
	//	Order("id asc").Group("entry_id").Find(&handle_procs)

	//我的抄送：查询抄送给当前用户的所有记录，附带关联的申请详情
	ccRecords := []models.CcRecord{}
	facades.Orm().Query().Model(&models.CcRecord{}).
		With("Entry", func(query orm.Query) orm.Query {
			return query.With("Emp").With("Flow").With("Process")
		}).
		Where("emp_id=?", emp.ID).Order("id desc").Find(&ccRecords)

	// 返回聚合后的首页数据，包含统计信息、我的申请、待办任务、工作流、已处理任务和抄送记录
	return httpfacades.NewResult(ctx).Success("", map[string]interface{}{
		"stats": map[string]int64{
			"dept_count":     deptCount,
			"emp_count":      empCount,
			"flow_count":     flowCount,
			"template_count": templateCount,
			"entry_count":    entryCount,
			"pending_count":  pendingCount,
		},
		"entries":      entries,
		"procs":        procs,
		"flows":        flows,
		"handle_procs": handle_procs,
		"cc_records":   ccRecords,
	})
}
