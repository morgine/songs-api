package env

import (
	"github.com/gin-gonic/gin"
	"github.com/morgine/songs/src/handler"
	"github.com/morgine/songs/src/message"
)

type adminSender int

func (a adminSender) SendData(ctx *gin.Context, data interface{}) {
	handler.SendJSON(ctx, data)
}

func (a adminSender) SendMsgSuccess(ctx *gin.Context, msg string) {
	handler.SendMessage(ctx, message.StatusOK, msg)
}

func (a adminSender) SendMsgWarning(ctx *gin.Context, msg string) {
	handler.SendMessage(ctx, message.StatusOK, msg)
}

func (a adminSender) SendMsgError(ctx *gin.Context, msg string) {
	handler.SendMessage(ctx, message.StatusOK, msg)
}

func (a adminSender) SendError(ctx *gin.Context, err error) {
	handler.SendError(ctx, err)
}
