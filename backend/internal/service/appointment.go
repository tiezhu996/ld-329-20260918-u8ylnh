package service

import (
	"time"

	"cyskillswap/internal/constants"
	"cyskillswap/internal/errors"
	"cyskillswap/internal/model"
	"cyskillswap/internal/repository"
)

// AppointmentService 承载交换预约的确认 / 撤回 / 取消闭环。
type AppointmentService struct {
	store *repository.AppointmentStore
}

func NewAppointmentService(store *repository.AppointmentStore) *AppointmentService {
	return &AppointmentService{store: store}
}

func (s *AppointmentService) List() []model.Appointment { return s.store.List() }

func (s *AppointmentService) Get(id int) (model.Appointment, error) {
	appt, ok := s.store.Get(id)
	if !ok {
		return model.Appointment{}, errors.NotFoundError()
	}
	return appt, nil
}

func (s *AppointmentService) HeldSlots(user string) []model.SlotStatus {
	return s.store.HeldSlots(user)
}

// Create 发起一笔交换预约，创建即占用双方时间档，初始为待确认。
func (s *AppointmentService) Create(req CreateAppointmentInput) (model.Appointment, error) {
	if err := req.Validate(); err != nil {
		return model.Appointment{}, err
	}
	if err := s.store.EnsureSlotsFree(req.Initiator, req.Responder, req.Slots); err != nil {
		return model.Appointment{}, err
	}
	now := time.Now().Unix()
	appt := model.Appointment{
		Initiator:  req.Initiator,
		Responder:  req.Responder,
		Pair:       req.Initiator + " ↔ " + req.Responder,
		Time:       req.Time,
		Place:      req.Place,
		Agenda:     req.Agenda,
		Slots:      req.Slots,
		Status:     constants.ApptStatusPending,
		StatusText: constants.AppointmentStatusText[constants.ApptStatusPending],
		Version:    1,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	return s.store.Create(appt)
}

// Confirm 任一方确认。确认与发起人撤回并发时，仓库串行化保证只有一个结果生效。
func (s *AppointmentService) Confirm(id int, actor string) (model.ActionResult, error) {
	if actor == "" {
		return model.ActionResult{}, errors.ValidationError("缺少操作人")
	}
	appt, freed, changed, err := s.store.Mutate(id, func(current model.Appointment) (model.Appointment, int, bool, error) {
		if err := ensureParticipant(current, actor); err != nil {
			return current, 0, false, err
		}
		alreadyConfirmed := current.InitiatorConfirmed && current.ResponderConfirmed
		roleConfirmed := actorConfirmed(current, actor)

		switch {
		case current.Status == constants.ApptStatusWithdrawn:
			// 撤回已先生效：确认不得"复活"预约。
			return current, 0, false, errors.TerminalStateError("预约已被发起人撤回，确认无效")
		case current.Status == constants.ApptStatusCancelled:
			return current, 0, false, errors.TerminalStateError("预约已取消，确认无效")
		case current.Status == constants.ApptStatusCancelRequested:
			return current, 0, false, errors.StatusConflictError("对方已提交取消申请，请先处理取消请求")
		case alreadyConfirmed:
			// 双方已确认后的重复确认：终态语义，幂等返回，不改变任何状态。
			return current, 0, false, nil
		case roleConfirmed:
			// 同一方重复点击确认：幂等无副作用。
			return current, 0, false, nil
		case current.Status != constants.ApptStatusPending &&
			current.Status != constants.ApptStatusConfirmedPartial:
			return current, 0, false, errors.StatusConflictError("当前状态不允许确认")
		}

		next := current
		next.Version++
		next.UpdatedAt = time.Now().Unix()
		if actor == current.Initiator {
			next.InitiatorConfirmed = true
		} else {
			next.ResponderConfirmed = true
		}
		if next.InitiatorConfirmed && next.ResponderConfirmed {
			next.Status = constants.ApptStatusConfirmed
			// 锁定后不能再绕过取消流程，状态机在 Withdraw/RequestCancel 中拦截。
		} else {
			next.Status = constants.ApptStatusConfirmedPartial
		}
		next.StatusText = constants.AppointmentStatusText[next.Status]
		return next, 0, true, nil
	})
	if err != nil {
		return model.ActionResult{}, err
	}
	return model.ActionResult{Appointment: appt, Changed: changed, SlotsFreed: freed}, nil
}
