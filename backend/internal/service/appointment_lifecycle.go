package service

import (
	"time"

	"cyskillswap/internal/constants"
	"cyskillswap/internal/errors"
	"cyskillswap/internal/model"
)

// Withdraw 仅发起人可撤回；待确认（含单方已确认）阶段可撤回。
// 撤回生效后双方时间档在同一临界区内立即释放。
func (s *AppointmentService) Withdraw(id int, actor string) (model.ActionResult, error) {
	if actor == "" {
		return model.ActionResult{}, errors.ValidationError("缺少操作人")
	}
	appt, freed, changed, err := s.store.Mutate(id, func(current model.Appointment) (model.Appointment, int, bool, error) {
		if actor != current.Initiator {
			if actor == current.Responder {
				return current, 0, false, errors.ForbiddenError("待确认时仅发起人可撤回")
			}
			return current, 0, false, errors.NotParticipantError()
		}

		switch {
		case current.Status == constants.ApptStatusWithdrawn:
			// 重复撤回：终态幂等，时间档不重复释放。
			return current, 0, false, nil
		case current.Status == constants.ApptStatusCancelled:
			return current, 0, false, errors.TerminalStateError("预约已取消，不能撤回")
		case current.Status == constants.ApptStatusConfirmed:
			return current, 0, false, errors.StatusConflictError("预约已双方确认锁定，不能直接撤回，请提交取消原因")
		case current.Status == constants.ApptStatusCancelRequested:
			return current, 0, false, errors.StatusConflictError("已有取消申请待处理，不能撤回")
		case current.Status != constants.ApptStatusPending &&
			current.Status != constants.ApptStatusConfirmedPartial:
			return current, 0, false, errors.StatusConflictError("当前状态不允许撤回")
		}

		next := current
		next.Status = constants.ApptStatusWithdrawn
		next.StatusText = constants.AppointmentStatusText[next.Status]
		next.InitiatorConfirmed = false
		next.ResponderConfirmed = false
		next.Version++
		next.UpdatedAt = time.Now().Unix()
		return next, -1, true, nil
	})
	if err != nil {
		return model.ActionResult{}, err
	}
	note := ""
	if changed {
		note = "撤回已生效，双方时间档已立即释放"
	}
	return model.ActionResult{Appointment: appt, Changed: changed, SlotsFreed: freed, Note: note}, nil
}

// RequestCancel 确认锁定后只能提交取消原因，等待对方处理，不直接释放时间档。
func (s *AppointmentService) RequestCancel(id int, actor, reason string) (model.ActionResult, error) {
	if actor == "" {
		return model.ActionResult{}, errors.ValidationError("缺少操作人")
	}
	if len([]rune(reason)) < constants.MinCancelReasonLen {
		return model.ActionResult{}, errors.ValidationError("请填写取消原因（至少 4 个字）")
	}
	if len([]rune(reason)) > constants.MaxReasonLength {
		return model.ActionResult{}, errors.ValidationError("取消原因过长")
	}
	appt, _, changed, err := s.store.Mutate(id, func(current model.Appointment) (model.Appointment, int, bool, error) {
		if err := ensureParticipant(current, actor); err != nil {
			return current, 0, false, err
		}
		switch {
		case current.Status == constants.ApptStatusCancelled:
			return current, 0, false, errors.TerminalStateError("预约已取消")
		case current.Status == constants.ApptStatusWithdrawn:
			return current, 0, false, errors.TerminalStateError("预约已撤回")
		case current.Status == constants.ApptStatusCancelRequested:
			if current.Cancel != nil && current.Cancel.By == actor {
				// 发起方重复提交：幂等返回已存在的申请，不改状态。
				return current, 0, false, nil
			}
			return current, 0, false, errors.StatusConflictError("已有取消申请等待你处理")
		case current.Status != constants.ApptStatusConfirmed:
			return current, 0, false, errors.StatusConflictError("仅双方已确认的预约需要走取消流程，待确认阶段可由发起人撤回")
		}
		next := current
		next.Status = constants.ApptStatusCancelRequested
		next.StatusText = constants.AppointmentStatusText[next.Status]
		next.Cancel = &model.CancelRequest{By: actor, Reason: reason, At: time.Now().Unix()}
		next.Version++
		next.UpdatedAt = time.Now().Unix()
		// 不释放时间档：需对方同意后才释放。
		return next, 0, true, nil
	})
	if err != nil {
		return model.ActionResult{}, err
	}
	return model.ActionResult{Appointment: appt, Changed: changed}, nil
}

// DecideCancel 对方同意 / 拒绝取消；同意则进入终态并立即释放时间档。
func (s *AppointmentService) DecideCancel(id int, actor, decision string) (model.ActionResult, error) {
	if actor == "" {
		return model.ActionResult{}, errors.ValidationError("缺少操作人")
	}
	if decision != constants.CancelDecisionApprove && decision != constants.CancelDecisionReject {
		return model.ActionResult{}, errors.ValidationError("decision 只能是 approve 或 reject")
	}
	appt, freed, changed, err := s.store.Mutate(id, func(current model.Appointment) (model.Appointment, int, bool, error) {
		if err := ensureParticipant(current, actor); err != nil {
			return current, 0, false, err
		}
		switch {
		case current.Status == constants.ApptStatusCancelled:
			return current, 0, false, nil
		case current.Status == constants.ApptStatusWithdrawn:
			return current, 0, false, errors.TerminalStateError("预约已撤回")
		case current.Status != constants.ApptStatusCancelRequested || current.Cancel == nil:
			return current, 0, false, errors.StatusConflictError("没有等待处理的取消申请")
		}
		if current.Cancel.By == actor {
			return current, 0, false, errors.ForbiddenError("取消申请需由对方处理，不能自己审批")
		}

		next := current
		next.Version++
		next.UpdatedAt = time.Now().Unix()
		next.Cancel.DecidedAt = time.Now().Unix()
		next.Cancel.Decision = decision
		if decision == constants.CancelDecisionApprove {
			next.Status = constants.ApptStatusCancelled
			next.StatusText = constants.AppointmentStatusText[next.Status]
			next.InitiatorConfirmed = false
			next.ResponderConfirmed = false
			return next, -1, true, nil
		}
		next.Status = constants.ApptStatusConfirmed
		next.StatusText = constants.AppointmentStatusText[next.Status]
		return next, 0, true, nil
	})
	if err != nil {
		return model.ActionResult{}, err
	}
	note := ""
	if changed && decision == constants.CancelDecisionApprove {
		note = "取消已生效，双方时间档已立即释放"
	}
	return model.ActionResult{Appointment: appt, Changed: changed, SlotsFreed: freed, Note: note}, nil
}

func ensureParticipant(appt model.Appointment, actor string) error {
	if actor != appt.Initiator && actor != appt.Responder {
		return errors.NotParticipantError()
	}
	return nil
}

func actorConfirmed(appt model.Appointment, actor string) bool {
	if actor == appt.Initiator {
		return appt.InitiatorConfirmed
	}
	return appt.ResponderConfirmed
}
