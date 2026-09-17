package service

import (
	"strings"

	"cyskillswap/internal/constants"
	"cyskillswap/internal/errors"
)

// CreateAppointmentInput 发起预约的入参。
type CreateAppointmentInput struct {
	Initiator string   `json:"initiator"`
	Responder string   `json:"responder"`
	Time      string   `json:"time"`
	Place     string   `json:"place"`
	Agenda    string   `json:"agenda"`
	Slots     []string `json:"slots"`
}

// Validate 统一做入参校验，错误走 BusinessError，不抛裸字符串。
func (in *CreateAppointmentInput) Validate() error {
	in.Initiator = strings.TrimSpace(in.Initiator)
	in.Responder = strings.TrimSpace(in.Responder)
	in.Time = strings.TrimSpace(in.Time)
	in.Place = strings.TrimSpace(in.Place)
	in.Agenda = strings.TrimSpace(in.Agenda)

	if in.Initiator == "" || len([]rune(in.Initiator)) > constants.MaxUserLength {
		return errors.ValidationError("发起人姓名必填且不能超长")
	}
	if in.Responder == "" || len([]rune(in.Responder)) > constants.MaxUserLength {
		return errors.ValidationError("响应人姓名必填且不能超长")
	}
	if in.Initiator == in.Responder {
		return errors.ValidationError("发起人不能与响应人相同")
	}
	if in.Time == "" || len([]rune(in.Time)) > constants.MaxTimeLength {
		return errors.ValidationError("交换时间必填且不能超长")
	}
	if in.Place == "" || len([]rune(in.Place)) > constants.MaxPlaceLength {
		return errors.ValidationError("交换地点必填且不能超长")
	}
	if len([]rune(in.Agenda)) > constants.MaxAgendaLength {
		return errors.ValidationError("协商议程过长")
	}
	if len(in.Slots) == 0 {
		return errors.ValidationError("至少选择一个时间档")
	}
	seen := make(map[string]struct{}, len(in.Slots))
	for _, slot := range in.Slots {
		slot = strings.TrimSpace(slot)
		if slot == "" {
			return errors.ValidationError("时间档不能为空")
		}
		if _, dup := seen[slot]; dup {
			return errors.ValidationError("时间档不能重复：" + slot)
		}
		seen[slot] = struct{}{}
	}
	return nil
}
