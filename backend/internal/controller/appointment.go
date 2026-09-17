package controller

import (
	"net/http"
	"strconv"
	"strings"

	"cyskillswap/internal/errors"
	"cyskillswap/internal/logger"
	"cyskillswap/internal/service"
	"github.com/gin-gonic/gin"
)

// ActorHeader 演示用：JWT 认证预留，当前以请求头 X-User-Name 标识操作人。
const ActorHeader = "X-User-Name"

// ActorQuery 便于浏览器直接刷新 / 联调，也支持 ?actor= 查询参数。
const ActorQuery = "actor"

type AppointmentController struct {
	service *service.AppointmentService
}

func NewAppointmentController(svc *service.AppointmentService) *AppointmentController {
	return &AppointmentController{service: svc}
}

func currentActor(c *gin.Context) string {
	actor := strings.TrimSpace(c.GetHeader(ActorHeader))
	if actor == "" {
		actor = strings.TrimSpace(c.Query(ActorQuery))
	}
	return actor
}

func parseAppointmentID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		respondError(c, errors.ValidationError("预约 ID 非法"))
		return 0, false
	}
	return id, true
}

func (ctl *AppointmentController) List(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"appointments": ctl.service.List()})
}

func (ctl *AppointmentController) Get(c *gin.Context) {
	id, ok := parseAppointmentID(c)
	if !ok {
		return
	}
	appt, err := ctl.service.Get(id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"appointment": appt})
}

func (ctl *AppointmentController) Create(c *gin.Context) {
	var input service.CreateAppointmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, errors.ValidationError("请求体格式不正确"))
		return
	}
	if input.Initiator == "" {
		input.Initiator = currentActor(c)
	}
	appt, err := ctl.service.Create(input)
	if err != nil {
		respondError(c, err)
		return
	}
	logger.Info("appointment created", appt.ID, appt.Pair)
	c.JSON(http.StatusCreated, gin.H{"appointment": appt})
}

func (ctl *AppointmentController) Confirm(c *gin.Context) {
	id, ok := parseAppointmentID(c)
	if !ok {
		return
	}
	result, err := ctl.service.Confirm(id, currentActor(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctl *AppointmentController) Withdraw(c *gin.Context) {
	id, ok := parseAppointmentID(c)
	if !ok {
		return
	}
	result, err := ctl.service.Withdraw(id, currentActor(c))
	if err != nil {
		respondError(c, err)
		return
	}
	logger.Info("appointment withdrawn", id, "freed", result.SlotsFreed)
	c.JSON(http.StatusOK, result)
}

type cancelRequest struct {
	Reason string `json:"reason"`
}

func (ctl *AppointmentController) RequestCancel(c *gin.Context) {
	id, ok := parseAppointmentID(c)
	if !ok {
		return
	}
	var body cancelRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, errors.ValidationError("请求体格式不正确"))
		return
	}
	result, err := ctl.service.RequestCancel(id, currentActor(c), body.Reason)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

type cancelDecision struct {
	Decision string `json:"decision"`
}

func (ctl *AppointmentController) DecideCancel(c *gin.Context) {
	id, ok := parseAppointmentID(c)
	if !ok {
		return
	}
	var body cancelDecision
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, errors.ValidationError("请求体格式不正确"))
		return
	}
	result, err := ctl.service.DecideCancel(id, currentActor(c), body.Decision)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctl *AppointmentController) Slots(c *gin.Context) {
	user := strings.TrimSpace(c.Param("user"))
	if user == "" {
		respondError(c, errors.ValidationError("缺少用户名"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user, "heldSlots": ctl.service.HeldSlots(user)})
}
