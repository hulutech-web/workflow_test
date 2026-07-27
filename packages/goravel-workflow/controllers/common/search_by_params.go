package common

import (
	"github.com/goravel/framework/contracts/database/orm"
)

// SearchByParamsService 通用参数搜索服务，用于根据请求参数构建动态查询条件
type SearchByParamsService struct {
}

// SearchByParams 根据传入的 params map 构建 ORM 查询条件闭包
// params: 包含字段名到值的映射，例如 {"name": "张三", "status": "1"}
// excepts: 需要从 params 中排除的键名，这些字段不会作为查询条件
// 返回值是一个闭包，可传入 orm.Query 的 Where 等方法中用于链式调用
func (s *SearchByParamsService) SearchByParams(params map[string]string, excepts ...string) func(methods orm.Query) orm.Query {
	// 先删除需要排除的字段，避免它们被当作查询条件
	for _, except := range excepts {
		delete(params, except)
	}

	// 返回查询构建闭包
	return func(query orm.Query) orm.Query {
		// 遍历所有参数，为每个非空的普通字段添加模糊查询条件
		for key, value := range params {
			// pappt_name 字段即使有值也跳过，由业务层单独处理
			if key == "pappt_name" && value != "" {
				continue
			}

			// 跳过空值以及分页/排序等系统参数，这些不应作为业务查询条件
			if value == "" || key == "pageSize" || key == "total" || key == "currentPage" || key == "sort" || key == "order" {
				continue
			} else {
				// 构建 LIKE 模糊查询：WHERE key LIKE '%value%'
				query = query.Where(key+" like ?", "%"+value+"%")
			}
		}
		return query
	}
}
