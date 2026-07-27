package captcha

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/wenlng/go-captcha-assets/helper"

	"github.com/wenlng/go-captcha-assets/resources/images"
	"github.com/wenlng/go-captcha/v2/base/option"
	"github.com/wenlng/go-captcha/v2/rotate"
)

// CaptchaService 验证码服务，提供图片旋转验证码的生成与校验功能
type CaptchaService struct {
}

// rotateCapt 全局的旋转验证码实例，在 init() 中初始化
var rotateCapt rotate.Captcha

// init 包初始化函数，构建旋转验证码实例并加载背景图片资源
func init() {
	// 创建旋转验证码构建器，设置旋转角度范围为 20° ~ 330°
	builder := rotate.NewBuilder(rotate.WithRangeAnglePos([]option.RangeVal{
		{Min: 20, Max: 330},
	}))

	// background images / 加载背景图片资源
	imgs, err := images.GetImages()
	if err != nil {
		log.Fatalln(err)
	}

	// set resources / 将背景图片设置到构建器中
	builder.SetResources(
		rotate.WithImages(imgs),
	)

	// 构建最终的旋转验证码实例
	rotateCapt = builder.Make()
}

// Generate 生成一组旋转验证码
// 返回值: captcha_key（验证码唯一键）, code（状态码）, image_base64（主图 Base64）, thumb_base64（缩略图 Base64）, error
func (c *CaptchaService) Generate() (string, int, string, string, error) {
	code := 0
	image_base64 := ""
	thumb_base64 := ""

	// 调用底层库生成验证码数据
	captchaData, err := rotateCapt.Generate()

	// 将验证码图案数据（点阵）序列化为 JSON 并计算 MD5 作为缓存键
	dotsByte, _ := json.Marshal(captchaData.GetData())
	captcha_key := helper.StringToMD5(string(dotsByte))

	// 加入缓存 / 将验证码数据写入缓存，供后续校验使用
	WriteCache(captcha_key, dotsByte)

	if err != nil {
		return captcha_key, code, image_base64, thumb_base64, err
	}

	// 将主图和缩略图分别转为 Base64 字符串
	image_base64, _ = captchaData.GetMasterImage().ToBase64()
	thumb_base64, _ = captchaData.GetThumbImage().ToBase64()

	return captcha_key, code, image_base64, thumb_base64, nil
}

// CheckAngle 校验用户提交的旋转角度是否与验证码匹配
// 参数 angle: 用户拖拽/输入的旋转角度
// 参数 key: 验证码唯一键，用于从缓存中读取验证码数据
// 返回值: code（0 表示校验通过，1 表示校验失败）, bool（true: 通过, false: 不通过）
func (c *CaptchaService) CheckAngle(angle string, key string) (int, bool) {
	code := 1

	// 角度或键为空，直接返回校验失败
	if angle == "" || key == "" {
		return code, false
	}

	// 从缓存中读取验证码数据
	cacheDataByte := ReadCache(key)
	if len(cacheDataByte) == 0 {
		return code, false
	}

	// 将缓存数据反序列化为旋转验证码的数据块
	var dct *rotate.Block
	if err := json.Unmarshal(cacheDataByte, &dct); err != nil {
		return code, false
	}

	// 解析用户提交的角度值
	sAngle, _ := strconv.ParseFloat(fmt.Sprintf("%v", angle), 64)

	// 调用底层库进行角度校验，允许 2° 的误差范围
	chkRet := rotate.CheckAngle(int64(sAngle), int64(dct.Angle), 2)

	if chkRet {
		code = 0
		return code, true
	} else {
		return code, false
	}
}
