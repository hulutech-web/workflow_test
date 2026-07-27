package models

import (
	"github.com/goravel/framework/database/orm"
	"strings"
)

// Dept represents a department in the organization hierarchy
// 部门模型，表示组织架构中的一个部门
type Dept struct {
	orm.Model
	// DeptName is the display name of the department
	// 部门名称
	DeptName string `gorm:"column:dept_name;not null;default:''" json:"dept_name" form:"dept_name"`
	// Pid is the parent department ID, 0 means top-level department
	// 父部门ID，0 表示顶级部门
	Pid uint `gorm:"column:pid;not null;default:0" json:"pid" form:"pid"`
	// DirectorID is the employee ID of the department director
	// 部门主管的员工ID
	DirectorID int `gorm:"column:director_id;not null;default:0" json:"director_id" form:"derector_id"` // 部门主管
	// ManagerID is the employee ID of the department manager
	// 部门经理的员工ID
	ManagerID int `gorm:"column:manager_id;not null;default:0" json:"manager_id" form:"manager_id"` // 部门经理
	// Rank is the sorting order of the department (lower value = higher priority)
	// 部门排序（值越小优先级越高）
	Rank int `gorm:"column:rank;not null;default:1" json:"rank" form:"rank"`
	// Html is the visual indentation string for tree display (e.g. repeated prefix characters)
	// 树形展示时的缩进 HTML 字符串（如重复的前缀字符）
	Html string `gorm:"column:html;null;default:''" json:"html" form:"html"`
	// Level is the depth level in the department tree (1 = top level)
	// 在部门树中的层级深度（1 表示顶级）
	Level int `gorm:"column:level;null;default:0" json:"level" form:"level"`
	// Director is the associated employee (director) loaded via foreign key DirectorID
	// 关联部门主管员工，通过外键 DirectorID 加载
	Director *Emp `gorm:"foreignkey:DirectorID"` // 关联主管
	// Manager is the associated employee (manager) loaded via foreign key ManagerID
	// 关联部门经理员工，通过外键 ManagerID 加载
	Manager *Emp `gorm:"foreignkey:ManagerID"` // 关联经理
}

// Recursion builds a flattened department tree list ordered by hierarchy.
// It traverses the department slice recursively, matching children by Pid,
// assigning each node an indentation html string and a depth level.
//
// models: the flat slice of all departments
// html:   the indentation string to repeat at each level (e.g. "—")
// pid:    the parent department ID to match for the current recursion level
// level:  the current tree depth (starting from 0)
//
// Recursion 递归构建按层级排序的部门列表。
// 通过 Pid 匹配子部门，为每个节点分配缩进字符串和层级深度。
//
// models: 所有部门的扁平切片
// html:   每层级要重复的缩进字符串（如 "—"）
// pid:    当前递归层级要匹配的父部门 ID
// level:  当前树的深度（从 0 开始）
func (d *Dept) Recursion(models []Dept, html string, pid uint, level int) []Dept {
	var result []Dept
	for i, dept := range models {
		if dept.Pid == pid {
			// Assign indentation html and depth level to the matched department
			// 为匹配到的部门分配缩进字符串和层级深度
			dept.Html = strings.Repeat(html, level)
			dept.Level = level + 1
			result = append(result, dept)
			// Recursively process remaining departments at the next depth level
			// 递归处理剩余部门，进入下一层级
			result = append(result, d.Recursion(append([]Dept{}, models[i+1:]...), html, dept.ID, level+1)...)
		}
	}
	return result
}
