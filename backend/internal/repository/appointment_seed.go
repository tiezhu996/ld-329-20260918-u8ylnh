package repository

import (
	"time"

	"cyskillswap/internal/constants"
	"cyskillswap/internal/model"
)

// seed 写入初始预约数据。原有预约（ID 1、2）保持可用：
// 原“双方已确认”映射为 confirmed，“等待对方确认”映射为 pending。
// 调用方必须已持有 s.mu 或在单线程上下文中调用。
func (s *AppointmentStore) seed() {
	now := time.Now().Format(time.RFC3339)
	items := []model.Appointment{
		{
			ID: 1, Pair: "林澈 ↔ 孟野", Initiator: "林澈", Participant: "孟野",
			Time: "周六 10:00", Place: "东校区湖边",
			State: constants.AppointmentStateConfirmed, Status: constants.AppointmentLabelConfirmed,
			Agenda:        "先拍宣传照，再约 2 次吉他课",
			Confirmations: []string{"林澈", "孟野"},
			Version:       1, UpdatedAt: now,
		},
		{
			ID: 2, Pair: "周芮 ↔ 许安", Initiator: "周芮", Participant: "许安",
			Time: "周二 19:30", Place: "线上会议室",
			State: constants.AppointmentStatePending, Status: constants.AppointmentLabelPending,
			Agenda:        "导入问卷 CSV 并完成基础可视化",
			Confirmations: []string{"周芮"},
			Version:       1, UpdatedAt: now,
		},
		{
			ID: 3, Pair: "林澈 ↔ 周芮", Initiator: "林澈", Participant: "周芮",
			Time: "周三 20:00", Place: "线上会议室",
			State: constants.AppointmentStatePending, Status: constants.AppointmentLabelPending,
			Agenda:        "Python 数据分析入门答疑，待周芮确认",
			Confirmations: []string{"林澈"},
			Version:       1, UpdatedAt: now,
		},
		{
			ID: 4, Pair: "孟野 ↔ 林澈", Initiator: "孟野", Participant: "林澈",
			Time: "周日 15:00", Place: "西校区琴房",
			State: constants.AppointmentStatePending, Status: constants.AppointmentLabelPending,
			Agenda:        "吉他扫弦入门第一课，待林澈确认",
			Confirmations: []string{"孟野"},
			Version:       1, UpdatedAt: now,
		},
	}
	s.items = map[int]*model.Appointment{}
	for i := range items {
		s.items[items[i].ID] = &items[i]
	}
}
