package controllers

import (
	"fmt"
	"goravel/packages/goravel-workflow/services/captcha"

	"github.com/goravel/framework/contracts/http"
	httpfacades "github.com/hulutech-web/http_result"
)

// CaptchaController 验证码控制器，提供图形验证码的生成与校验功能
type CaptchaController struct {
	//Dependent services // 依赖的服务
	*captcha.CaptchaService
}

// NewCaptchaController 创建验证码控制器实例，注入 CaptchaService 依赖
func NewCaptchaController() *CaptchaController {
	return &CaptchaController{
		//Inject servicesGetCaptcha // 注入验证码服务
	}
}

// GetCaptcha 获取图形验证码（GET /api/captcha/get）
// 生成验证码图片并返回 captcha_key、验证码文本、Base64 编码的图片及缩略图
func (r *CaptchaController) GetCaptcha(ctx http.Context) http.Response {
	// 调用 CaptchaService 生成验证码，返回 key、文本、原图 base64、缩略图 base64
	captcha_key, code, image_base64, thumb_base64, err := r.Generate()
	if err != nil {
		// 生成失败，返回 500 错误
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "error", err)
	}
	// 生成成功，返回验证码相关数据
	return httpfacades.NewResult(ctx).Success("", http.Json{
		"captcha_key":  captcha_key,
		"code":         code,
		"image_base64": image_base64,
		"thumb_base64": thumb_base64,
	})
}

// ValidateCaptcha 校验图形验证码（POST /api/captcha/validate）
// 接收用户旋转角度和 captcha_key，调用服务层校验是否匹配
func (r *CaptchaController) ValidateCaptcha(ctx http.Context) http.Response {
	// 从请求中获取用户输入的旋转角度
	angle := ctx.Request().InputInt64("angle")
	// 从请求中获取验证码 key（默认为 "captcha_key"）
	captcha_key := ctx.Request().Input("captcha_key", "captcha_key")
	// 调用 CaptchaService.CheckAngle 校验旋转角度是否正确
	code, isOk := r.CheckAngle(fmt.Sprintf("%d", angle), captcha_key)
	// 返回校验结果：是否通过及状态码
	return httpfacades.NewResult(ctx).Success("", http.Json{"is_ok": isOk, "code": code})
}
