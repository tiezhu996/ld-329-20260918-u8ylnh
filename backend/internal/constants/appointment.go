package constants

// 交换预约状态机：
//
//	待确认 pending ──任一方确认──▶ 单方已确认 confirmed_partial
//	                             └──另一方确认──▶ 双方已确认 confirmed
//	pending / confirmed_partial ──发起人撤回──▶ 已撤回 withdrawn（终态，时间档释放）
//	confirmed ──任一方提交取消原因──▶ 取消待处理 cancel_requested
//	                            └──对方同意──▶ 已取消 cancelled（终态，时间档释放）
//	                            └──对方拒绝──▶ confirmed（回到生效态）
const (
	ApptStatusPending          = "pending"
	ApptStatusConfirmedPartial = "confirmed_partial"
	ApptStatusConfirmed        = "confirmed"
	ApptStatusWithdrawn        = "withdrawn"
	ApptStatusCancelRequested  = "cancel_requested"
	ApptStatusCancelled        = "cancelled"
)

// 角色。
const (
	RoleInitiator = "initiator"
	RoleResponder = "responder"
)

// 取消请求处理结果。
const (
	CancelDecisionApprove = "approve"
	CancelDecisionReject  = "reject"
)

// 可用于发起预约的地点类型。
const (
	PlaceOnline  = "线上"
	PlaceOffline = "线下"
)

// 字段长度限制，统一从这里取，避免魔法数字。
const (
	MaxUserLength      = 80
	MaxPlaceLength     = 120
	MaxTimeLength      = 80
	MaxAgendaLength    = 500
	MaxReasonLength    = 300
	MinCancelReasonLen = 4
)

// 终态集合：终态不接受任何再次确认 / 撤回 / 取消。
var AppointmentTerminalStatuses = map[string]struct{}{
	ApptStatusWithdrawn: {},
	ApptStatusCancelled: {},
}

// 状态对应的中文展示文案（旧前端只读 statusText 字段）。
var AppointmentStatusText = map[string]string{
	ApptStatusPending:          "待确认",
	ApptStatusConfirmedPartial: "单方已确认",
	ApptStatusConfirmed:        "双方已确认",
	ApptStatusWithdrawn:        "已撤回",
	ApptStatusCancelRequested:  "取消待对方处理",
	ApptStatusCancelled:        "已取消",
}
