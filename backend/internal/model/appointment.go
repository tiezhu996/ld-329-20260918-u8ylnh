package model

// CancelRequest 记录确认锁定后任一方发起的取消申请。
// 已确认的预约不能直接撤回，只能提交取消原因等待对方处理。
type CancelRequest struct {
	By        string `json:"by"`
	Reason    string `json:"reason"`
	At        int64  `json:"at"`
	DecidedAt int64  `json:"decidedAt,omitempty"`
	Decision  string `json:"decision,omitempty"`
}

// Appointment 交换预约聚合根。
// Pair / Time / Place / Status / Agenda 与旧版只读接口保持兼容；
// 新增字段承载确认、撤回、取消闭环。
type Appointment struct {
	ID                 int            `json:"id"`
	Initiator          string         `json:"initiator"`
	Responder          string         `json:"responder"`
	Pair               string         `json:"pair"`
	Time               string         `json:"time"`
	Place              string         `json:"place"`
	Agenda             string         `json:"agenda"`
	Status             string         `json:"status"`
	StatusText         string         `json:"statusText"`
	InitiatorConfirmed bool           `json:"initiatorConfirmed"`
	ResponderConfirmed bool           `json:"responderConfirmed"`
	Slots              []string       `json:"slots"`
	SlotsHeld          bool           `json:"slotsHeld"`
	Cancel             *CancelRequest `json:"cancel,omitempty"`
	Version            int            `json:"version"`
	CreatedAt          int64          `json:"createdAt"`
	UpdatedAt          int64          `json:"updatedAt"`
}

// SlotStatus 描述某个用户在某个时间档上的占用情况。
type SlotStatus struct {
	User string `json:"user"`
	Slot string `json:"slot"`
	Held bool   `json:"held"`
}

// ActionResult 状态流转操作结果（幂等重复请求时 changed=false）。
type ActionResult struct {
	Appointment Appointment `json:"appointment"`
	Changed     bool        `json:"changed"`
	SlotsFreed  []string    `json:"slotsFreed,omitempty"`
	Note        string      `json:"note,omitempty"`
}
