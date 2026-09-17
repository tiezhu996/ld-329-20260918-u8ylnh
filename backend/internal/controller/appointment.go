package controller

import (
	"net/http"
	"strconv"

	"cyskillswap/internal/errors"
	"cyskillswap/internal/model"
	"cyskillswap/internal/service"
	"github.com/gin-gonic/gin"
)

// AppointmentDetail 按 ID 回读单个预约。
func AppointmentDetail(c *gin.Context) {
	id, ok := parseAppointmentID(c)
	if !ok {
		return
	}
	appt, err := service.GetAppointment(id)
	respondAppointment(c, appt, err)
}

// ConfirmAppointment 任一方确认预约。
func ConfirmAppointment(c *gin.Context) {
	id, ok := parseAppointmentID(c)
	if !ok {
		return
	}
	var input model.ActorInput
	if !bindJSON(c, &input) {
		return
	}
	appt, err := service.ConfirmAppointment(id, input.Actor)
	respondAppointment(c, appt, err)
}

// WithdrawAppointment 发起人撤回预约。
func WithdrawAppointment(c *gin.Context) {
	id, ok := parseAppointmentID(c)
	if !ok {
		return
	}
	var input model.ActorInput
	if !bindJSON(c, &input) {
		return
	}
	appt, err := service.WithdrawAppointment(id, input.Actor)
	respondAppointment(c, appt, err)
}

// RequestCancel 确认锁定后提交取消原因。
func RequestCancel(c *gin.Context) {
	id, ok := parseAppointmentID(c)
	if !ok {
		return
	}
	var input model.CancelAppointmentInput
	if !bindJSON(c, &input) {
		return
	}
	appt, err := service.RequestCancel(id, input.Actor, input.Reason)
	respondAppointment(c, appt, err)
}

// RespondCancel 对方处理取消申请。
func RespondCancel(c *gin.Context) {
	id, ok := parseAppointmentID(c)
	if !ok {
		return
	}
	var input model.RespondCancelInput
	if !bindJSON(c, &input) {
		return
	}
	appt, err := service.RespondCancel(id, input.Actor, input.Action)
	respondAppointment(c, appt, err)
}

// SlotLocks 查询某用户被进行中预约占用的时间档。
func SlotLocks(c *gin.Context) {
	user := c.Query("user")
	c.JSON(http.StatusOK, gin.H{"user": user, "locks": service.SlotLocks(user)})
}

func parseAppointmentID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		respondError(c, errors.BadRequest("APPOINTMENT_ID_INVALID", "预约 ID 不合法"))
		return 0, false
	}
	return id, true
}

func bindJSON(c *gin.Context, input any) bool {
	if err := c.ShouldBindJSON(input); err != nil {
		respondError(c, errors.BadRequest("REQUEST_BODY_INVALID", "请求体格式不正确"))
		return false
	}
	return true
}

func respondAppointment(c *gin.Context, appt model.Appointment, err error) {
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, appt)
}

func respondError(c *gin.Context, err error) {
	if biz, ok := err.(errors.BusinessError); ok {
		c.JSON(biz.Status, biz)
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "服务内部错误"})
}
