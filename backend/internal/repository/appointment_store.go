package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"cyskillswap/internal/constants"
	"cyskillswap/internal/errors"
	"cyskillswap/internal/logger"
	"cyskillswap/internal/model"
)

// AppointmentStore 预约存储：互斥锁保证"读-改-写-持久化"整体串行，
// 确认与撤回同时到达时只会有一个先生效；每次变更原子落盘，刷新后可回读。
type AppointmentStore struct {
	mu    sync.Mutex
	path  string
	items map[int]*model.Appointment
}

var (
	defaultStore     *AppointmentStore
	defaultStoreLock sync.Mutex
)

// Appointments 返回默认预约存储（懒加载单例）。
func Appointments() *AppointmentStore {
	defaultStoreLock.Lock()
	defer defaultStoreLock.Unlock()
	if defaultStore == nil {
		defaultStore = NewAppointmentStore(storePathFromEnv())
	}
	return defaultStore
}

// SetAppointmentsForTest 替换默认存储，仅供测试使用。
func SetAppointmentsForTest(s *AppointmentStore) {
	defaultStoreLock.Lock()
	defer defaultStoreLock.Unlock()
	defaultStore = s
}

func storePathFromEnv() string {
	if path := os.Getenv("APPOINTMENT_STORE_PATH"); path != "" {
		return path
	}
	return constants.DefaultAppointmentStorePath
}

// NewAppointmentStore 创建存储：优先从文件回读，文件不存在时写入种子数据。
func NewAppointmentStore(path string) *AppointmentStore {
	s := &AppointmentStore{path: path, items: map[int]*model.Appointment{}}
	if err := s.load(); err != nil {
		logger.Warn("appointment store load failed, seeding defaults:", err)
		s.seed()
		if err := s.saveLocked(); err != nil {
			logger.Error("appointment store seed save failed:", err)
		}
	}
	return s
}

func (s *AppointmentStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	var items []model.Appointment
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}
	s.items = map[int]*model.Appointment{}
	for i := range items {
		item := items[i]
		s.items[item.ID] = &item
	}
	logger.Info("appointment store loaded", len(s.items), "appointments from", s.path)
	return nil
}

// saveLocked 原子落盘：先写临时文件再重命名，避免半截文件。
// 调用方必须已持有 s.mu。
func (s *AppointmentStore) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	items := s.listLocked()
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// listLocked 返回按 ID 排序的深拷贝列表。调用方必须已持有 s.mu。
func (s *AppointmentStore) listLocked() []model.Appointment {
	ids := make([]int, 0, len(s.items))
	for id := range s.items {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	items := make([]model.Appointment, 0, len(ids))
	for _, id := range ids {
		items = append(items, cloneAppointment(s.items[id]))
	}
	return items
}

func cloneAppointment(a *model.Appointment) model.Appointment {
	clone := *a
	clone.Confirmations = append([]string{}, a.Confirmations...)
	if a.CancelRequest != nil {
		req := *a.CancelRequest
		clone.CancelRequest = &req
	}
	return clone
}

// List 返回全部预约的深拷贝，按 ID 排序。
func (s *AppointmentStore) List() []model.Appointment {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.listLocked()
}

// Get 返回单个预约的深拷贝。
func (s *AppointmentStore) Get(id int) (model.Appointment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	appt, ok := s.items[id]
	if !ok {
		return model.Appointment{}, errors.NotFound(constants.ErrCodeAppointmentNotFound, "预约不存在")
	}
	return cloneAppointment(appt), nil
}

// Mutate 在互斥锁内执行状态转换：fn 返回 changed=true 时才推进版本并落盘。
// 并发的确认/撤回/取消请求在此串行化，同一时刻只有一个结果生效。
func (s *AppointmentStore) Mutate(id int, fn func(*model.Appointment) (bool, error)) (model.Appointment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	appt, ok := s.items[id]
	if !ok {
		return model.Appointment{}, errors.NotFound(constants.ErrCodeAppointmentNotFound, "预约不存在")
	}
	changed, err := fn(appt)
	if err != nil {
		return model.Appointment{}, err
	}
	if changed {
		appt.Version++
		appt.UpdatedAt = time.Now().Format(time.RFC3339)
		if err := s.saveLocked(); err != nil {
			logger.Error("appointment store save failed:", err)
			return model.Appointment{}, errors.New(500, constants.ErrCodeStoreSaveFailed, "预约状态保存失败，请重试")
		}
	}
	return cloneAppointment(appt), nil
}
