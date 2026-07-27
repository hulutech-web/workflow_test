package models

import (
	"github.com/goravel/framework/database/orm"
)

// Product 产品模型，用于存储产品信息
// Product represents a product entity
type Product struct {
	orm.Model
	Name          string  `gorm:"column:name;type:varchar(255);not null" form:"name" json:"name"`                   // Name 产品名称 名称
	Special       string  `gorm:"column:special;type:varchar(255);not null" form:"special" json:"special"`             // Special 产品规格 规格
	Dimension     string  `gorm:"column:dimension;type:varchar(255);not null" form:"dimension" json:"dimension"`       // Dimension 产品尺寸（单位/维度） 尺寸
	Quantity      int     `gorm:"column:quantity;type:int(10);not null" form:"quantity" json:"quantity"`                // Quantity 产品数量 数量
	Unit          string  `gorm:"column:unit;type:varchar(255);not null" form:"unit" json:"unit"`                      // Unit 计量单位（如：个、箱、台） 单位
	UnitPrice     float64 `gorm:"column:unit_price;type:float(10,2);not null" form:"unit_price" json:"unit_price"`     // UnitPrice 产品单价 单价
	DiscountPrice float64 `gorm:"column:discount_price;type:float(10,2);not null" form:"discount_price" json:"discount_price"` // DiscountPrice 折后价格 折后单价
	Amount        float64 `gorm:"column:amount;type:float(10,2);not null" form:"amount" json:"amount"`                 // Amount 总金额（单价 × 数量 或折后价 × 数量） 总金额
	Description   string  `gorm:"column:description;type:varchar(255);" form:"description" json:"description"`          // Description 产品描述/备注 描述
	ImageURL      string  `gorm:"column:image_url;type:varchar(255);" form:"image_url" json:"image_url"`              // ImageURL 产品图片链接 图片链接
}
