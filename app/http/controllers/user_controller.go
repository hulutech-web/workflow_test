package controllers

import (
	"goravel/app/http/requests"
	"goravel/app/models"

	"github.com/spf13/cast"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	httpfacade "github.com/hulutech-web/http_result"
)

type UserController struct {
	//Dependent services
}

func NewUserController() *UserController {
	return &UserController{
		//Inject services
	}
}

// Index 分页查询，支持搜索，路由参数?name=xxx&pageSize=1&currentPage=1&sort=xxx&order=xxx,等其他任意的查询参数
// @Summary      分页查询
// @Description  分页查询
// @Tags         UserController
// @Accept       json
// @Produce      json
// @Id UserIndex
// @Security ApiKeyAuth
// @Param Authorization header string false "Bearer 用户令牌"
// @Param  name  query  string  false  "name"
// @Param  pageSize  query  string  false  "pageSize"
// @Param  currentPage  query  string  false  "currentPage"
// @Param  sort  query  string  false  "sort"
// @Param  order  query  string  false  "order"
// @Success 200 {string} json {}
// @Router       /api/user [get]
func (r *UserController) Index(ctx http.Context) http.Response {
	users := []models.User{}
	queries := ctx.Request().Queries()
	res, err := httpfacade.NewResult(ctx).SearchByParams(queries, nil).ResultPagination(&users, []httpfacade.WithConfig{{
		Relation: "Roles",
		Callback: nil,
	}})
	if err != nil {
		return httpfacade.NewResult(ctx).Error(http.StatusInternalServerError, "", err.Error())
	}
	return res
}

// List 列表查询
// @Summary      列表查询
// @Description  列表查询
// @Tags         UserController
// @Accept       json
// @Produce      json
// @Id UserList
// @Param  username  query  string  false  "username"
// @Security ApiKeyAuth
// @Param Authorization header string false "Bearer 用户令牌"
// @Success 200 {string} json {}
// @Router       /api/user/list [get]
func (r *UserController) List(ctx http.Context) http.Response {
	users := []models.User{}
	queries := ctx.Request().Queries()
	_, err := httpfacade.NewResult(ctx).SearchByParams(queries, map[string]interface{}{
		"identity": "用户",
	}).List(&users)
	if err != nil {
		return httpfacade.NewResult(ctx).Error(http.StatusInternalServerError, "", err.Error())
	}
	return httpfacade.NewResult(ctx).Success("", users)
}

// EmployeeList 员工列表查询
// @Summary      列表查询
// @Description  列表查询
// @Tags         UserController
// @Accept       json
// @Produce      json
// @Id UserEmployeeList
// @Param  username  query  string  false  "username"
// @Security ApiKeyAuth
// @Param Authorization header string false "Bearer 用户令牌"
// @Success 200 {string} json {}
// @Router       /api/user/employee_list [get]
func (r *UserController) EmployeeList(ctx http.Context) http.Response {
	users := []models.User{}
	queries := ctx.Request().Queries()
	_, err := httpfacade.NewResult(ctx).SearchByParams(queries, map[string]interface{}{
		"identity": "员工",
	}).List(&users)
	if err != nil {
		return httpfacade.NewResult(ctx).Error(http.StatusInternalServerError, "", err.Error())
	}
	return httpfacade.NewResult(ctx).Success("", users)
}
func (r *UserController) Show(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt("id")
	user := models.User{}
	facades.Orm().Query().Model(&models.User{}).Where("id = ?", id).First(&user)
	return httpfacade.NewResult(ctx).Success("", user)
}

// Store 新增
// @Summary      新增
// @Description  新增
// @Tags         UserController
// @Accept       json
// @Produce      json
// @Id UserStore
// @Security ApiKeyAuth
// @Param Authorization header string false "Bearer 用户令牌"
// @Param userData body requests.UserRequest true "用户数据"
// @Success 200 {string} json {}
// @Router       /api/user [post]
func (r *UserController) Store(ctx http.Context) http.Response {
	var userRequest requests.UserRequest
	errors, err := ctx.Request().ValidateRequest(&userRequest)
	if err != nil {
		return httpfacade.NewResult(ctx).Error(http.StatusInternalServerError, "数据错误", err.Error())
	}
	if errors != nil {
		return httpfacade.NewResult(ctx).ValidError("", errors.All())
	}
	user := models.User{}
	//todo add request values
	err1 := facades.Orm().Query().Model(&models.User{}).Create(&user)
	if err1 != nil {
		return httpfacade.NewResult(ctx).Error(http.StatusInternalServerError, "数据错误", err1.Error())
	}
	return httpfacade.NewResult(ctx).Success("创建成功", nil)
}

// Update
// @Summary      更新
// @Description  更新
// @Tags         UserController
// @Accept       json
// @Produce      json
// @Id UserUpdate
// @Security ApiKeyAuth
// @Param Authorization header string false "Bearer 用户令牌"
// @Param userData body requests.UserRequest true "用户数据"
// @Success 200 {string} json {}
// @Param id path string true "id"  // 关键：指定参数位置为 path
// @Router       /api/user/{id} [put]
func (r *UserController) Update(ctx http.Context) http.Response {
	id := ctx.Request().Route("id")
	user := models.User{}
	facades.Orm().Query().Model(&models.User{}).Where("id=?", id).Find(&user)
	var userRequest requests.UserRequest
	errors, err := ctx.Request().ValidateRequest(&userRequest)
	if err != nil {
		return httpfacade.NewResult(ctx).Error(http.StatusInternalServerError, "数据错误", err.Error())
	}
	if errors != nil {
		return httpfacade.NewResult(ctx).ValidError("", errors.All())
	}
	//todo add request values
	pwd := ""
	if userRequest.Password == user.Password {
		// 如果密码一直，不做改动
		pwd = user.Password
	} else {
		pwd, _ = facades.Hash().Make(userRequest.Password)
	}
	user = models.User{
		Realname: userRequest.Realname,
		Phone:    userRequest.Phone,
		Username: userRequest.Username,
		Password: pwd,
		Sex:      userRequest.Sex,
		Status:   userRequest.Status,
		Remark:   userRequest.Remark,
		Avatar:   userRequest.Avatar,
	}
	user.ID = cast.ToUint(id)
	err1 := facades.Orm().Query().Model(&models.User{}).Where("id=?", id).Save(&user)
	if err1 != nil {
		return httpfacade.NewResult(ctx).Error(http.StatusInternalServerError, "数据错误,请联系管理员", err1.Error())
	}
	var role models.Role
	facades.Orm().Query().Model(&models.Role{}).Where("id=?", userRequest.RoleID).Find(&role)
	err = user.SyncRole(&role)
	if err != nil {
		return httpfacade.NewResult(ctx).Error(http.StatusInternalServerError, "数据错误", err.Error())
	}
	return httpfacade.NewResult(ctx).Success("修改成功", nil)
}

// Destroy 删除
// @Summary      删除
// @Description  删除
// @Tags         UserController
// @Accept       json
// @Produce      json
// @Id UserDestroy
// @Security ApiKeyAuth
// @Success 200 {string} json {}
// @Param id path string true "id"  // 关键：指定参数位置为 path
// @Router       /api/user/{id} [delete]
func (r *UserController) Destroy(ctx http.Context) http.Response {
	id := ctx.Request().Route("id")
	facades.Orm().Query().Model(&models.User{}).Where("id=?", id).Delete(&models.User{})
	return httpfacade.NewResult(ctx).Success("删除成功", nil)
}

// Option 选项
// @Summary      选项
// @Description  选项
// @Tags         UserController
// @Accept       json
// @Produce      json
// @Id UserOption
// @Security ApiKeyAuth
// @Param  username  query  string  false  "username"
// @Success 200 {string} json {}
// @Router       /api/user/option [get]
func (r *UserController) Option(ctx http.Context) http.Response {
	type Opt struct {
		Value uint   `json:"value"`
		Label string `json:"label"`
	}
	username := ctx.Request().Query("username")

	var opts []Opt
	if username == "" {
		facades.Orm().Query().Model(&models.User{}).Select("id as value,username as label").Scan(&opts)
	} else {
		facades.Orm().Query().Model(&models.User{}).Where("username like ?", "%"+username+"%").Select("id as value,username as label").Scan(&opts)
	}

	return httpfacade.NewResult(ctx).Success("", opts)
}

// EmployeeOption 选项
// @Summary      选项
// @Description  选项
// @Tags         UserController
// @Accept       json
// @Produce      json
// @Id UserEmployeeOption
// @Security ApiKeyAuth
// @Param  username  query  string  false  "username"
// @Success 200 {string} json {}
// @Router       /api/user/employee_option [get]
func (r *UserController) EmployeeOption(ctx http.Context) http.Response {
	type Opt struct {
		Value uint   `json:"value"`
		Label string `json:"label"`
		Phone string `json:"phone"`
	}
	username := ctx.Request().Query("username")

	var opts []Opt
	if username == "" {
		facades.Orm().Query().Model(&models.User{}).Select("id as value,realname as label,phone").
			Where("status=?", "正常").
			Where("identity=?", "员工").Scan(&opts)
	} else {
		facades.Orm().Query().Model(&models.User{}).Where("realname like ?", "%"+username+"%").
			Where("status=?", "正常").
			Where("identity=?", "员工").
			Select("id as value,realname as label,phone").Scan(&opts)
	}

	return httpfacade.NewResult(ctx).Success("", opts)
}

// MiniInfo 当前登录人信息
// @Summary      选项
// @Description  选项
// @Tags         UserController
// @Accept       json
// @Produce      json
// @Id UserMiniInfo
// @Security ApiKeyAuth
// @Success 200 {string} json {}
// @Router       /api/mini/user/mini_info [get]
func (r *UserController) MiniInfo(ctx http.Context) http.Response {
	user := models.User{}
	err2 := facades.Auth(ctx).User(&user)
	if err2 != nil {
		return httpfacade.NewResult(ctx).Error(http.StatusInternalServerError, "用户不存在", err2.Error())
	}
	err := facades.Orm().Query().Model(&models.User{}).Where("id=?", user.ID).Find(&user)
	if err != nil {
		return httpfacade.NewResult(ctx).Error(http.StatusInternalServerError, "用户不存在", err.Error())
	}
	return httpfacade.NewResult(ctx).Success("", user)
}
