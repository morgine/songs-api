package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/morgine/songs/src/message"
	"net/http"
)


type Message struct {
	Status  message.Status
	Message string
	Data    interface{}
}

func SendMessage(ctx *gin.Context, status message.Status, message string) {
	ctx.AbortWithStatusJSON(http.StatusOK, Message{
		Status:  status,
		Message: message,
	})
}

func SendJSON(ctx *gin.Context, data interface{}) {
	ctx.AbortWithStatusJSON(http.StatusOK, Message{
		Data:   data,
	})
}

func SendError(ctx *gin.Context, err error) {
	if status, ok := err.(message.Status); ok {
		ctx.AbortWithStatusJSON(http.StatusOK, Message{
			Status:  status,
			Message: message.StatusText[status],
		})
	} else {
		ctx.AbortWithStatusJSON(http.StatusOK, Message{
			Status:  message.ErrUnknown,
			Message: err.Error(),
		})
	}
}
