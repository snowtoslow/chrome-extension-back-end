package http

import (
	"chrome-extension-back-end/user"
	"github.com/gin-gonic/gin"
)

func RegisterHTTPEndpoints(router *gin.RouterGroup, uc user.UseCase) {
	h := NewHandler(uc)

	bookmarks := router.Group("/user")
	{
		bookmarks.POST("", h.CreateUser)
		bookmarks.GET("/:id", h.GetUserById)
	}
}
