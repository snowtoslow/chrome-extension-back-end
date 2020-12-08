package http

import (
	"chrome-extension-back-end/models"
	"chrome-extension-back-end/user"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type Handler struct {
	useCase user.UseCase
}

func NewHandler(usecase user.UseCase) *Handler {
	return &Handler{
		useCase: usecase,
	}
}

func (h Handler) CreateUser(c *gin.Context) {
	var createdUser *models.User
	if err := c.ShouldBindJSON(&createdUser); err != nil {
		log.Print(err)
		c.JSON(http.StatusBadRequest, gin.H{"msg": err})
		return
	}
	err := h.useCase.CreateUser(c.Request.Context(), createdUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": createdUser.Id})
}

func (h Handler) GetUserById(c *gin.Context) {
	id := c.Params.ByName("id")
	foundUser, err := h.useCase.GetUserById(c.Request.Context(), id)
	if err != nil {
		log.Println("ERR!", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed", "message": user.InvalidIdError})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "user": &foundUser})
}
