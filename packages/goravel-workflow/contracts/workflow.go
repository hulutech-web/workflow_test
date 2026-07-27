package contracts

// Workflow 工作流核心接口，定义了审批流程引擎的契约规范。
// 当前为空接口，具体的方法签名由实现类（services/workflow/workflow.go）定义。
// 主要能力包括：审批流转（Transfer/Pass）、驳回（UnPass/UnPassTo）、
// 撤回（Revoke）、加签（AddSign）、转交（TransferProc）、评论（AddComment/GetComments）、
// 钩子注册（RegisterHook）以及插件执行（ExecPluginMethod）等。
type Workflow interface {}
