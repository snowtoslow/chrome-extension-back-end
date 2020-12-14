package http

import (
	"chrome-extension-back-end/user"
	"github.com/gin-gonic/gin"
)

func RegisterHTTPEndpoints(router *gin.RouterGroup, uc user.UseCase) {
	h := NewHandler(uc)

	users := router.Group("/user")
	{
		users.POST("", h.CreateUser)
		users.GET("/:id", h.GetUserById)
		users.PATCH("", h.UpdateData)
	}
}
