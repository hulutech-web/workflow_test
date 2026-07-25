package models

import (
	"encoding/json"
	"sort"

	"github.com/goravel/framework/database/orm"
)

type Menu struct {
	orm.Model
	PID          uint   `gorm:"column:pid;default:0" form:"pid" json:"pid"`
	Title        string `gorm:"column:title" form:"title" json:"title"`
	Name         string `gorm:"column:name" form:"name" json:"name"`
	Path         string `gorm:"column:path" form:"path" json:"path"`
	Component    string `gorm:"column:component" form:"component" json:"component"`
	Icon         string `gorm:"column:icon" form:"icon" json:"icon"`
	Locale       string `gorm:"column:locale" form:"locale" json:"locale"`
	RequiresAuth bool   `gorm:"column:requires_auth;default:true" form:"requires_auth" json:"requiresAuth"`
	Roles        string `gorm:"column:roles;type:text" form:"roles" json:"roles"`
	HideInMenu   bool   `gorm:"column:hide_in_menu;default:false" form:"hideInMenu" json:"hideInMenu"`
	Order        int    `gorm:"column:order;default:0" form:"order" json:"order"`
	Sort         int    `gorm:"column:sort;default:0" form:"sort" json:"sort"`
	Target       string `gorm:"column:target" form:"target" json:"target"`
	Badge        string `gorm:"column:badge" form:"badge" json:"badge"`
	Children     []Menu `gorm:"foreignKey:PID;references:ID" json:"children"`
}

func (m *Menu) GetRoles() []string {
	if m.Roles == "" {
		return []string{"*"}
	}
	var roles []string
	err := json.Unmarshal([]byte(m.Roles), &roles)
	if err != nil {
		return []string{"*"}
	}
	return roles
}

// ToRouteRecord 将 Menu 转换为前端路由格式
func (m *Menu) ToRouteRecord() map[string]interface{} {
	route := map[string]interface{}{
		"path": m.Path,
		"name": m.Name,
		"meta": map[string]interface{}{
			"locale":       m.Locale,
			"requiresAuth": m.RequiresAuth,
			"icon":         m.Icon,
			"order":        m.Order,
			"roles":        []string{"*"}, // 可以从 Roles 字段解析
		},
	}

	// 处理组件
	if m.Component != "" {
		route["component"] = m.Component
	}

	// 处理子路由
	if len(m.Children) > 0 {
		children := make([]map[string]interface{}, 0)
		for _, child := range m.Children {
			children = append(children, child.ToRouteRecord())
		}
		route["children"] = children
	}

	return route
}

// BuildMenuTree 构建菜单树形结构
// menus 是所有的菜单列表
// parentID 是当前层级的父ID，顶层菜单的parentID通常为0
func (m *Menu) BuildMenuTree(menus []Menu, parentID uint) []Menu {
	var tree []Menu

	for _, menu := range menus {
		if menu.PID == parentID {
			// 递归查找子菜单
			children := m.BuildMenuTree(menus, menu.ID)
			if len(children) > 0 {
				menu.Children = children
			}
			tree = append(tree, menu)
		}
	}

	// 按Sort字段排序
	sort.Slice(tree, func(i, j int) bool {
		return tree[i].Sort < tree[j].Sort
	})

	return tree
}
