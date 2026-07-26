<p align="center">
  <img src="https://goravel/packages/goravel-workflow/blob/master/assets/workflow.png?raw=true" width="300" />
</p>

### 演示

<p align="center">
  <img src="https://goravel/packages/goravel-workflow/blob/master/assets/flow_preview.png?raw=true" width="800" />
</p>

### 在线演示
[地址](http://workflow.xiaohongpao.top/#/auth/login)

### 文档

使用手册请访问[文档](https://hulutech-web.github.io/goravel-workflow.github.io/)


### 一、安装
```shell
go get  goravel/packages/goravel-workflow
```
#### 1.1 注册服务提供者:config/app.go
```shell

import	"goravel/packages/goravel-workflow"
```

#### 1.2 注册服务提供者:config/app.go
```go
func init() {
"providers": []foundation.ServiceProvider{
	....
	&workflow.ServiceProvider{},
}
}
```
### 二、发布资源，默认将发布2类资源，一是配置文件，而是数据表迁移
#### 2.1 发布资源:config/app.go
```shell
go run . artisan vendor:publish --package=goravel/packages/goravel-workflow
```

可选标签：
- `--tag=migrations` — 只发布迁移文件
- `--tag=seeders` — 只发布种子数据
- `--tag=config` — 只发布配置文件
- 不指定标签则全部发布
#### 2.2 发布迁移文件:database

```shell
artisan vendor:publish --package=goravel/packages/goravel-workflow --tag=migrations
```

发布后得到两个 Go 迁移文件：
- `2024_06_24_000000_create_workflow_base_tables.go` — 创建工作流专用表（flows、flowtypes、processes、processvars、templates、templateforms、entries、entrydatas、procs、attachments、products，共 11 张表）。注：users、depts、emps 表由宿主应用提供，不包含在此迁移中。
- `2026_07_23_000000_add_workflow_features.go` — 添加新增功能字段（concurrency_type、approver_rule、unpass_target_process_id、cc_emp_ids）及新建表（proc_add_signs、proc_comments、cc_records）

#### 2.3 执行迁移建表
在database/seeders/database_seeder.go下的添加
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
如果启动项目报错，请检查路由是否有重名，并修改路由
#### 2.6 模型映射
发布资源后，config/workflow.go中的配置文件中有默认的关联映射，根据需要自行修改和修改
### 三、实现Hook接口（可选）
用户自定义User结构中注入流程框架，并实现框架中的Hook接口
```go
type User struct {
	orm.Model
	Name     string `gorm:"column:name;type:varchar(255);not null" form:"name" json:"name"`
	WorkNo   string `gorm:"column:workno;not null;unique_index:users_workno_unique" json:"workno" form:"workno"`
	Password string `gorm:"column:password;type:varchar(255);not null" form:"password" json:"password"`
	...
	Workflow *Workflow
	orm.SoftDeletes
}
```
实现接口
```go
// 通知发起人，在被驳回时调用，或者整个流程结束时调用。
func (u *User) NotifySendOne(id uint) error {

	fmt.Printf("custom ======User %d unpasshook called.\n", id)
	return nil
}

// 通知下一个审批人，当当前环节的审批人通过时，触发。
func (u *User) NotifyNextAuditor(id uint) error {
	fmt.Printf("custom ======User %d passhook called.\n", id)
	return nil
}

```

### 实例化workflow
框架提供了2个``hooks``，供开发者自行实现逻辑，可以发送邮件通知，短信通知等
``app/providers/app_services_provider.go``
实例化workflow，并注入服务
```go
func (receiver *AppServiceProvider) Boot(app foundation.Application) {
	wf := workflow.NewBaseWorkflow()
	// 注册子级的方法到工作流中
	user := &models.User{Workflow: wf}
	wf.RegisterHook("NotifySendOneHook", reflect.ValueOf(user.NotifySendOne))
	wf.RegisterHook("NotifyNextAuditorHook", reflect.ValueOf(user.NotifyNextAuditor))
}

回调参数将在User结构中的NotifySendOne和NotifyNextAuditor方法中执行后续操作，由开发者自行实现

```
### 四、框架路由说明
```go
package routes

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/facades"
	controllers "goravel/packages/goravel-workflow/controllers"
	"goravel/packages/goravel-workflow/middleware"
)

func Api(app foundation.Application) {
	router := app.MakeRoute()

	authController := controllers.NewAuthController()
	router.Post("/api/auth/login", authController.AdminLogin)
	router.Post("/api/h5/login", authController.H5Login)
	captchaController := controllers.NewCaptchaController()
	router.Get("/api/captcha/get", captchaController.GetCaptcha)
	router.Post("/api/captcha/validate", captchaController.ValidateCaptcha)

	facades.Route().Middleware(middleware.Jwt()).Prefix("/api").Group(func(router route.Router) {

		//文件上传
		uploadCtrl := controllers.NewUploadController()
		router.Post("/upload", uploadCtrl.Upload)

		homeCtrl := controllers.NewHomeController()
		router.Get("/home", homeCtrl.Index)

		//	部门
		deptCtrl := controllers.NewDeptController()
		router.Resource("dept", deptCtrl)
		router.Post("dept/bindmanager", deptCtrl.BindManager)
		router.Post("dept/binddirector", deptCtrl.BindDirector)

		//	员工
		empCtrl := controllers.NewEmpController()
		router.Resource("emp", empCtrl)
		router.Post("emp/search", empCtrl.Search)
		router.Get("emp/options", empCtrl.Options)
		router.Post("emp/bind", empCtrl.BindUser)
		//流程
		flowCtrl := controllers.NewFlowController()
		router.Resource("flow", flowCtrl)
		router.Get("flow/list", flowCtrl.List)
		router.Get("flow/create", flowCtrl.Create)
		//流程设计
		router.Get("flow/flowchart/{id}", flowCtrl.FlowDesign)
		router.Post("flow/publish", flowCtrl.Publish)

		//entry节点
		entryCtrl := controllers.NewEntryController()
		router.Get("flow/{id}/entry", entryCtrl.Create)
		router.Post("entry", entryCtrl.Store)
		router.Get("entry/{id}", entryCtrl.Show)
		router.Get("entry/{id}/entrydata", entryCtrl.EntryData)
		//流程重发
		router.Post("entry/resend", entryCtrl.Resend)
		//撤回流程
		router.Post("entry/revoke", entryCtrl.Revoke)
		//流程轨迹
		flowlinkCtrl := controllers.NewFlowlinkController()
		router.Post("flowlink", flowlinkCtrl.Update)
		//模板控件
		templateformCtrl := controllers.NewTemplateformController()
		router.Get("template/{id}/templateform", templateformCtrl.Index)
		router.Post("templateform", templateformCtrl.Store)
		router.Put("templateform/{id}", templateformCtrl.Update)
		router.Delete("templateform/{id}", templateformCtrl.Destroy)
		router.Get("templateform/{id}", templateformCtrl.Show)
		//模板
		templateCtrl := controllers.NewTemplateController()
		router.Resource("template", templateCtrl)

		//	流程
		processCtrl := controllers.NewProcessController()
		router.Resource("process", processCtrl)
		router.Get("process/attribute", processCtrl.Attribute)
		router.Post("process/con", processCtrl.Condition)
		router.Post("process/list", processCtrl.List)

		//	审批流转
		procCtrl := controllers.NewProcController()
		router.Get("proc/{entry_id}", procCtrl.Index)
		//同意
		router.Post("pass", procCtrl.Pass)
		//驳回
		router.Post("unpass", procCtrl.UnPass)
		//撤回
		router.Post("revoke", procCtrl.Revoke)
		//加签
		router.Post("addsign", procCtrl.AddSign)
		//转交
		router.Post("transfer", procCtrl.TransferProc)
		//评论
		router.Post("comment", procCtrl.AddComment)
		router.Get("comments/{entry_id}", procCtrl.GetComments)

		//抄送
		ccCtrl := controllers.NewCcController()
		router.Get("cc/list", ccCtrl.List)
		router.Get("cc/entry/{entry_id}", ccCtrl.GetEntryCC)
	})
}

```

### 五、新增功能说明

| 功能 | 说明 | API |
|------|------|-----|
| **撤回** | 发起人撤回未处理的流程 | `POST /api/entry/revoke` / `POST /api/revoke` |
| **加签** | 当前审批人可添加额外审批人 | `POST /api/addsign` |
| **转交** | 将审批任务转交给其他员工 | `POST /api/transfer` |
| **评论** | 多轮对话式评论/留言 | `POST /api/comment` / `GET /api/comments/{entry_id}` |
| **抄送** | 节点完成后自动抄送相关人员 | `GET /api/cc/list` / `GET /api/cc/entry/{entry_id}` |
| **超时检查** | 定时检查超时的待办任务 | `go run . artisan workflow:timeout-check` |
| **会签** | 同一步骤所有审批人都需通过，全部通过才进入下一步 | Flowlink.ConcurrencyType = 1 |
| **或签** | 同一步骤一人通过即进入下一步，其余审批人自动跳过 | Flowlink.ConcurrencyType = 2 |
| **驳回到指定节点** | 驳回到任意历史节点，而非仅上一步 | ProcController.UnPass 支持 targetProcessID 参数 |
| **表单指定审批人** | 从表单字段动态读取审批人 ID | Flowlink.Auditor = -1003, ApproverRule 存储字段名 |
| **动态表达式审批人** | 通过表达式规则计算审批人 | Flowlink.Auditor = -1004, ApproverRule 存储映射键 |

#### 状态码说明
- `entries.status`: -2 = 已撤销
- `procs.status`: 3 = 已转交, 4 = 已跳过（或签）, 9 = 会签通过

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
| -1003 | 从表单字段读取审批人（ApproverRule 存储字段名） |
| -1004 | 动态表达式计算审批人（ApproverRule 存储映射键） |

### 六、前端集成
请访问前端框架[goravel-workflow-vue](https://goravel/packages/goravel-workflow-vue)下载安装扩展，并按照文档进行集成

### 七、接口文档
请访问前端框架[goravel-workflow-doc](https://goravel/packages/goravel-workflow-vuepress)进行查看

# # # 八、流程场景介绍

## 一、核心概念

| 概念 | 表名 | 说明 |
|------|------|------|
| **Flow（流程）** | `flows` | 业务流程定义，如"请假流程"、"报销流程"。包含流程名称、关联的表单模板、流程图 JSON 等。 |
| **Process（环节）** | `processes` | 流程中的单个审批节点，如"主管审批"、"经理审批"。每个环节配置审批人规则、并发方式、超时时间、抄送人等。 |
| **Flowlink（连线）** | `flowlinks` | 环节之间的流转关系，定义条件分支和审批人计算规则。 |
| **Entry（流程实例）** | `entries` | 一次具体的流程发起，如"张三的请假申请"。关联发起员工、所属流程、表单数据。 |
| **Proc（审批任务）** | `procs` | Entry 在每个 Process 上产生的具体审批任务，代表一个待办审批操作。 |

**关系链**：一个 Flow 包含多个 Process，Process 之间通过 Flowlink 连接。用户发起一个 Flow 时创建 Entry，Entry 在第一个 Process 上生成 Proc，审批通过后流转至下一个 Process 再生成新 Proc。

## 二、典型流程场景

### 1. 依次审批（顺序审批）

**场景描述**：审批人按固定顺序依次审批，前一个人批准后自动流转到下一个人。

**配置方式**：在 Flowlink 表中设置 Process 的连接关系，ConcurrencyType = 0（默认值）。

**典型用例**：请假申请 → 主管审批 → 经理审批 → 总经理审批 → 完成

**数据模型**：`flowlinks.next_process_id` 指定下一步环节 ID，`flowlinks.concurrency_type = 0`

---

### 2. 会签（多人全部同意才通过）

**场景描述**：多个审批人必须全部同意，流程才能进入下一环节。只要有一人驳回，整个会签失败。

**配置方式**：在 Flowlink 的 ApproverRule 中设置多个审批人，ConcurrencyType = 1。

**典型用例**：部门例会需要三位部门负责人全部签字确认；项目立项需要技术、财务、法务三方会签。

**数据模型**：`flowlinks.concurrency_type = 1`

**状态特征**：会签中的 Proc 状态为 `9 (Consensus)`，表示部分人已审批通过但尚未全部完成。当所有审批人均通过后自动进入下一步；若有人驳回则整个 Entry 被驳回。

---

### 3. 或签（一人同意即可通过）

**场景描述**：多个审批人中任意一人同意，该环节即通过，其余未处理的审批人自动跳过。

**配置方式**：在 Flowlink 的 ApproverRule 中设置多个审批人，ConcurrencyType = 2。

**典型用例**：领导出差时指定多位代理人，任一代理人审批即可；文档审核由多人并行审阅，一人通过即可。

**数据模型**：`flowlinks.concurrency_type = 2`

**状态特征**：未审批的 Proc 状态被标记为 `4 (Skipped)`。

---

### 4. 条件分支（按条件路由到不同节点）

**场景描述**：根据表单提交的数据值，流程走向不同的审批路径。

**配置方式**：在 Flowlink 表中设置 Type = "Condition" 的记录，Expression 字段存储条件表达式（JSON 格式）。

**典型用例**：
- 请假天数 ≤ 3 天：主管审批即可结束
- 请假天数 > 3 天：需主管 + 经理 + 总经理三级审批
- 报销金额 ≤ 1000：部门经理审批
- 报销金额 > 1000：部门经理 + 财务总监审批

**数据模型**：`flowlinks.type = "Condition"`，`flowlinks.expression` 存储条件规则

**校验安全**：操作符仅允许 `=`, `!=`, `>`, `<`, `>=`, `<=`, `like`, `in`, `not in`, `between`，字段名经正则 `^[a-zA-Z0-9_]+$` 校验，使用参数化 SQL 查询防止注入。

---

### 5. 驳回到指定节点

**场景描述**：审批人可以将流程驳回到任意之前的历史节点，而非只能驳回到上一节点。

**接口**：POST `/api/unpass`（驳回到上一节点）、POST `/api/unpassto`（驳回到指定节点，传入 targetProcessID 参数）

**典型用例**：总经理发现材料有问题，直接驳回到最初的主管环节重新审核。

**状态变更**：目标节点的 Proc 恢复为 `Pending(0)` 状态，中间节点的 Proc 被标记为 `Skipped(4)`。

---

### 6. 加签（临时增加审批人）

**场景描述**：当前审批人觉得某件事需要其他人一起判断，可以临时增加审批人参与当前环节的审批。

**接口**：POST `/api/addsign`

**典型用例**：主管审批时不确定某个专业问题，临时加技术总监一起审批。

**数据模型**：`proc_add_signs` 表记录加签信息（加签人 ID、加签位置 before/after）。新增的 Proc 与原有审批人按并发规则执行。

---

### 7. 转交（将审批权转给他人）

**场景描述**：当前审批人将自己的审批任务完全转交给另一个人处理。

**接口**：POST `/api/transfer`

**典型用例**：经理出差，将审批权限临时转交给副经理。

**状态变更**：原 Proc 状态变为 `3 (Transferred)`，新审批人生成新的 Pending Proc。

---

### 8. 撤回（发起人撤销流程）

**场景描述**：流程发起人可以在流程尚未完成时撤回自己的申请。

**接口**：POST `/api/entry/revoke` 或 POST `/api/revoke`

**典型用例**：张三提交请假申请后发现日期填错了，在主管审批前主动撤回修改后重新提交。

**状态变更**：Entry 状态变为 `Revoked(-2)`，所有 Pending 状态的 Proc 同步变为 `Revoked(-2)`。

**校验规则**：仅发起人本人可撤回，且 Entry 状态必须为 Pending(0)，所有 Proc 均未被处理（auditor_id = 0）。

---

### 9. 子流程（主流程嵌套子流程）

**场景描述**：主流程的某个环节触发一个独立的子流程，子流程完成后主流程继续。

**配置方式**：在 Process 表中设置 Position = 2（子流程入口），ChildFlowID 指向子流程的 Flow ID。ChildAfter 控制子流程完成后主流程的走向，ChildBackProcess 控制子流程可驳回到的节点。

**典型用例**：采购申请主流程中，超过一定金额自动触发"招标评审"子流程；人事任免主流程中嵌入"背景调查"子流程。

**数据模型**：`processes.position = 2`，`processes.child_flow_id`，`processes.child_after`（1=同时结束父流程，2=返回父流程），`processes.child_back_process`

**父子联动**：
- ChildAfter = 1：子流程完成后，父流程也标记为 Completed(9)
- ChildAfter = 2：子流程完成后返回父流程指定节点（由 ChildBackProcess 决定）

---

### 10. 抄送

**场景描述**：流程的某个环节审批完成后，自动通知相关人员知悉，但无需他们审批。

**配置方式**：在 Process 表中设置 CcEmpIDs 字段，填写需要抄送的员工 ID 列表（逗号分隔）。

**典型用例**：审批通过后抄送给 HR 备案；项目立项通过后抄送给财务部。

**数据模型**：`processes.cc_emp_ids` 存储抄送人 ID，`cc_records` 表记录抄送详情（含已读/未读状态 status: 0=未读, 1=已读）。

**触发时机**：审批通过后调用 `Transfer()` 方法时自动触发抄送。

---

### 11. 超时自动处理

**场景描述**：审批人在规定时间内未处理，系统自动执行预设动作。

**配置方式**：在 Process 表中设置 LimitTime 字段（单位：秒）。

**典型用例**：普通报销单 3 天内主管未审批则自动通过；紧急事项 24 小时未处理自动升级给上级主管。

**数据模型**：`processes.limit_time` 字段

**执行方式**：通过 Artisan 命令 `go run . artisan workflow:timeout-check` 定时检查，建议配置系统 Cron 定期执行。超时后将 Proc 标记为 Rejected(-1)，Entry 同步标记为 Rejected(-1)。

---

## 三、审批人计算逻辑

每个 Process 的审批人通过 Flowlink 表的 ApproverRule 字段定义，支持三种类型，按优先级依次查找：

**优先级**：Sys（系统角色）> Emp（指定员工）> Dept（指定部门）

### 1. Sys — 系统角色自动计算

| Auditor 值 | 含义 | 计算方式 |
|------------|------|----------|
| -1000 | 发起人自己 | 取 Entry.EmpID |
| -1001 | 部门主管 | 取 Entry.Emp.Dept.DirectorID |
| -1002 | 部门经理 | 取 Entry.Emp.Dept.ManagerID |
| -1003 | 表单指定 | 从 EntryData 中读取 ApproverRule 指定的字段值作为员工 ID |
| -1004 | 动态表达式 | 根据 ApproverRule 配置的映射键动态计算审批人 |
| 其他数值 | 指定 ID | 直接使用数值作为员工 ID |

### 2. Emp — 指定员工

Flowlink.Type = "Emp"，ApproverRule 存储逗号分隔的员工 ID 列表（如 `"1,3,5"`）。

### 3. Dept — 指定部门

Flowlink.Type = "Dept"，ApproverRule 存储逗号分隔的部门 ID 列表，自动取各部门的 DirectorID 作为审批人。

最终结果经过 `uniqueSlice()` 去重后返回。

---

## 四、状态码速查

### Entry 状态（entries.status）

| 状态码 | 常量名 | 含义 | 说明 |
|--------|--------|------|------|
| 0 | Pending | 进行中 | 流程刚发起，等待第一个审批人处理 |
| 9 | Completed | 已完成 | 所有审批节点均已通过 |
| -1 | Rejected | 已驳回 | 流程被驳回 |
| -2 | Revoked | 已撤销 | 发起人主动撤回 |
| -9 | Draft | 草稿 | 保存为草稿未提交 |

### Proc 状态（procs.status）

| 状态码 | 常量名 | 含义 | 说明 |
|--------|--------|------|------|
| 0 | Pending | 待审批 | 等待该审批人处理 |
| 1 | Approved | 已通过 | 审批人同意 |
| -1 | Rejected | 已驳回 | 审批人驳回 |
| -2 | Revoked | 已撤销 | 关联 Entry 被撤销 |
| 3 | Transferred | 已转交 | 审批权已转给他人 |
| 4 | Skipped | 已跳过 | 条件不满足或被或签跳过 |
| 9 | Consensus | 会签通过 | 会签模式中第一个审批人自动通过 |

---

## 五、事务安全说明

| 操作 | 事务范围 | 说明 |
|------|----------|------|
| 创建 Entry | 完整事务 | 同时创建 Entry + EntryData + 首个 Proc，保证原子性 |
| 审批通过 | 非事务包裹 | 各子操作独立执行，通过状态机保证流转正确性 |
| 驳回 | 完整事务 | 更新 Proc 状态 + 回退目标 Proc 状态 |
| 撤回 | 完整事务 | 批量更新 Entry 和所有 Pending Proc 状态 |
| 加签 | 完整事务 | 创建 ProcAddSign 记录 + 新建 Proc |
| 转交 | 完整事务 | 创建新 Proc + 标记原 Proc 为已转交 |
| 初始化审批链 | 完整事务 | 计算审批人 + 批量创建 Procs + 更新 Entry.process_id |

---

## 六、状态机转换规则

### Entry 合法转换

| 当前状态 | 合法转换 | 触发操作 |
|----------|----------|----------|
| 0 (Pending) | → 9 (Completed) | 最后一步审批通过 |
| 0 (Pending) | → -1 (Rejected) | 被驳回 |
| 0 (Pending) | → -2 (Revoked) | 发起人撤回 |
| 0 (Pending) | → 0 (Pending) | 审批通过，流转至下一步 |
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

