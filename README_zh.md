<p align="center">
  <img src="https://github.com/hulutech-web/workflow_test/blob/master/assets/workflow.png?raw=true" width="300" />
</p>

<div align="center">

# Goravel Workflow

**一个基于 [Goravel](https://www.goravel.dev/) 框架的工作流审批引擎扩展包**


支持会签、或签、条件分支、驳回到指定节点、加签、转交、撤回、子流程、抄送、超时检查等完整审批流程功能。
> 本项目为完整的项目，包含前端以及后端API接口，开发者可自行改造  
> 项目基于goravel最新版本开发 v1.18


</div>

### [**这是个什么项目**](#custom-target)

### 文档

使用手册请访问[文档](https://hulutech-web.github.io/goravel-workflow.github.io/)

---

### 一、克隆项目

```shell
git clone git@github.com:hulutech-web/workflow_test.git
```

#### 1.1 注册服务提供者：bootstrap/providers.go

```go
import "github.com/goravel/packages/goravel-workflow"
```

#### 1.2 注册服务提供者：bootstrap/providers.go

```go
func Providers() []foundation.ServiceProvider {
        return []foundation.ServiceProvider{
            ...,
			&workflow.ServiceProvider{},
        }

}
```

### 二、发布资源

默认将发布配置文件和数据表迁移文件。

#### 2.1 发布资源

```shell
go run . artisan vendor:publish --package=./packages/goravel-workflow
```

可选标签：
- `--tag=migrations` — 只发布迁移文件
- `--tag=seeders` — 只发布种子数据
- `--tag=config` — 只发布配置文件
- 不指定标签则全部发布

#### 2.2 迁移文件

发布后得到两个 Go 迁移文件：
- `2024_06_24_000000_create_workflow_base_tables.go` — 创建工作流专用表（flows、flowtypes、processes、processvars、templates、templateforms、entries、entrydatas、procs、attachments、products，共 11 张表）。注：users、depts、emps 表由宿主应用提供。

#### 2.3 执行迁移建表

在 `database/seeders/database_seeder.go` 下添加：

```go
func (s *DatabaseSeeder) Run() error {
    return facades.Seeder().Call([]seeder.Seeder{
        &WorkflowDatabaseSeeder{},
    })
}
```

#### 2.4 执行迁移

```shell
go run . artisan migrate --seed
```

#### 2.5 检查路由重名

如果启动项目报错，请检查路由是否有重名，并修改路由。

#### 2.6 模型映射

发布资源后，`config/workflow.go` 中有默认的关联映射，根据需要自行修改。

### 三、实现 Hook 接口（可选）

用户自定义 User 结构中注入流程框架，并实现框架中的 Hook 接口。

```go
type User struct {
    orm.Model
    Name     string `gorm:"column:name;type:varchar(255);not null"`
    WorkNo   string `gorm:"column:workno;not null;unique_index:users_workno_unique"`
    Password string `gorm:"column:password;type:varchar(255);not null"`
    Workflow *workflow.Workflow
    orm.SoftDeletes
}
```

实现通知方法：

```go
// 通知发起人，在被驳回时调用，或者整个流程结束时调用。
func (u *User) NotifySendOne(id uint) error {
    // 发送邮件、短信等通知
    return nil
}

// 通知下一个审批人，当当前环节的审批人通过时触发。
func (u *User) NotifyNextAuditor(id uint) error {
    // 通知下一审批人
    return nil
}
```

### 四、实例化 Workflow

```go
func (receiver *AppServiceProvider) Boot(app foundation.Application) {
    wf := workflow.NewBaseWorkflow()
    user := &models.User{Workflow: wf}
    wf.RegisterHook("NotifySendOneHook", reflect.ValueOf(user.NotifySendOne))
    wf.RegisterHook("NotifyNextAuditorHook", reflect.ValueOf(user.NotifyNextAuditor))
}
```

回调参数将在 User 结构中的 NotifySendOne 和 NotifyNextAuditor 方法中执行后续操作，由开发者自行实现。

### 五、路由说明 api.go

```go
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

```

### 六、核心功能

| 功能 | 说明 | API |
|------|------|-----|
| **撤回** | 发起人撤回未处理的流程 | `POST /api/entry/revoke` |
| **加签** | 当前审批人可添加额外审批人 | `POST /api/addsign` |
| **转交** | 将审批任务转交给其他员工 | `POST /api/transfer` |
| **评论** | 多轮对话式评论/留言 | `POST /api/comment` |
| **抄送** | 节点完成后自动抄送相关人员 | `GET /api/cc/list` |
| **超时检查** | 定时检查超时的待办任务 | `artisan workflow:timeout-check` |
| **会签** | 同一步骤所有审批人都需通过 | Flowlink.ConcurrencyType = 1 |
| **或签** | 同一步骤一人通过其余跳过 | Flowlink.ConcurrencyType = 2 |
| **驳回到指定节点** | 驳回到任意历史节点 | UnPass 支持 targetProcessID 参数 |
| **表单指定审批人** | 从表单字段动态读取审批人 ID | Auditor = -1003 |
| **动态表达式审批人** | 通过表达式规则计算审批人 | Auditor = -1004 |
| **子流程** | 主流程嵌套子流程 | Process.Position = 2, ChildFlowID |

### 七、状态码说明

#### Entry 状态（entries.status）

| 值 | 含义 |
|----|------|
| 0 | 进行中 |
| 9 | 已完成 |
| -1 | 已驳回 |
| -2 | 已撤销 |
| -9 | 草稿 |

#### Proc 状态（procs.status）

| 值 | 含义 |
|----|------|
| 0 | 待审批 |
| 1 | 已通过 |
| -1 | 已驳回 |
| -2 | 已撤销 |
| 3 | 已转交 |
| 4 | 已跳过 |
| 9 | 会签通过 |

#### 并发类型（Flowlink.ConcurrencyType）

| 值 | 含义 |
|----|------|
| 0 | 依次审批（默认） |
| 1 | 会签：所有审批人都需通过 |
| 2 | 或签：一人通过其余跳过 |

#### 审批人特殊值（Flowlink.Auditor）

| 值 | 含义 |
|----|------|
| -1000 | 发起人自己 |
| -1001 | 部门主管 |
| -1002 | 部门经理 |
| -1003 | 从表单字段读取审批人 |
| -1004 | 动态表达式计算审批人 |

### 八、前端集成

请访问前端框架 [goravel-workflow-vue](https://github.com/goravel/packages/goravel-workflow-vue) 下载安装扩展，并按照文档进行集成，本项目已集成。

### 九、接口文档

请访问前端框架 [goravel-workflow-vuepress](https://github.com/goravel/packages/goravel-workflow-vuepress) 进行查看。

---

### 十、流程场景介绍

## 1. 核心概念

| 概念 | 表名 | 说明 |
|------|------|------|
| **Flow（流程）** | `flows` | 业务流程定义，如"请假流程"、"报销流程"。包含流程名称、关联的表单模板、流程图 JSON 等。 |
| **Process（环节）** | `processes` | 流程中的单个审批节点，如"主管审批"、"经理审批"。每个环节配置审批人规则、并发方式、超时时间、抄送人等。 |
| **Flowlink（连线）** | `flowlinks` | 环节之间的流转关系，定义条件分支和审批人计算规则。 |
| **Entry（流程实例）** | `entries` | 一次具体的流程发起，如"张三的请假申请"。关联发起员工、所属流程、表单数据。 |
| **Proc（审批任务）** | `procs` | Entry 在每个 Process 上产生的具体审批任务，代表一个待办审批操作。 |

**关系链**：一个 Flow 包含多个 Process，Process 之间通过 Flowlink 连接。用户发起一个 Flow 时创建 Entry，Entry 在第一个 Process 上生成 Proc，审批通过后流转至下一个 Process 再生成新 Proc。

## 2. 典型流程场景

### 依次审批（顺序审批）

审批人按固定顺序依次审批，前一个人批准后自动流转到下一个人。

- **配置**：`flowlinks.next_process_id` 指定下一步环节 ID，`concurrency_type = 0`
- **用例**：请假申请 → 主管审批 → 经理审批 → 总经理审批 → 完成

### 会签（多人全部同意才通过）

多个审批人必须全部同意，流程才能进入下一环节。只要有一人驳回，整个会签失败。

- **配置**：`flowlinks.concurrency_type = 1`
- **状态特征**：会签中的 Proc 状态为 `9 (Consensus)`，表示部分人已审批通过但尚未全部完成
- **用例**：项目立项需要技术、财务、法务三方会签

### 或签（一人同意即可通过）

多个审批人中任意一人同意，该环节即通过，其余未处理的审批人自动跳过。

- **配置**：`flowlinks.concurrency_type = 2`
- **状态特征**：未审批的 Proc 状态被标记为 `4 (Skipped)`
- **用例**：领导出差时指定多位代理人，任一代理人审批即可

### 条件分支（按条件路由到不同节点）

根据表单提交的数据值，流程走向不同的审批路径。

- **配置**：`flowlinks.type = "Condition"`，`expression` 存储条件规则（JSON 格式）
- **校验安全**：操作符仅允许 `=`, `!=`, `>`, `<`, `>=`, `<=`, `like`, `in`, `not in`, `between`，使用参数化 SQL 查询防止注入
- **用例**：请假天数 ≤ 3 天主管审批即可；> 3 天需主管+经理+总经理三级审批

### 驳回到指定节点

审批人可以将流程驳回到任意之前的历史节点。

- **接口**：`POST /api/unpass`（上一节点）、`POST /api/unpassto`（指定节点，传入 targetProcessID）
- **状态变更**：目标节点的 Proc 恢复为 `Pending(0)`，中间节点的 Proc 标记为 `Skipped(4)`

### 加签（临时增加审批人）

当前审批人可以临时增加审批人参与当前环节的审批。

- **接口**：`POST /api/addsign`
- **数据表**：`proc_add_signs` 记录加签信息（加签人 ID、位置 before/after）

### 转交（将审批权转给他人）

当前审批人将自己的审批任务完全转交给另一个人处理。

- **接口**：`POST /api/transfer`
- **状态变更**：原 Proc 状态变为 `3 (Transferred)`，新审批人生成新的 Pending Proc

### 撤回（发起人撤销流程）

流程发起人可以在流程尚未完成时撤回自己的申请。

- **接口**：`POST /api/entry/revoke` 或 `POST /api/revoke`
- **状态变更**：Entry 状态变为 `Revoked(-2)`，所有 Pending Proc 同步变为 `Revoked(-2)`
- **校验**：仅发起人本人可撤回，且 Entry 状态必须为 Pending(0)，所有 Proc 均未被处理

### 子流程（主流程嵌套子流程）

主流程的某个环节触发一个独立的子流程，子流程完成后主流程继续。

- **配置**：`processes.position = 2`，`child_flow_id` 指向子流程 Flow ID
- **父子联动**：ChildAfter = 1（同时结束父流程），ChildAfter = 2（返回父流程指定节点）
- **用例**：采购申请主流程中超过一定金额自动触发"招标评审"子流程

### 抄送

流程的某个环节审批完成后，自动通知相关人员知悉，无需审批。

- **配置**：`processes.cc_emp_ids` 存储抄送人 ID（逗号分隔）
- **数据表**：`cc_records` 记录抄送详情（status: 0=未读, 1=已读）
- **触发时机**：审批通过后调用 `Transfer()` 方法时自动触发

### 超时自动处理

审批人在规定时间内未处理，系统自动执行预设动作。

- **配置**：`processes.limit_time` 字段（单位：秒）
- **执行方式**：Artisan 命令 `go run . artisan workflow:timeout-check`，建议配置系统 Cron 定期执行
- **用例**：普通报销单 3 天内主管未审批则自动通过

## 3. 审批人计算逻辑

审批人通过 Flowlink 表的 ApproverRule 字段定义，按优先级依次查找：

**优先级**：Sys（系统角色）> Emp（指定员工）> Dept（指定部门）

### Sys — 系统角色

| Auditor 值 | 含义 | 计算方式 |
|------------|------|----------|
| -1000 | 发起人自己 | 取 Entry.EmpID |
| -1001 | 部门主管 | 取 Entry.Emp.Dept.DirectorID |
| -1002 | 部门经理 | 取 Entry.Emp.Dept.ManagerID |
| -1003 | 表单指定 | 从 EntryData 中读取 ApproverRule 指定的字段值 |
| -1004 | 动态表达式 | 根据 ApproverRule 配置的映射键动态计算 |

### Emp — 指定员工

`flowlinks.type = "Emp"`，ApproverRule 存储逗号分隔的员工 ID 列表。

### Dept — 指定部门

`flowlinks.type = "Dept"`，ApproverRule 存储逗号分隔的部门 ID 列表，自动取各部门的 DirectorID 作为审批人。

最终结果经过 `uniqueSlice()` 去重后返回。

## 4. 事务安全说明

| 操作 | 事务范围 | 说明 |
|------|----------|------|
| 创建 Entry | 完整事务 | 同时创建 Entry + EntryData + 首个 Proc |
| 审批通过 | 非事务包裹 | 各子操作独立执行，通过状态机保证流转正确性 |
| 驳回 | 完整事务 | 更新 Proc 状态 + 回退目标 Proc 状态 |
| 撤回 | 完整事务 | 批量更新 Entry 和所有 Pending Proc 状态 |
| 加签 | 完整事务 | 创建 ProcAddSign 记录 + 新建 Proc |
| 转交 | 完整事务 | 创建新 Proc + 标记原 Proc 为已转交 |
| 初始化审批链 | 完整事务 | 计算审批人 + 批量创建 Procs + 更新 Entry.process_id |

## 5. 状态机转换规则

### Entry 合法转换

| 当前状态 | 合法转换 | 触发操作 |
|----------|----------|----------|
| 0 (Pending) | → 9 (Completed) | 最后一步审批通过 |
| 0 (Pending) | → -1 (Rejected) | 被驳回 |
| 0 (Pending) | → -2 (Revoked) | 发起人撤回 |
| -1 (Rejected) | → 0 (Pending) | 重新发起（Resend） |

### Proc 合法转换

| 当前状态 | 合法转换 | 触发操作 |
|----------|----------|----------|
| 0 (Pending) | → 1 (Approved) | 审批通过 |
| 0 (Pending) | → -1 (Rejected) | 审批驳回 |
| 0 (Pending) | → -2 (Revoked) | 流程撤回 |
| 0 (Pending) | → 3 (Transferred) | 转交他人 |
| 0 (Pending) | → 4 (Skipped) | 或签跳过 / 驳回跳过 |
| 0 (Pending) | → 9 (Consensus) | 会签中第一个审批人自动通过 |

---

## License

The Goravel Workflow package is open-sourced software licensed under the [MIT license](https://opensource.org/licenses/MIT).

---
<span id="custom-target"></span>
## 前言：老板的一天

### 故事开始

周一早上，你刚坐到工位上，老板给你发来一条消息：

> "小王啊，咱们那个 OA 系统太简陋了，员工请假就是提交个表单，主管看了一眼就完了。现在公司大了，业务复杂了，我要求支持多种审批场景——有的要多人会签，有的要根据金额自动分流，还要能加签、转交、撤回……你找个方案，尽快弄上来。"

你叹了口气，打开电脑，开始寻找解决方案。

---

### 第一步：最简单的请假流程

你决定从最简单的场景开始——**请假申请**。

张三打开系统，填写请假申请表：姓名、请假类型（年假/事假/病假）、请假天数、起止时间。提交后，流程自动发起。

这就是**流程实例（Entry）**的诞生。在系统中，一个完整的流程由以下核心概念组成：

| 概念 | 表名 | 说明 |
|------|------|------|
| **Flow（流程）** | `flows` | 业务流程定义，如"请假流程"、"报销流程" |
| **Process（环节）** | `processes` | 流程中的单个审批节点，如"主管审批"、"经理审批" |
| **Flowlink（连线）** | `flowlinks` | 环节之间的流转关系，包含条件分支和审批人规则 |
| **Entry（流程实例）** | `entries` | 一次具体的流程发起，如"张三的请假申请" |
| **Proc（审批任务）** | `procs` | Entry 在每个 Process 上产生的具体审批任务 |

**关系链**：Flow → Process → Flowlink → Entry → Proc

首先，你在后台画好流程图：张三提交 → 主管审批 → 经理审批 → 完成。

这就是**依次审批（顺序审批）**：审批人按固定顺序依次审批，前一个人批准后自动流转到下一个人。

配置方式很简单——在 Flowlink 表中设置 `next_process_id` 指定下一步环节 ID，`concurrency_type = 0`（默认值）。

张三提交请假申请后，系统自动创建 Entry（状态 Pending=0），并在"主管审批"环节生成 Proc 任务，推送给李四（张三的主管）。

李四收到通知，打开待办列表，看到张三的请假申请。他点击"通过"，流程自动流转到下一个环节——"经理审批"，生成新的 Proc 任务推送给王五（部门经理）。

王五也点击了"通过"，流程结束，Entry 状态变为 Completed(9)。

一切顺利。但你心里清楚，公司的需求远不止这么简单。

---

### 第二步：老板说，三天以上的请假要总经理审批

周三，老板又来找你了：

> "对了，张三那种请一天假的简单流程没问题。但如果有人请五天长假呢？得让总经理也审一遍。还有，报销也是——金额小的部门经理批就行，超过一万必须财务总监也签。"

你点点头，打开了**条件分支**的配置界面。

在 Flowlink 表中，你设置了 Type = "Condition" 的记录，Expression 字段存储条件表达式（JSON 格式）：

- 请假天数 ≤ 3：主管审批 → 完成
- 请假天数 > 3：主管审批 → 经理审批 → 总经理审批 → 完成
- 报销金额 ≤ 10000：部门经理审批 → 完成
- 报销金额 > 10000：部门经理审批 → 财务总监审批 → 完成

系统根据表单数据自动判断流程走向。操作符仅允许 `=`、`!=`、`>`、`<`、`>=`、`<=`、`like`、`in`、`not in`、`between`，字段名经正则校验，使用参数化 SQL 查询防止注入。

---

### 第三步：老板说，这个项目需要三个部门一起签字

周五，项目经理老陈跑来找你：

> "我们那个新项目立项，不能我一个人说了算。技术部、财务部、法务部，三个部门负责人都得同意才能立项。但也不是谁先签谁说了算——必须三个人全部签完，立项才算通过。"

这就是**会签（多人全部同意才通过）**。

你在 Process 的 Flowlink 中设置了多个审批人，ConcurrencyType = 1。

技术总监赵六、财务总监孙七、法务总监周八同时收到了审批任务。赵六先点了"通过"，系统记录他的审批，但流程没有继续——因为孙七和周八还没表态。

两天后，孙七也点了"通过"。此时已有两人同意，但仍不足三人。

又过了一天，周八终于点了"通过"。会签完成，三个审批人全部同意，流程进入下一环节。

如果期间周八点了"驳回"，那么整个会签失败，立项申请被驳回。

会签中的 Proc 状态为 `9 (Consensus)`，表示部分人已审批通过但尚未全部完成。

---

### 第四步：老板说，出差的时候审批不能卡住

下周一，老板出差了，电话里说：

> "我这两天在外地开会，有笔五千块的采购费等着我批。你让我的副手替我审批，我回来再补签就行。"

你理解了——这是**转交（将审批权转给他人）**。

老板调用 POST `/api/transfer`，将自己的审批任务转交给副手。原 Proc 状态变为 `3 (Transferred)`，副手收到新的 Pending 审批任务。

---

### 第五步：员工说，这个审批人我不认识，帮我加个人

周三，李四在审批一笔报销时遇到了难题：

> "这笔费用涉及海外市场推广，我不懂这块的业务，能不能让市场部的总监也看看？"

你给他加了**加签（临时增加审批人）**功能。

李四调用 POST `/api/addsign`，添加市场部总监作为额外审批人。`proc_add_signs` 表记录加签信息（加签人 ID、加签位置 before/after）。新增的 Proc 与原有审批人按并发规则执行——李四和市场总监都通过后，报销流程继续。

---

### 第六步：张三说，我日期填错了，撤回重填

周四上午，张三急匆匆发来消息：

> "糟了！我昨天提交的请假申请，日期写错了！主管还没批呢，能撤回吗？"

你告诉他可以。张三调用 POST `/api/entry/revoke`，成功撤回了流程。

Entry 状态变为 `Revoked(-2)`，所有 Pending 状态的 Proc 同步变为 `Revoked(-2)`。

张三修改日期后重新提交，Entry 状态恢复为 `Pending(0)`，流程重新开始。

注意：撤回有严格的校验规则——仅发起人本人可撤回，且 Entry 状态必须为 Pending(0)，所有 Proc 均未被处理（auditor_id = 0）。

---

### 第七步：老板说，审批完了也得让其他人知道啊

月底，财务总监审批通过了一笔大额报销。但人事部和行政部也需要知道这笔费用——人事部要做薪酬备案，行政部要做费用统计。

你配置了**抄送**功能。

在 Process 表中设置 CcEmpIDs 字段，填入人事部和行政部相关员工的 ID。审批通过后，系统自动调用 `Transfer()` 方法触发抄送，`cc_records` 表记录抄送详情（status: 0=未读, 1=已读）。

人事部和小赵收到了抄送通知："您关注的一笔大额报销已通过审批，请知悉。"

---

### 第八步：老陈说，审批太慢了，超时要自动处理

季度末，采购流程积压严重。老陈抱怨：

> "有些小额采购单卡在主管那里好几天没人批，效率太低了。能不能设个时限，超时自动通过？"

你配置了**超时自动处理**。

在 Process 表中设置 LimitTime 字段（单位：秒）。普通采购单设为 3 天（259200 秒），紧急事项设为 24 小时（86400 秒）。

系统通过 Artisan 命令 `go run . artisan workflow:timeout-check` 定时检查，建议配置系统 Cron 定期执行。超时后 Proc 标记为 Rejected(-1)，Entry 同步标记为 Rejected(-1)。

你也可以配置为超时自动通过——在超时检查逻辑中修改状态为 Approved(1) 即可。

---

### 第九步：老板说，驳回到上一步太麻烦了

财务总监审批一笔报销时发现了问题：

> "这张发票的附件不全，材料是从主管环节就开始缺的，为什么要我驳回到部门经理，再由部门经理驳回到主管？直接驳回到主管不是更快？"

你实现了**驳回到指定节点**。

财务总监调用 POST `/api/unpassto`，传入 targetProcessID 参数，直接将流程驳回到"主管审批"环节。目标节点的 Proc 恢复为 `Pending(0)` 状态，中间节点的 Proc 被标记为 `Skipped(4)`。

主管重新收到任务，补齐材料后再次提交。

---

### 第十步：老板说，不同业务要走不同的流程

半年后，公司业务扩张，除了请假和报销，还新增了"采购申请"、"合同审批"、"人事任免"等多种业务流程。

你发现每个流程的结构类似，但审批路径不同。于是你设计了**子流程**机制。

比如"采购申请"主流程中，超过 10 万金额的采购自动触发"招标评审"子流程。子流程完成后，根据 ChildAfter 配置决定主流程走向：
- ChildAfter = 1：子流程完成后，父流程也标记为完成
- ChildAfter = 2：子流程完成后返回父流程指定节点

配置方式：在 Process 表中设置 Position = 2（子流程入口），ChildFlowID 指向子流程的 Flow ID。

---

### 第十一步：老板说，审批人不能手动指定，要根据组织架构自动算

随着公司规模扩大，员工来来去去，每次审批流程都要手动指定审批人，维护成本太高。

你实现了**审批人自动计算**机制。

每个 Process 通过 Flowlink 表的 ApproverRule 字段定义审批人，支持三种类型，按优先级依次查找：

**优先级**：Sys（系统角色）> Emp（指定员工）> Dept（指定部门）

#### Sys — 系统角色自动计算

| Auditor 值 | 含义 | 计算方式 |
|------------|------|----------|
| -1000 | 发起人自己 | 取 Entry.EmpID |
| -1001 | 部门主管 | 取 Entry.Emp.Dept.DirectorID |
| -1002 | 部门经理 | 取 Entry.Emp.Dept.ManagerID |
| -1003 | 表单指定 | 从 EntryData 中读取 ApproverRule 指定的字段值作为员工 ID |
| -1004 | 动态表达式 | 根据 ApproverRule 配置的映射键动态计算 |

例如：报销流程的第一个审批人设为 Auditor = -1001（部门主管），系统会自动查找当前发起人的直属主管，无论主管是谁、换了什么人，流程都能正确流转。

#### Emp — 指定员工

Flowlink.Type = "Emp"，ApproverRule 存储逗号分隔的员工 ID 列表（如 `"1,3,5"`）。

#### Dept — 指定部门

Flowlink.Type = "Dept"，ApproverRule 存储逗号分隔的部门 ID 列表，自动取各部门的 DirectorID 作为审批人。

最终结果经过 `uniqueSlice()` 去重后返回。

---


---

### 第十二步：老板说，审批的时候记得通知我啊

流程跑了一段时间后，老板又提了个要求：

> "每次有人提交申请或者轮到我审批的时候，你得提醒我一下啊。总不能让我天天盯着系统看吧？"

你笑了——这是**通知 Hook** 的功能。

在实现 Hook 接口时，你实现了两个方法：

```go
// 通知发起人，在被驳回时调用，或者整个流程结束时调用。
func (u *User) NotifySendOne(id uint) error {
// 发送邮件、短信等通知
return nil
}

// 通知下一个审批人，当当前环节的审批人通过时触发。
func (u *User) NotifyNextAuditor(id uint) error {
// 通知下一审批人
return nil
}
```

现在张三提交了请假申请，李四（主管）的手机立刻收到一条通知："您有一条新的审批待办——张三的请假申请，请及时处理。"

李四点击"通过"后，王五（经理）也收到了通知："您有一条新的审批待办——张三的请假申请，请及时处理。"

流程完成后，系统自动调用 `NotifySendOne`，张三收到邮件："您的请假申请已通过所有审批，流程已结束。"

如果流程被驳回，张三也会收到通知："您的请假申请已被驳回，请查看审批意见后重新提交。"

通知方式可以自行扩展——邮件、短信、企业微信、钉钉、飞书，全由开发者自己实现。

---

