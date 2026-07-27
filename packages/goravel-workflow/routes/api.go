// Package routes 工作流引擎 API 路由注册
// 定义所有与工作流相关的前后端交互接口，包括认证、部门、员工、流程、审批等模块的路由映射
package routes

import (
	// 工作流控制器
	controllers "goravel/packages/goravel-workflow/controllers"
	// JWT 中间件
	"goravel/packages/goravel-workflow/middleware"

	// Goravel 框架核心契约接口
	"github.com/goravel/framework/contracts/foundation"
	// 路由契约接口，提供路由注册与分组能力
	"github.com/goravel/framework/contracts/route"
	// 框架门面，用于获取全局服务实例
	"github.com/goravel/framework/facades"
)

// Api 注册工作流引擎的所有 API 路由
// 包含公开路由（无需认证）和受 JWT 中间件保护的私有路由两组。
// 公开路由包括登录、验证码等接口；私有路由涵盖部门管理、员工管理、
// 流程定义、审批流转、抄送、归档等核心业务功能。
func Api(app foundation.Application) {
	// 通过应用程序实例创建路由器
	router := app.MakeRoute()

	// ============================================================
	// 公开路由区域 —— 无需 JWT 认证即可访问
	// ============================================================

	// 认证控制器：处理管理端与 H5 端登录
	authController := controllers.NewAuthController()
	router.Post("/api/auth/login", authController.AdminLogin) // 管理后台登录
	router.Post("/api/h5/login", authController.H5Login)       // H5 移动端登录

	// 验证码控制器：生成与校验图形验证码
	captchaController := controllers.NewCaptchaController()
	router.Get("/api/captcha/get", captchaController.GetCaptcha)            // 获取验证码图片
	router.Post("/api/captcha/validate", captchaController.ValidateCaptcha) // 校验验证码

	// ============================================================
	// 私有路由区域 —— 需要 JWT 认证，统一加 /api 前缀
	// ============================================================
	facades.Route().Middleware(middleware.NewJwt()).Prefix("/api").Group(func(router route.Router) {

		// -------------------- 文件上传 --------------------
		// 文件上传
		uploadCtrl := controllers.NewUploadController()
		router.Post("/upload", uploadCtrl.Upload) // 上传文件（图片、附件等）

		// -------------------- 首页 --------------------
		// 首页控制器：获取仪表盘或工作台概览数据
		homeCtrl := controllers.NewHomeController()
		router.Get("/home", homeCtrl.Index) // 首页数据接口

		// -------------------- 部门管理 --------------------
		// 部门
		deptCtrl := controllers.NewDeptController()
		router.Resource("dept", deptCtrl)                       // RESTful 资源路由（CRUD）
		router.Get("dept/list", deptCtrl.List)                  // 部门列表（带分页、搜索）
		router.Post("dept/bindmanager", deptCtrl.BindManager)   // 绑定部门经理
		router.Post("dept/binddirector", deptCtrl.BindDirector) // 绑定部门主管

		// -------------------- 员工管理 --------------------
		// 员工
		empCtrl := controllers.NewEmpController()
		router.Resource("emp", empCtrl)            // RESTful 资源路由（CRUD）
		router.Post("emp/search", empCtrl.Search)  // 员工搜索（关键字模糊查询）
		router.Get("emp/options", empCtrl.Options) // 员工下拉选项（用于前端选择器）
		router.Post("emp/bind", empCtrl.BindUser)  // 绑定员工与系统用户账号

		// -------------------- 流程定义 --------------------
		// 流程
		flowCtrl := controllers.NewFlowController()
		router.Resource("flow", flowCtrl)          // RESTful 资源路由（CRUD）
		router.Get("flow/list", flowCtrl.List)     // 流程列表（带分页、搜索、筛选）
		router.Get("flow/create", flowCtrl.Create) // 创建流程页面数据
		// 流程设计
		router.Get("flow/flowchart/{id}", flowCtrl.FlowDesign) // 流程设计器：获取流程图 JSON 数据
		router.Post("flow/publish", flowCtrl.Publish)          // 发布流程（将草稿状态改为已发布）

		// -------------------- 流程实例（Entry）--------------------
		// entry 节点
		entryCtrl := controllers.NewEntryController()
		router.Get("flow/{id}/entry", entryCtrl.Create)         // 发起流程：获取对应流程的发起表单
		router.Get("entry", entryCtrl.Index)                    // 流程实例列表（我的申请/待我审批等）
		router.Post("entry", entryCtrl.Store)                   // 提交新流程实例（正式发起）
		router.Get("entry/{id}", entryCtrl.Show)                // 查看流程实例详情
		router.Put("entry/{id}", entryCtrl.Update)              // 更新流程实例信息
		router.Get("entry/{id}/entrydata", entryCtrl.EntryData) // 获取流程实例的表单数据
		// 流程重发
		router.Post("entry/resend", entryCtrl.Resend) // 重发流程（驳回或完成后重新发起，Circle 轮次 +1）
		// 撤回流程
		router.Post("entry/revoke", entryCtrl.Revoke) // 发起人撤回待审批的流程实例
		// 流程轨迹
		flowlinkCtrl := controllers.NewFlowlinkController()
		router.Post("flowlink", flowlinkCtrl.Update) // 更新流程流转连线配置（步骤间关系）

		// -------------------- 表单模板 --------------------
		// 模板
		templateCtrl := controllers.NewTemplateController()
		router.Resource("template", templateCtrl)         // RESTful 资源路由（CRUD）
		router.Get("template/option", templateCtrl.Option) // 模板下拉选项（用于流程关联模板时选择）
		// 模板控件
		templateformCtrl := controllers.NewTemplateformController()
		router.Get("template/{id}/templateform", templateformCtrl.Index)   // 获取某模板下的所有表单控件列表
		router.Post("templateform", templateformCtrl.Store)                 // 创建表单控件
		router.Put("templateform/{id}", templateformCtrl.Update)            // 更新表单控件
		router.Delete("templateform/{id}", templateformCtrl.Destroy)        // 删除表单控件
		router.Get("templateform/{id}", templateformCtrl.Show)              // 查看单个表单控件详情
		router.Post("flow/templateform", templateformCtrl.FlowTemplateForm) // 流程关联的表单控件数据

		// -------------------- 流程步骤 --------------------
		// 流程
		processCtrl := controllers.NewProcessController()
		router.Resource("process", processCtrl)                // RESTful 资源路由（CRUD）
		router.Get("process/attribute", processCtrl.Attribute) // 获取步骤属性（并发类型、审批人规则等配置项）
		router.Post("process/con", processCtrl.Condition)      // 设置步骤条件分支表达式
		router.Post("process/list", processCtrl.List)          // 步骤列表（带排序和筛选）

		// -------------------- 审批流转 --------------------
		// 审批流转
		procCtrl := controllers.NewProcController()
		router.Get("proc/{entry_id}/rejectable", procCtrl.RejectableProcesses) // 获取可驳回的目标步骤列表
		router.Get("proc/{entry_id}", procCtrl.Index)                          // 获取流程实例下的审批任务列表
		// 同意
		router.Post("pass", procCtrl.Pass) // 审批通过（触发 Transfer 流转引擎，进入下一步）
		// 驳回
		router.Post("unpass", procCtrl.UnPass) // 驳回审批（退回到上一步或指定节点）
		// 撤回
		router.Post("revoke", procCtrl.Revoke) // 撤回当前审批任务
		// 加签
		router.Post("addsign", procCtrl.AddSign) // 加签：当前审批人邀请额外人员参与审批（前加签/后加签）
		// 转交
		router.Post("transfer", procCtrl.TransferProc) // 转交：将当前审批任务转给其他人处理
		// 评论
		router.Post("comment", procCtrl.AddComment)              // 添加审批评论/意见
		router.Get("comments/{entry_id}", procCtrl.GetComments)  // 获取流程实例的所有评论记录

		// -------------------- 抄送 --------------------
		// 抄送
		ccCtrl := controllers.NewCcController()
		router.Get("cc", ccCtrl.Index)                       // 抄送列表（我收到的抄送）
		router.Post("cc", ccCtrl.Store)                      // 创建抄送记录（通常在审批通过后自动触发）
		router.Get("cc/entry/{entry_id}", ccCtrl.GetEntryCC) // 查看某流程实例的抄送记录

		// -------------------- 归档审批 --------------------
		// 归档审批
		archiveCtrl := controllers.NewEntryArchiveController()
		router.Get("archive", archiveCtrl.Index)     // 归档列表（已完成的流程实例归档查看）
		router.Get("archive/{id}", archiveCtrl.Show) // 查看归档详情
	})
}
