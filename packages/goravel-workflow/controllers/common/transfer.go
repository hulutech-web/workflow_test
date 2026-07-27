package common

// Plumb 流程流转数据结构，用于表示工作流中步骤之间的连接关系。
// 包含节点总数和以节点ID为键的节点映射表。
type Plumb struct {
	Total int             `json:"total"` // 节点总数
	List  map[string]Node `json:"list"`  // 节点列表，key为节点标识，value为节点信息
}

// Node 流程节点结构体，表示工作流中的一个审批步骤节点。
type Node struct {
	ID          int    `json:"id"`           // 节点唯一标识
	FlowId      int    `json:"flow_id"`      // 所属流程定义ID
	ProcessName string `json:"process_name"` // 步骤名称（如"部门主管审批"）
	ProcessTo   string `json:"process_to"`   // 目标步骤标识，用于前端流程图画线连接
	Icon        string `json:"icon"`         // 节点图标样式
	Style       string `json:"style"`        // 节点CSS样式（如颜色、尺寸等）
}

// ProcessCondition 流程条件规则结构体，定义条件分支路由的判断逻辑。
// 用于在工作流流转时，根据表单数据动态选择下一步骤。
type ProcessCondition struct {
	Id         int    `json:"id"`          // 条件记录唯一标识
	Index      int    `json:"index"`       // 条件序号，用于排序多个条件的求值顺序
	Field      string `json:"field"`       // 条件字段名，对应表单中的字段
	Operator   string `json:"operator"`    // 条件运算符（如 =, !=, >, <, >=, <=, like, in, not in, between）
	Value      string `json:"value"`       // 条件值，用于与表单字段值进行比较
	Extra      string `json:"extra"`       // 扩展字段名，用于范围运算（如 between 的第二个字段）
	ExtraValue string `json:"extra_value"` // 扩展条件值，用于范围运算（如 between 的上限值）
}

// ProcessRequest 创建/更新流程步骤的请求结构体。
// 包含步骤的基本信息、审批人配置、条件规则、样式设置等完整定义。
type ProcessRequest struct {
	ProcessName      string             `json:"process_name"`      // 步骤名称
	ProcessPosition  int                `json:"process_position"`   // 步骤位置：0=首步骤，1=普通步骤，2=进入子流程
	ChildFlowId      int                `json:"child_flow_id"`      // 子流程定义ID（当 position=2 时使用）
	ChildAfter       int                `json:"child_after"`        // 子流程完成后的行为：1=同时结束父流程，2=返回父流程继续
	ChildBackProcess int                `json:"child_back_process"` // 返回父流程时跳转到的目标步骤ID（ChildAfter=2 时生效）
	AutoPerson       string             `json:"auto_person"`        // 自动审批人配置（JSON格式，如指定特定员工或规则）
	RangeEmpIds      []int              `json:"range_emp_ids"`      // 审批人范围-员工ID列表
	RangeEmpText     []string           `json:"range_emp_text"`     // 审批人范围-员工姓名列表（前端展示用）
	RangeDeptIds     []int              `json:"range_dept_ids"`     // 审批人范围-部门ID列表
	RangeDeptText    []string           `json:"range_dept_text"`    // 审批人范围-部门名称列表（前端展示用）
	ProcessMode      string             `json:"process_mode"`       // 审批模式（如单人审批、多人审批等）
	ProcessCondition []ProcessCondition `json:"process_condition"`  // 条件分支规则列表，支持多条件路由
	StyleWidth       int                `json:"style_width"`        // 节点样式-宽度（像素）
	StyleHeight      int                `json:"style_height"`       // 节点样式-高度（像素）
	StyleColor       string             `json:"style_color"`        // 节点样式-背景颜色
	StyleIcon        string             `json:"style_icon"`         // 节点样式-图标
	ConcurrencyType  int                `json:"concurrency_type"`   // 并发审批类型：0=依次审批，1=会签（全部通过），2=或签（一人通过即可）
	ApproverRule     string             `json:"approver_rule"`      // 审批人规则：当 Auditor=-1003 时为表单字段名，-1004 时为表达式映射键
	LimitTime        int                `json:"limit_time"`         // 审批时限（秒），超时未处理将自动驳回
	CcEmpIDs         string             `json:"cc_emp_ids"`         // 抄送人ID列表（逗号分隔），审批完成后自动生成抄送记录
}
