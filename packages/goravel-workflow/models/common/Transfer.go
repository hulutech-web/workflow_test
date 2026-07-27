package common

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/goravel/framework/facades"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// FieldValue represents a JSON array of string values stored as a database field
// 字段值类型，表示存储为数据库字段的 JSON 字符串数组
type FieldValue []string

// RuleItem defines a single condition/rule configuration item
// 规则项，定义单个条件/规则的配置项
type RuleItem struct {
	// RuleName is the rule identifier/key 规则名称/标识键
	RuleName string `json:"rule_name" form:"rule_name"`
	// RuleTitle is the human-readable label for this rule 规则标题/显示名称
	RuleTitle string `json:"rule_title" form:"rule_title"`
	// RuleValue is the expected value for this rule 规则匹配值
	RuleValue string `json:"rule_value" form:"rule_value"`
}

// Rule is a collection of RuleItem entries representing a complete rule set
// 规则集合，表示一组完整的规则项
type Rule []RuleItem

// Scan implements sql.Scanner for reading JSON from database into Rule
// 实现 sql.Scanner 接口，从数据库读取 JSON 并反序列化为 Rule
func (t *Rule) Scan(value interface{}) error {
	bytesValue, _ := value.([]byte)
	return json.Unmarshal(bytesValue, t)
}

// Value implements driver.Valuer for writing Rule as JSON to database
// 实现 driver.Valuer 接口，将 Rule 序列化为 JSON 写入数据库
func (t Rule) Value() (driver.Value, error) {
	// If t is nil, return nil 如果 t 为 nil，返回 nil
	return json.Marshal(t)
}

// Scan implements sql.Scanner for reading JSON from database into FieldValue
// 实现 sql.Scanner 接口，从数据库读取 JSON 并反序列化为 FieldValue
func (t *FieldValue) Scan(value interface{}) error {
	bytesValue, _ := value.([]byte)
	return json.Unmarshal(bytesValue, t)
}

// Value implements driver.Valuer for writing FieldValue as JSON to database
// 实现 driver.Valuer 接口，将 FieldValue 序列化为 JSON 写入数据库
func (t FieldValue) Value() (driver.Value, error) {
	return json.Marshal(t)
}

// Coordinates represents a geographic map location with longitude and latitude
// 地图坐标位置，包含经度和纬度
type Coordinates struct {
	// Longitude is the east-west coordinate 经度（东西方向坐标）
	Longitude float64 `json:"longitude"`
	// Latitude is the north-south coordinate 纬度（南北方向坐标）
	Latitude float64 `json:"latitude"`
}

// CoordRes represents a GeoJSON-compatible coordinate result for spatial database operations
// 坐标结果，表示兼容 GeoJSON 格式的空间数据库坐标数据
type CoordRes struct {
	// Type is the GeoJSON geometry type (e.g. "Point") GeoJSON 几何类型（例如 "Point"）
	Type string `json:"type"`
	// Coordinates is the [longitude, latitude] pair 坐标数组 [经度, 纬度]
	Coordinates []float64 `json:"coordinates"`
}

// GormDataType returns the SQL data type "geometry" for geographic spatial data
// 返回 SQL 数据类型 "geometry"，用于地理空间数据
func (c CoordRes) GormDataType() string {
	return "geometry"
}

// GormValue generates a SQL expression from coordinate data for spatial queries
// 根据坐标数据生成 SQL 表达式，用于空间查询
func (c CoordRes) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	// If Coordinates is nil, return SQL NULL 如果坐标为空，返回 SQL NULL
	if c.Coordinates == nil {
		// Assign NULL value 赋值为 NULL
		return clause.Expr{SQL: "NULL", Vars: []interface{}{}}
	}

	// Generate SQL expression using ST_GeomFromText with SRID 4326 (WGS 84)
	// 使用 ST_GeomFromText 生成 SQL 表达式，SRID 4326（WGS 84 坐标系）
	return clause.Expr{
		SQL:  "ST_GeomFromText(?,4326)",
		Vars: []interface{}{fmt.Sprintf("POINT(%f %f)", c.Coordinates[1], c.Coordinates[0])},
	}
}

// Scan implements sql.Scanner to decode spatial data from database into CoordRes
// 实现 sql.Scanner 接口，从数据库解码空间数据为 CoordRes
func (c *CoordRes) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid value type, expected []byte")
	}
	if len(bytes) == 0 {
		c = nil
		return nil
	}
	var coordRes string
	querySql := "SELECT ST_AsGeoJSON(?) as coord"
	param := string(bytes)
	facades.Orm().Query().Raw(querySql, param).Pluck("coord", &coordRes)
	err := json.Unmarshal([]byte(coordRes), c)
	return err
}
