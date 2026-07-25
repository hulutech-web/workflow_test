package seeders

import (
	"encoding/json"
	models "goravel/app/models"

	"github.com/goravel/framework/facades"
)

type MenuSeeder struct {
}

// Signature The name and signature of the seeder.
func (s *MenuSeeder) Signature() string {
	return "MenuSeeder"
}

// Run executes the seeder logic.
func (s *MenuSeeder) Run() error {
	menus := []models.Menu{
		{
			PID:          0,
			Title:        "工作台",
			Name:         "dashboard",
			Path:         "/dashboard",
			Component:    "@/layout/default-layout.vue",
			Icon:         "icon-dashboard",
			Locale:       "menu.dashboard",
			RequiresAuth: true,
			Roles:        mustMarshalJSON([]string{"*"}),
			Order:        0,
			Sort:         0,

			Children: []models.Menu{
				{
					PID:          0,
					Title:        "工作台",
					Name:         "Workplace",
					Path:         "workplace",
					Component:    "@/views/dashboard/workplace/index.vue",
					Locale:       "menu.dashboard.workplace",
					RequiresAuth: true,
					Roles:        mustMarshalJSON([]string{"*"}),
					Order:        0,
					Sort:         0,
				},
			},
		},
		{
			PID:          0,
			Title:        "项目管理",
			Name:         "project",
			Path:         "/project",
			Component:    "@/layout/default-layout.vue",
			Icon:         "icon-apps",
			Locale:       "menu.project",
			RequiresAuth: true,
			Roles:        mustMarshalJSON([]string{"*"}),
			Order:        1,
			Sort:         1,

			Children: []models.Menu{
				{
					PID:          0,
					Title:        "项目设计",
					Name:         "ProjectDesign",
					Path:         "design",
					Component:    "@/views/project/design/index.vue",
					Locale:       "menu.project.design",
					RequiresAuth: true,
					Roles:        mustMarshalJSON([]string{"*"}),
					Order:        0,
					Sort:         0,
				},
				{
					PID:          0,
					Title:        "章节管理",
					Name:         "ChapterManage",
					Path:         ":id/chapter_manage",
					Component:    "@/views/chapter/manage.vue",
					Locale:       "menu.project.chapter",
					RequiresAuth: true,
					Roles:        mustMarshalJSON([]string{"*"}),
					HideInMenu:   true, // 在菜单中隐藏
					Order:        1,
					Sort:         1,
				},
			},
		},
		{
			PID:          0,
			Title:        "系统配置",
			Name:         "system",
			Path:         "/system",
			Component:    "@/layout/default-layout.vue",
			Icon:         "icon-settings",
			Locale:       "menu.system",
			RequiresAuth: true,
			Roles:        mustMarshalJSON([]string{"*"}),
			Order:        2,
			Sort:         2,

			Children: []models.Menu{
				{
					PID:          0,
					Title:        "全局配置",
					Name:         "SystemSettings",
					Path:         "settings",
					Component:    "@/views/system/settings/index.vue",
					Locale:       "menu.system.settings",
					RequiresAuth: true,
					Roles:        mustMarshalJSON([]string{"*"}),
					Order:        0,
					Sort:         0,
				},
			},
		},
	}

	for i := range menus {
		if err := facades.Orm().Query().Create(&menus[i]); err != nil {
			return err
		}

		if len(menus[i].Children) > 0 {
			for j := range menus[i].Children {
				menus[i].Children[j].PID = menus[i].ID
				if err := facades.Orm().Query().Create(&menus[i].Children[j]); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func mustMarshalJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(data)
}
