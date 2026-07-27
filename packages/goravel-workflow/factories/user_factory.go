package factories

import "github.com/brianvoe/gofakeit/v6"

// UserFactory 用户模型工厂，用于生成测试用的假用户数据。
type UserFactory struct {
}

// Definition 定义模型的默认状态 / Define the model's default state.
func (f *UserFactory) Definition() map[string]any {
	return map[string]any{
		"nickName":  gofakeit.Name(),           // 昵称：随机生成人名
		"AvatarUrl": gofakeit.ImageURL(100, 100), // 头像：随机生成100x100的占位图片URL
		"Mobile":    gofakeit.Phone(),            // 手机号：随机生成电话号码
		"Openid":    gofakeit.UUID(),             // 微信OpenID：随机生成UUID
		"Unionid":   gofakeit.UUID(),             // 微信UnionID：随机生成UUID
		"Address":   []string{},                  // 地址：空字符串数组
	}
}
