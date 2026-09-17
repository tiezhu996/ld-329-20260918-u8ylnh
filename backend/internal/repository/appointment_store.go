package repository

import (
	"sort"
	"strings"
	"sync"

	"cyskillswap/internal/errors"
	"cyskillswap/internal/model"
)

// AppointmentStore 交换预约的内存持久化实现。
// 所有状态流转都在 Mutate 内的临界区完成，保证"确认 / 撤回同时到达只有一个结果"。
// 进程内持久化：刷新后仍可回读终态，重复请求不会改变终态。
type AppointmentStore struct {
	mu      sync.Mutex
	nextID  int
	records map[int]model.Appointment
	// 时间档账本：user|slot -> 持有该档的预约 ID。
	slots map[string]int
}

func NewAppointmentStore() *AppointmentStore {
	store := &AppointmentStore{
		nextID:  1,
		records: make(map[int]model.Appointment),
		slots:   make(map[string]int),
	}
	store.bootstrap(seedAppointments())
	return store
}

func slotKey(user, slot string) string { return user + "|" + slot }

// bootstrap 装载预置数据并登记对应时间档占用。
func (s *AppointmentStore) bootstrap(seed []model.Appointment) {
	for _, appt := range seed {
		if appt.ID >= s.nextID {
			s.nextID = appt.ID + 1
		}
		if appt.SlotsHeld {
			s.holdSlotsLocked(appt.ID, appt.Initiator, appt.Responder, appt.Slots)
		}
		s.records[appt.ID] = appt
	}
}

// List 按 ID 升序返回全部预约。
func (s *AppointmentStore) List() []model.Appointment {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := make([]int, 0, len(s.records))
	for id := range s.records {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	result := make([]model.Appointment, 0, len(ids))
	for _, id := range ids {
		result = append(result, s.records[id])
	}
	return result
}

// Get 返回单个预约，found=false 时调用方按 404 处理。
func (s *AppointmentStore) Get(id int) (model.Appointment, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	appt, ok := s.records[id]
	return appt, ok
}

// Create 写入新预约；临界区内完成冲突复检与时间档占用，避免 TOCTOU。
func (s *AppointmentStore) Create(appt model.Appointment) (model.Appointment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if conflict := s.slotConflictLocked(0, appt.Initiator, appt.Responder, appt.Slots); conflict != "" {
		return model.Appointment{}, errors.SlotUnavailableError(conflict)
	}
	appt.ID = s.nextID
	s.nextID++
	s.holdSlotsLocked(appt.ID, appt.Initiator, appt.Responder, appt.Slots)
	appt.SlotsHeld = true
	s.records[appt.ID] = appt
	return appt, nil
}

// TxFunc 在临界区内读取当前快照并产出下一状态。
// holdDelta：+1 占用 / -1 释放 / 0 不变；changed=false 表示幂等重复请求。
type TxFunc func(current model.Appointment) (next model.Appointment, holdDelta int, changed bool, err error)

// Mutate 串行化单个预约的状态流转，状态判定与时间档释放/占用在同一临界区提交。
func (s *AppointmentStore) Mutate(id int, tx TxFunc) (model.Appointment, []string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	current, ok := s.records[id]
	if !ok {
		return model.Appointment{}, nil, false, errors.NotFoundError()
	}
	next, holdDelta, changed, err := tx(current)
	if err != nil || !changed {
		return current, nil, false, err
	}

	switch holdDelta {
	case -1:
		freed := s.releaseSlotsLocked(id, current.Initiator, current.Responder, current.Slots)
		next.SlotsHeld = false
		s.records[id] = next
		return next, freed, true, nil
	case 1:
		if conflict := s.slotConflictLocked(id, next.Initiator, next.Responder, next.Slots); conflict != "" {
			return current, nil, false, errors.SlotUnavailableError(conflict)
		}
		s.holdSlotsLocked(id, next.Initiator, next.Responder, next.Slots)
		next.SlotsHeld = true
		s.records[id] = next
		return next, nil, true, nil
	default:
		next.SlotsHeld = current.SlotsHeld
		s.records[id] = next
		return next, nil, true, nil
	}
}

// EnsureSlotsFree 在创建前检查双方时间档是否空闲。
func (s *AppointmentStore) EnsureSlotsFree(initiator, responder string, slots []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if msg := s.slotConflictLocked(0, initiator, responder, slots); msg != "" {
		return errors.SlotUnavailableError(msg)
	}
	return nil
}

// HeldSlots 返回指定用户当前被占用的时间档（撤回/取消生效后立即消失）。
func (s *AppointmentStore) HeldSlots(user string) []model.SlotStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]model.SlotStatus, 0)
	for key, apptID := range s.slots {
		parts := strings.SplitN(key, "|", 2)
		if len(parts) != 2 || parts[0] != user || apptID == 0 {
			continue
		}
		result = append(result, model.SlotStatus{User: user, Slot: parts[1], Held: true})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Slot < result[j].Slot })
	return result
}
