package controllers

import (
	"goravel/packages/goravel-workflow/models"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	httpfacades "github.com/hulutech-web/http_result"
	"github.com/spf13/cast"
)

// EmpController 员工控制器，提供员工的分页查询、搜索、选项列表及用户绑定等功能
type EmpController struct {
	//Dependent services // 依赖服务
}

// NewEmpController 创建员工控制器实例
func NewEmpController() *EmpController {
	return &EmpController{
		//Inject services // 注入服务
	}
}

// Index 员工列表分页查询，支持按查询参数过滤，并预加载关联的部门（Dept）信息
func (r *EmpController) Index(ctx http.Context) http.Response {
	emps := []models.Emp{}
	queries := ctx.Request().Queries()
	// 使用 SearchByParams 进行分页查询，通过 WithConfig 预加载 Dept 关联
	result, _ := httpfacades.NewResult(ctx).SearchByParams(queries, nil).ResultPagination(&emps, []httpfacades.WithConfig{
		{
			Relation: "Dept",
			Callback: nil,
		},
	})
	return result
}

// Show 查看单个员工详情（暂未实现）
func (r *EmpController) Show(ctx http.Context) http.Response {
	return nil
}

// Store 新增员工（暂未实现）
func (r *EmpController) Store(ctx http.Context) http.Response {
	return nil
}

// Update 更新员工信息（暂未实现）
func (r *EmpController) Update(ctx http.Context) http.Response {
	return nil
}

// Destroy 删除员工（暂未实现）
func (r *EmpController) Destroy(ctx http.Context) http.Response {
	return nil
}

// Search 员工搜索，根据姓名或工号模糊匹配，返回 label/value 格式的下拉选项
func (r *EmpController) Search(ctx http.Context) http.Response {
	// Opt 下拉选项结构体，用于前端选择器组件
	type Opt struct {
		Label string `json:"label"`
		Value string `json:"value"`
	}
	name := ctx.Request().Input("name", "")
	opts := []Opt{}
	// 按姓名或工号模糊匹配，返回 name 作为 label、id 作为 value
	facades.Orm().Query().Model(&models.Emp{}).Where("name like ?", "%"+name+"%").OrWhere("workno like ?", "%"+name+"%").Select("name as label,id as value").Scan(&opts)
	return httpfacades.NewResult(ctx).Success("", opts)
}

// Options 获取全部员工列表，用于下拉选择等场景
func (r *EmpController) Options(ctx http.Context) http.Response {
	emps := []models.Emp{}
	facades.Orm().Query().Model(&models.Emp{}).Find(&emps)
	return httpfacades.NewResult(ctx).Success("", emps)
}

// BindUser 员工绑定用户，将员工记录与系统登录用户关联
func (r *EmpController) BindUser(ctx http.Context) http.Response {
	emp_id := ctx.Request().InputInt("emp_id")
	user_id := ctx.Request().InputInt("user_id")
	// 将员工记录的 user_id 字段更新为指定的用户 ID
	facades.Orm().Query().Model(&models.Emp{}).Where("id=?", emp_id).Update("user_id", cast.ToUint(user_id))
	return httpfacades.NewResult(ctx).Success("绑定成功", nil)
}
