package controller

import (
	"net/http"

	"cyskillswap/internal/constants"
	"cyskillswap/internal/service"
	"github.com/gin-gonic/gin"
)

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": constants.ServiceName})
}
func Overview(c *gin.Context)     { c.JSON(http.StatusOK, service.Overview()) }
func Skills(c *gin.Context)       { c.JSON(http.StatusOK, service.Skills()) }
func Needs(c *gin.Context)        { c.JSON(http.StatusOK, service.Needs()) }
func Matches(c *gin.Context)      { c.JSON(http.StatusOK, service.Matches()) }
func Appointments(c *gin.Context) { c.JSON(http.StatusOK, service.Appointments()) }
func Reviews(c *gin.Context)      { c.JSON(http.StatusOK, service.Reviews()) }
func Messages(c *gin.Context)     { c.JSON(http.StatusOK, service.Messages()) }
func Profile(c *gin.Context)      { c.JSON(http.StatusOK, service.Profile()) }
