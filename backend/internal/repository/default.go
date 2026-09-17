package repository

import "sync"

var (
	defaultAppointmentStore     *AppointmentStore
	defaultAppointmentStoreOnce sync.Once
)

// DefaultAppointmentStore 返回进程内唯一的预约仓储，
// 让看板概览与预约操作接口共享同一份数据，刷新后仍可回读。
func DefaultAppointmentStore() *AppointmentStore {
	defaultAppointmentStoreOnce.Do(func() {
		defaultAppointmentStore = NewAppointmentStore()
	})
	return defaultAppointmentStore
}
