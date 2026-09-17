package controller

import (
	"net/http"

	"cyskillswap/internal/errors"
	"github.com/gin-gonic/gin"
)

// respondError 统一把业务错误翻译成 HTTP 响应，错误码与文案集中在 errors 包维护。
func respondError(c *gin.Context, err error) {
	if biz, ok := err.(errors.BusinessError); ok {
		c.JSON(biz.Status, gin.H{"error": biz})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": errors.BusinessError{
			Code:    "INTERNAL_ERROR",
			Message: "服务内部错误",
			Status:  http.StatusInternalServerError,
		},
	})
}
