package routes

import (
	"cyskillswap/internal/constants"
	"cyskillswap/internal/controller"
	"cyskillswap/internal/repository"
	"cyskillswap/internal/service"
	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	api := r.Group(constants.APIPrefix)
	api.GET("/health", controller.Health)
	api.GET("/dashboard/overview", controller.Overview)
	api.GET("/skills", controller.Skills)
	api.GET("/needs", controller.Needs)
	api.GET("/matches", controller.Matches)
	api.GET("/reviews", controller.Reviews)
	api.GET("/messages", controller.Messages)
	api.GET("/profile", controller.Profile)

	// 交换预约：确认 / 撤回 / 取消闭环。
	apptStore := repository.DefaultAppointmentStore()
	apptService := service.NewAppointmentService(apptStore)
	apptController := controller.NewAppointmentController(apptService)

	api.GET("/appointments", apptController.List)
	api.POST("/appointments", apptController.Create)
	api.GET("/appointments/:id", apptController.Get)
	api.POST("/appointments/:id/confirm", apptController.Confirm)
	api.POST("/appointments/:id/withdraw", apptController.Withdraw)
	api.POST("/appointments/:id/cancel-requests", apptController.RequestCancel)
	api.POST("/appointments/:id/cancel-decisions", apptController.DecideCancel)
	api.GET("/users/:user/slots", apptController.Slots)
}
