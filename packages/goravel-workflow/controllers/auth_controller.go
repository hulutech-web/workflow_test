package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/validation"
	"goravel/packages/goravel-workflow/models"
)

// AuthController 认证控制器，处理管理员登录、H5登录、用户登录和退出等认证相关操作
type AuthController struct {
}

// NewAuthController 创建认证控制器实例，用于注册路由时注入依赖
func NewAuthController() *AuthController {
	return &AuthController{
		//Inject services
		// 注入服务依赖（当前为空，预留扩展点）
	}
}

// AdminLogin 管理员登录接口，使用工号(workno)和密码进行身份验证
// 验证流程：参数校验 → 查询用户 → 密码验证 → 生成JWT Token
func (r *AuthController) AdminLogin(ctx http.Context) http.Response {
	var user models.Emp
	ctx.Request().Bind(&user)
	password := user.Password
	//验证
	// 参数验证：工号和密码均为必填项
	validator, err := facades.Validation().Make(ctx, map[string]any{
		"workno":   ctx.Request().Input("workno", ""),
		"password": ctx.Request().Input("password", ""),
	}, map[string]any{
		"workno":   "required",
		"password": "required",
	}, validation.Messages(map[string]string{
		"workno.required":   "工号不能为空",
		"password.required": "密码不能为空",
	}))
	if err != nil {
		return ctx.Response().Status(http.StatusInternalServerError).Json(http.Json{
			"message": "参数校验失败",
		})
	}
	if validator.Fails() {
		return ctx.Response().Status(http.StatusUnprocessableEntity).Json(http.Json{
			"errors": validator.Errors().All(),
		})
	}
	//手机号密码验证
	// 根据工号查询用户是否存在
	facades.Orm().Query().Model(&user).Where("workno", user.WorkNo).Find(&user)

	if user.ID == 0 {
		ctx.Request().AbortWithStatusJson(401, http.Json{
			"message": "error",
			"fail":    "用户不存在,请点击注册",
		})
		return nil
	}
	var user_exist models.Emp
	facades.Orm().Query().Model(&user_exist).Where("workno=?", user.WorkNo).Find(&user_exist)
	//解密
	// 解密（已注释的管理员权限校验，仅ID=1的管理员可登录）
	//if user_exist.ID != 1 {
	//	return ctx.Response().Status(http.StatusInternalServerError).Json(http.Json{
	//		"message": "无权登录",
	//	})
	//}

	// 哈希密码比对验证
	if !facades.Hash().Check(password, user_exist.Password) {
		return ctx.Response().Status(http.StatusInternalServerError).Json(http.Json{
			"message": "密码错误",
		})
	} else {
		//	生成token
		// 登录成功，生成JWT Token并返回用户信息
		token, err1 := facades.Auth(ctx).Login(&user_exist)
		if err1 != nil {
			return ctx.Response().Status(http.StatusInternalServerError).Json(http.Json{
				"message": "token生成失败",
			})
		}

		return ctx.Response().Status(http.StatusOK).Json(http.Json{
			"message": "登录成功",
			"data": struct {
				Token string     `json:"token"`
				User  models.Emp `json:"user"`
			}{
				Token: token,
				User:  user_exist,
			},
		})
	}
}

// H5Login H5端登录接口，使用手机号(mobile)和密码进行身份验证
// 验证流程：参数校验 → 查询用户 → 管理员权限校验 → 密码验证 → 生成JWT Token
// 与AdminLogin的区别：使用手机号作为账号标识，且强制校验仅管理员(ID=1)可登录
func (r *AuthController) H5Login(ctx http.Context) http.Response {
	var user models.Emp
	ctx.Request().Bind(&user)
	password := user.Password
	//验证
	// 参数验证：手机号和密码均为必填项
	validator, err := facades.Validation().Make(ctx, map[string]any{
		"mobile":   ctx.Request().Input("mobile", ""),
		"password": ctx.Request().Input("password", ""),
	}, map[string]any{
		"mobile":   "required",
		"password": "required",
	}, validation.Messages(map[string]string{
		"mobile.required":   "手机号不能为空",
		"password.required": "密码不能为空",
	}))
	if err != nil {
		return ctx.Response().Status(http.StatusInternalServerError).Json(http.Json{
			"message": "参数校验失败",
		})
	}
	if validator.Fails() {
		return ctx.Response().Status(http.StatusUnprocessableEntity).Json(http.Json{
			"errors": validator.Errors().All(),
		})
	}
	//手机号密码验证
	// 根据工号查询用户是否存在
	facades.Orm().Query().Model(&user).Where("workno", user.WorkNo).Find(&user)

	if user.ID == 0 {
		ctx.Request().AbortWithStatusJson(401, http.Json{
			"message": "error",
			"fail":    "用户不存在,请点击注册",
		})
		return nil
	}
	var user_exist models.Emp
	facades.Orm().Query().Model(&user_exist).Where("workno", user.WorkNo).Find(&user_exist)
	//解密
	// H5端强制管理员权限校验，仅允许ID=1的管理员登录
	if user_exist.ID != 1 {
		return ctx.Response().Status(http.StatusInternalServerError).Json(http.Json{
			"message": "无权登录",
		})
	}

	// 哈希密码比对验证
	if !facades.Hash().Check(password, user_exist.Password) {
		return ctx.Response().Status(http.StatusInternalServerError).Json(http.Json{
			"message": "密码错误",
		})
	} else {
		//	生成token
		// 登录成功，生成JWT Token并返回用户信息
		token, err1 := facades.Auth(ctx).Login(&user_exist)
		if err1 != nil {
			return ctx.Response().Status(http.StatusInternalServerError).Json(http.Json{
				"message": "token生成失败",
			})
		}

		return ctx.Response().Status(http.StatusOK).Json(http.Json{
			"message": "登录成功",
			"data": struct {
				Token string     `json:"token"`
				User  models.Emp `json:"user"`
			}{
				Token: token,
				User:  user_exist,
			},
		})
	}
}

// Login 第三方/小程序登录接口，通过手机号直接登录，无需密码验证
// 登录
// 验证流程：手机号非空校验 → 查询用户 → 直接生成JWT Token（无密码验证）
// 适用于已通过第三方平台（微信等）完成身份认证的场景
func (r *AuthController) Login(ctx http.Context) http.Response {
	var user models.Emp
	mobile := ctx.Request().Input("mobile", "")
	openid := ctx.Request().Input("openid", "")
	unionid := ctx.Request().Input("unionid", "")
	// 记录登录请求参数，便于调试和审计
	facades.Log().Info("mobile", mobile)
	facades.Log().Info("openid", openid)
	facades.Log().Info("unionid", unionid)
	if mobile == "" {
		ctx.Request().AbortWithStatusJson(401, http.Json{
			"error": "手机号不能为空",
		})
		return nil
	}
	// 根据手机号查询用户是否存在
	facades.Orm().Query().Model(&models.Emp{}).Where("mobile=?", mobile).Find(&user)
	if user.ID == 0 {
		ctx.Request().AbortWithStatusJson(500, http.Json{
			"error": "用户不存在,请点击注册",
		})
		return nil
	} else {
		// 用户存在，直接生成JWT Token进行授权登录
		if token, err2 := facades.Auth(ctx).Login(&user); err2 != nil {
			return ctx.Response().Json(http.StatusInternalServerError, http.Json{
				"error": "用户授权失败",
			})

		} else {
			return ctx.Response().Success().Json(http.Json{
				"data": map[string]interface{}{
					"token": token,
					"user":  user,
				},
				"message": "登录成功",
			})
		}
	}
}

// Logout 退出登录接口，清除当前用户的登录状态
// Logout 退出登录
// 当前为空实现，Token失效由客户端移除Token或服务端Token过期策略处理
func (r *AuthController) Logout(ctx http.Context) http.Response {
	return nil
}
