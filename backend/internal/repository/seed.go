package repository

import (
	"cyskillswap/internal/constants"
	"cyskillswap/internal/model"
)

// seedAppointments 预置两条预约，保持与旧看板数据一致，
// 但补齐发起人 / 响应人 / 时间档 / 确认位，供确认与撤回闭环演示。
func seedAppointments() []model.Appointment {
	return []model.Appointment{
		{
			ID: 1, Initiator: "林澈", Responder: "孟野",
			Pair: "林澈 ↔ 孟野", Time: "周六 10:00", Place: "东校区湖边",
			Agenda:             "先拍宣传照，再约 2 次吉他课",
			Status:             constants.ApptStatusConfirmed,
			StatusText:         constants.AppointmentStatusText[constants.ApptStatusConfirmed],
			InitiatorConfirmed: true, ResponderConfirmed: true,
			Slots: []string{"周六上午"}, SlotsHeld: true,
		},
		{
			ID: 2, Initiator: "周芮", Responder: "许安",
			Pair: "周芮 ↔ 许安", Time: "周二 19:30", Place: "线上会议室",
			Agenda:             "导入问卷 CSV 并完成基础可视化",
			Status:             constants.ApptStatusPending,
			StatusText:         constants.AppointmentStatusText[constants.ApptStatusPending],
			InitiatorConfirmed: false, ResponderConfirmed: false,
			Slots: []string{"周二晚"}, SlotsHeld: true,
		},
	}
}
