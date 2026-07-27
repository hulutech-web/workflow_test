// Package middleware 提供 HTTP 中间件，包括 JWT 认证拦截、跨域处理等。
package middleware

import (
	"errors" // 标准库 errors，用于错误包装与比较

	"github.com/goravel/framework/auth"           // Goravel JWT 认证核心库，提供 token 解析与校验
	"github.com/goravel/framework/contracts/http" // Goravel HTTP 上下文契约接口
	"github.com/goravel/framework/facades"        // Goravel 门面模式入口，用于获取认证等核心服务
)

// Jwt JWT 认证中间件结构体，实现 http.Middleware 接口。
// 用于拦截需要登录态的路由请求，从 Authorization 请求头中提取 Bearer token 并校验其合法性。
type Jwt struct{}

// Handle 执行 JWT 认证中间件逻辑。
// 从请求头的 Authorization 字段中提取 Bearer token，调用框架认证服务进行解析与验证。
// token 为空、解析失败或已过期均返回 HTTP 401 状态码并中断请求链；
// 验证通过则调用 Next() 将请求向下传递至后续中间件或最终处理器。
func (a *Jwt) Handle(ctx http.Context) {
	// 获取请求头中的 Authorization 字段，预期格式为 "Bearer <token>"
	token := ctx.Request().Header("Authorization", "")
	// 如果 token 为空（请求未携带认证信息），返回 401 未授权并终止请求
	if token == "" {
		ctx.Request().AbortWithStatus(401)
		return
	}
	// 去除 "Bearer " 前缀（共 7 个字符），提取纯 JWT token 字符串
	token = token[7:]

	// 调用框架门面 Auth 服务解析 token，返回解析后的载荷 payload
	_, err := facades.Auth(ctx).Parse(token)
	// 调试用：如需查看 payload 内容可取消下行注释
	//fmt.Println(payload)
	if err != nil {
		// token 解析或验证失败（签名无效、格式错误等），返回 401 未授权并终止请求
		ctx.Request().AbortWithStatus(401)
		return
	}
	// 使用 errors.Is 判断错误链中是否包含 token 过期错误
	is := errors.Is(err, auth.ErrorTokenExpired)
	if is {
		// token 已过期，返回 401 未授权并终止请求
		ctx.Request().AbortWithStatus(401)
		return
	}
	// 认证校验全部通过，放行至后续中间件或请求处理器
	ctx.Request().Next()
}

// Signature 返回 JWT 中间件的唯一签名标识字符串。
// Goravel 框架通过此签名在中间件注册表中区分不同中间件实例。
func (a *Jwt) Signature() string {
	return "jwt"
}

// NewJwt 创建并返回 JWT 中间件的新实例。
// 通常在路由组注册时调用，例如：middleware.Jwt()，作为中间件参数传入路由定义。
func NewJwt() http.Middleware {
	return &Jwt{}
}
