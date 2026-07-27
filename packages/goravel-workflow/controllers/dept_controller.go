package controllers

import (
	"goravel/packages/goravel-workflow/models"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	httpfacades "github.com/hulutech-web/http_result"
)

// DeptController 部门控制器，负责处理部门相关的 HTTP 请求。
// 提供部门的 CRUD 操作，包括部门列表、递归展示、绑定主管及绑定经理等功能。
type DeptController struct {
	//Dependent services / 依赖的服务
}

// NewDeptController 创建并返回一个新的 DeptController 实例。
// 在该方法中可注入后续所需的服务依赖。
func NewDeptController() *DeptController {
	return &DeptController{
		//Inject services / 注入服务
	}
}

// Index 获取部门列表并以树形结构返回。
// 通过事务查询所有部门，预加载 Dirctor（主管）和 Manager（经理）关联，
// 然后调用 Recursion 方法将扁平数据转换为带有层级前缀的树形结构。
func (r *DeptController) Index(ctx http.Context) http.Response {
	depts := []models.Dept{}
	// 开启数据库事务，保证查询的原子性
	tx, _ := facades.Orm().Query().Begin()
	// 查询所有部门，同时预加载主管和经理信息
	tx.Model(&models.Dept{}).With("Director").With("Manager").Find(&depts)
	// 提交事务
	tx.Commit()
	deptInstance := models.Dept{}

	// 将扁平部门列表递归转换为树形结构，前缀为 "|---"，从根节点开始（父级ID=0，层级=0）
	result := deptInstance.Recursion(depts, "|---", 0, 0)
	return httpfacades.NewResult(ctx).Success("", result)
}

// List 获取部门的扁平列表（不进行递归转换）。
// 直接查询所有部门记录并以平铺形式返回，适用于下拉选择等场景。
func (r *DeptController) List(ctx http.Context) http.Response {
	depts := []models.Dept{}
	// 查询所有部门，不预加载关联关系，仅返回基础字段
	facades.Orm().Query().Model(&models.Dept{}).Find(&depts)
	return httpfacades.NewResult(ctx).Success("", depts)
}

// Show 查看单个部门的详细信息。
// 当前为占位方法，待后续实现。
func (r *DeptController) Show(ctx http.Context) http.Response {
	return nil
}

// Store 新增一个部门。
// 当前为占位方法，待后续实现。
func (r *DeptController) Store(ctx http.Context) http.Response {
	return nil
}

// Update 更新一个部门的信息。
// 当前为占位方法，待后续实现。
func (r *DeptController) Update(ctx http.Context) http.Response {
	return nil
}

// Destroy 删除一个部门。
// 当前为占位方法，待后续实现。
func (r *DeptController) Destroy(ctx http.Context) http.Response {
	return nil
}

// BindManager 为指定部门绑定经理（Manager）。
// 从请求参数中获取 manager_id 和 dept_id，然后将部门的 manager_id 字段更新为指定值。
func (r *DeptController) BindManager(ctx http.Context) http.Response {
	// 从请求中获取经理ID和部门ID
	manager_id := ctx.Request().InputInt("manager_id")
	dept_id := ctx.Request().InputInt("dept_id")
	// 更新部门的经理关联
	facades.Orm().Query().Model(&models.Dept{}).Where("id = ?", dept_id).Update("manager_id", manager_id)
	return httpfacades.NewResult(ctx).Success("设置成功", nil)
}

// BindDirector 为指定部门绑定主管（Director）。
// 从请求参数中获取 director_id 和 dept_id，然后将部门的 director_id 字段更新为指定值。
func (r *DeptController) BindDirector(ctx http.Context) http.Response {
	// 从请求中获取主管ID和部门ID
	director_id := ctx.Request().InputInt("director_id")
	dept_id := ctx.Request().InputInt("dept_id")
	// 更新部门的主管关联
	facades.Orm().Query().Model(&models.Dept{}).Where("id = ?", dept_id).Update("director_id", director_id)
	return httpfacades.NewResult(ctx).Success("设置成功", nil)
}
