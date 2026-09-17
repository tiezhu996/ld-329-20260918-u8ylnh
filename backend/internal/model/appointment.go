package model

// CancelRequest 记录一次取消申请：确认锁定后只能提交取消原因，等待对方处理。
type CancelRequest struct {
	By        string `json:"by"`
	Reason    string `json:"reason"`
	Status    string `json:"status"` // pending / accepted / rejected
	CreatedAt string `json:"createdAt"`
	HandledBy string `json:"handledBy,omitempty"`
	HandledAt string `json:"handledAt,omitempty"`
}

// Appointment 交换预约，包含双方确认、撤回与取消闭环所需的全部状态。
type Appointment struct {
	ID            int            `json:"id"`
	Pair          string         `json:"pair"`
	Initiator     string         `json:"initiator"`
	Participant   string         `json:"participant"`
	Time          string         `json:"time"`
	Place         string         `json:"place"`
	State         string         `json:"state"`
	Status        string         `json:"status"` // 展示标签，由 State 派生
	Agenda        string         `json:"agenda"`
	Confirmations []string       `json:"confirmations"`
	CancelRequest *CancelRequest `json:"cancelRequest,omitempty"`
	SlotReleased  bool           `json:"slotReleased"`
	ReleasedAt    string         `json:"releasedAt,omitempty"`
	Version       int            `json:"version"`
	UpdatedAt     string         `json:"updatedAt"`
}

// IsParty 判断给定用户是否为预约双方之一。
func (a *Appointment) IsParty(user string) bool {
	return user == a.Initiator || user == a.Participant
}

// HasConfirmed 判断给定用户是否已确认。
func (a *Appointment) HasConfirmed(user string) bool {
	for _, name := range a.Confirmations {
		if name == user {
			return true
		}
	}
	return false
}

// BothConfirmed 判断双方是否均已确认。
func (a *Appointment) BothConfirmed() bool {
	return a.HasConfirmed(a.Initiator) && a.HasConfirmed(a.Participant)
}

// ActorInput 需要操作人身份的请求体。
type ActorInput struct {
	Actor string `json:"actor"`
}

// CancelAppointmentInput 提交取消原因的请求体。
type CancelAppointmentInput struct {
	Actor  string `json:"actor"`
	Reason string `json:"reason"`
}

// RespondCancelInput 处理取消申请的请求体。
type RespondCancelInput struct {
	Actor  string `json:"actor"`
	Action string `json:"action"` // accept / reject
}

// SlotLock 表示某用户被预约占用的时间档。
type SlotLock struct {
	User          string `json:"user"`
	Time          string `json:"time"`
	AppointmentID int    `json:"appointmentId"`
	Pair          string `json:"pair"`
	State         string `json:"state"`
}
