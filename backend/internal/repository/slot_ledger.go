package repository

import "fmt"

// 时间档账本的纯账本操作，调用方必须已持有 store.mu。

// holdSlotsLocked 登记双方在这些时间档上的占用。
func (s *AppointmentStore) holdSlotsLocked(apptID int, initiator, responder string, slots []string) {
	for _, slot := range slots {
		s.slots[slotKey(initiator, slot)] = apptID
		s.slots[slotKey(responder, slot)] = apptID
	}
}

// releaseSlotsLocked 仅释放仍由本预约持有的时间档，
// 返回被释放的档位数组（撤回 / 取消生效后双方时间档立即释放）。
func (s *AppointmentStore) releaseSlotsLocked(apptID int, initiator, responder string, slots []string) []string {
	freed := make([]string, 0, len(slots))
	for _, slot := range slots {
		for _, user := range []string{initiator, responder} {
			key := slotKey(user, slot)
			if holder, ok := s.slots[key]; ok && holder == apptID {
				delete(s.slots, key)
			}
		}
		freed = append(freed, slot)
	}
	return freed
}

// slotConflictLocked 检查双方时间档是否被其它未结束预约占用。
// 空串表示无冲突；非空串为可直接展示给用户的冲突说明。
func (s *AppointmentStore) slotConflictLocked(selfID int, initiator, responder string, slots []string) string {
	for _, user := range []string{initiator, responder} {
		for _, slot := range slots {
			holder, ok := s.slots[slotKey(user, slot)]
			if ok && holder != 0 && holder != selfID {
				return fmt.Sprintf("%s 的时间档「%s」已被%s占用", user, slot, s.describeHolder(holder))
			}
		}
	}
	return ""
}

func (s *AppointmentStore) describeHolder(apptID int) string {
	if appt, ok := s.records[apptID]; ok {
		return fmt.Sprintf("预约 #%d（%s）", apptID, appt.Pair)
	}
	return fmt.Sprintf("预约 #%d", apptID)
}
