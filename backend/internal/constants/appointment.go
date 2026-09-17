package constants

// 预约状态机状态
const (
	AppointmentStatePending       = "pending"        // 待确认：仅发起人可撤回
	AppointmentStateConfirmed     = "confirmed"      // 双方已确认：锁定，只能走取消流程
	AppointmentStateCancelPending = "cancel_pending" // 已提交取消原因，等待对方处理
	AppointmentStateWithdrawn     = "withdrawn"      // 已撤回（终态，时间档已释放）
	AppointmentStateCancelled     = "cancelled"      // 已取消（终态，时间档已释放）
)

// 预约状态展示标签（保留原有中文文案，确保旧数据可读）
const (
	AppointmentLabelPending       = "等待对方确认"
	AppointmentLabelConfirmed     = "双方已确认"
	AppointmentLabelCancelPending = "取消待对方处理"
	AppointmentLabelWithdrawn     = "已撤回"
	AppointmentLabelCancelled     = "已取消"
)

// AppointmentStateLabels 状态机状态到展示标签的映射
var AppointmentStateLabels = map[string]string{
	AppointmentStatePending:       AppointmentLabelPending,
	AppointmentStateConfirmed:     AppointmentLabelConfirmed,
	AppointmentStateCancelPending: AppointmentLabelCancelPending,
	AppointmentStateWithdrawn:     AppointmentLabelWithdrawn,
	AppointmentStateCancelled:     AppointmentLabelCancelled,
}

// 取消申请处理状态
const (
	CancelStatusPending  = "pending"
	CancelStatusAccepted = "accepted"
	CancelStatusRejected = "rejected"
)

// 取消申请处理动作
const (
	CancelActionAccept = "accept"
	CancelActionReject = "reject"
)

// 预约相关错误码
const (
	ErrCodeActorRequired         = "APPOINTMENT_ACTOR_REQUIRED"
	ErrCodeAppointmentNotFound   = "APPOINTMENT_NOT_FOUND"
	ErrCodeNotParty              = "APPOINTMENT_NOT_PARTY"
	ErrCodeOnlyInitiatorWithdraw = "APPOINTMENT_ONLY_INITIATOR_WITHDRAW"
	ErrCodeLocked                = "APPOINTMENT_LOCKED"
	ErrCodeTerminal              = "APPOINTMENT_TERMINAL"
	ErrCodeCancelInProgress      = "APPOINTMENT_CANCEL_IN_PROGRESS"
	ErrCodeCancelReasonRequired  = "APPOINTMENT_CANCEL_REASON_REQUIRED"
	ErrCodeCancelAlreadyPending  = "APPOINTMENT_CANCEL_ALREADY_PENDING"
	ErrCodeNoCancelRequest       = "APPOINTMENT_NO_CANCEL_REQUEST"
	ErrCodeNotCancelResponder    = "APPOINTMENT_NOT_CANCEL_RESPONDER"
	ErrCodeInvalidCancelAction   = "APPOINTMENT_INVALID_CANCEL_ACTION"
	ErrCodeStoreSaveFailed       = "APPOINTMENT_STORE_SAVE_FAILED"
)

// 预约存储默认文件路径（可用 APPOINTMENT_STORE_PATH 覆盖）
const DefaultAppointmentStorePath = "data/appointments.json"
