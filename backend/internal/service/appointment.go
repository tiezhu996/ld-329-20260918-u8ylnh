package service

import (
	"strings"
	"time"

	"cyskillswap/internal/constants"
	"cyskillswap/internal/errors"
	"cyskillswap/internal/model"
	"cyskillswap/internal/repository"
)

// 交换预约确认/撤回/取消闭环状态机：
//
//	pending        --confirm(双方均确认)--> confirmed
//	pending        --withdraw(仅发起人)--> withdrawn（终态，释放双方时间档）
//	confirmed      --提交取消原因---------> cancel_pending
//	cancel_pending --对方 accept---------> cancelled（终态，释放双方时间档）
//	cancel_pending --对方 reject---------> confirmed
//
// 所有转换在存储层互斥锁内完成：确认与撤回同时到达时只有一个先生效，
// 另一个按最新状态校验并失败；终态下的重复请求幂等返回，不改变终态。

// ListAppointments 返回全部预约。
func ListAppointments() []model.Appointment {
	return repository.Appointments().List()
}

// GetAppointment 按 ID 回读预约（刷新后状态仍可读取）。
func GetAppointment(id int) (model.Appointment, error) {
	return repository.Appointments().Get(id)
}

// ConfirmAppointment 任一方确认预约；双方均确认后锁定。
func ConfirmAppointment(id int, actor string) (model.Appointment, error) {
	if err := requireActor(actor); err != nil {
		return model.Appointment{}, err
	}
	return repository.Appointments().Mutate(id, func(appt *model.Appointment) (bool, error) {
		if err := requireParty(appt, actor); err != nil {
			return false, err
		}
		switch appt.State {
		case constants.AppointmentStatePending:
			if appt.HasConfirmed(actor) {
				return false, nil // 重复确认：幂等，不改变状态
			}
			appt.Confirmations = append(appt.Confirmations, actor)
			if appt.BothConfirmed() {
				appt.State = constants.AppointmentStateConfirmed
			}
			refreshLabel(appt)
			return true, nil
		case constants.AppointmentStateConfirmed:
			return false, nil // 已锁定：重复确认幂等返回
		case constants.AppointmentStateCancelPending:
			return false, errors.Conflict(constants.ErrCodeCancelInProgress, "取消申请待处理，暂不能确认")
		default:
			return false, terminalError(appt)
		}
	})
}

// WithdrawAppointment 撤回预约：仅待确认状态且仅发起人可撤回，生效后立即释放双方时间档。
func WithdrawAppointment(id int, actor string) (model.Appointment, error) {
	if err := requireActor(actor); err != nil {
		return model.Appointment{}, err
	}
	return repository.Appointments().Mutate(id, func(appt *model.Appointment) (bool, error) {
		if err := requireParty(appt, actor); err != nil {
			return false, err
		}
		if actor != appt.Initiator {
			return false, errors.Forbidden(constants.ErrCodeOnlyInitiatorWithdraw, "待确认阶段仅发起人可撤回预约")
		}
		switch appt.State {
		case constants.AppointmentStatePending:
			releaseSlot(appt, constants.AppointmentStateWithdrawn)
			return true, nil
		case constants.AppointmentStateWithdrawn:
			return false, nil // 重复撤回：幂等，终态不变
		case constants.AppointmentStateConfirmed:
			return false, errors.Conflict(constants.ErrCodeLocked, "预约已双方确认锁定，请提交取消原因等待对方处理")
		case constants.AppointmentStateCancelPending:
			return false, errors.Conflict(constants.ErrCodeCancelInProgress, "取消申请待处理，暂不能撤回")
		default:
			return false, terminalError(appt)
		}
	})
}

// RequestCancel 确认锁定后提交取消原因，等待对方处理。
func RequestCancel(id int, actor, reason string) (model.Appointment, error) {
	if err := requireActor(actor); err != nil {
		return model.Appointment{}, err
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return model.Appointment{}, errors.BadRequest(constants.ErrCodeCancelReasonRequired, "请填写取消原因")
	}
	return repository.Appointments().Mutate(id, func(appt *model.Appointment) (bool, error) {
		if err := requireParty(appt, actor); err != nil {
			return false, err
		}
		switch appt.State {
		case constants.AppointmentStateConfirmed:
			appt.CancelRequest = &model.CancelRequest{
				By: actor, Reason: reason,
				Status:    constants.CancelStatusPending,
				CreatedAt: time.Now().Format(time.RFC3339),
			}
			appt.State = constants.AppointmentStateCancelPending
			refreshLabel(appt)
			return true, nil
		case constants.AppointmentStateCancelPending:
			if appt.CancelRequest != nil && appt.CancelRequest.By == actor {
				return false, nil // 重复提交：幂等，保留原取消申请
			}
			return false, errors.Conflict(constants.ErrCodeCancelAlreadyPending, "对方已提交取消申请，请直接处理该申请")
		case constants.AppointmentStatePending:
			return false, errors.Conflict(constants.ErrCodeLocked, "预约尚未确认锁定，发起人可直接撤回")
		default:
			return false, terminalError(appt)
		}
	})
}

// RespondCancel 对方处理取消申请：同意则取消并释放时间档，拒绝则回到已确认。
func RespondCancel(id int, actor, action string) (model.Appointment, error) {
	if err := requireActor(actor); err != nil {
		return model.Appointment{}, err
	}
	if action != constants.CancelActionAccept && action != constants.CancelActionReject {
		return model.Appointment{}, errors.BadRequest(constants.ErrCodeInvalidCancelAction, "无效的处理动作，仅支持 accept / reject")
	}
	return repository.Appointments().Mutate(id, func(appt *model.Appointment) (bool, error) {
		if err := requireParty(appt, actor); err != nil {
			return false, err
		}
		if appt.State != constants.AppointmentStateCancelPending || appt.CancelRequest == nil ||
			appt.CancelRequest.Status != constants.CancelStatusPending {
			return false, errors.Conflict(constants.ErrCodeNoCancelRequest, "当前没有待处理的取消申请")
		}
		if actor == appt.CancelRequest.By {
			return false, errors.Forbidden(constants.ErrCodeNotCancelResponder, "取消申请需等待对方处理")
		}
		now := time.Now().Format(time.RFC3339)
		appt.CancelRequest.HandledBy = actor
		appt.CancelRequest.HandledAt = now
		if action == constants.CancelActionAccept {
			appt.CancelRequest.Status = constants.CancelStatusAccepted
			releaseSlot(appt, constants.AppointmentStateCancelled)
		} else {
			appt.CancelRequest.Status = constants.CancelStatusRejected
			appt.State = constants.AppointmentStateConfirmed
			refreshLabel(appt)
		}
		return true, nil
	})
}

// SlotLocks 返回指定用户被进行中的预约占用的时间档；
// 撤回或取消生效后锁定立即消失，即双方时间档立即释放。
func SlotLocks(user string) []model.SlotLock {
	locks := []model.SlotLock{}
	for _, appt := range repository.Appointments().List() {
		if appt.SlotReleased || !appt.IsParty(user) || isTerminal(appt.State) {
			continue
		}
		locks = append(locks, model.SlotLock{
			User: user, Time: appt.Time,
			AppointmentID: appt.ID, Pair: appt.Pair, State: appt.State,
		})
	}
	return locks
}

func requireActor(actor string) error {
	if strings.TrimSpace(actor) == "" {
		return errors.BadRequest(constants.ErrCodeActorRequired, "缺少操作人身份")
	}
	return nil
}

func requireParty(appt *model.Appointment, actor string) error {
	if !appt.IsParty(actor) {
		return errors.Forbidden(constants.ErrCodeNotParty, "仅预约双方可以操作该预约")
	}
	return nil
}

// releaseSlot 将预约置为终态并立即释放双方时间档。
func releaseSlot(appt *model.Appointment, terminalState string) {
	appt.State = terminalState
	appt.SlotReleased = true
	appt.ReleasedAt = time.Now().Format(time.RFC3339)
	refreshLabel(appt)
}

func refreshLabel(appt *model.Appointment) {
	if label, ok := constants.AppointmentStateLabels[appt.State]; ok {
		appt.Status = label
	}
}

func isTerminal(state string) bool {
	return state == constants.AppointmentStateWithdrawn || state == constants.AppointmentStateCancelled
}

func terminalError(appt *model.Appointment) error {
	return errors.Conflict(constants.ErrCodeTerminal, "预约"+appt.Status+"，终态不可变更")
}
