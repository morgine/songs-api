package src

import (
	"github.com/gin-gonic/gin"
)

func WithSkipHandler(skipFunc func(ctx *gin.Context) (skip bool), h gin.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !skipFunc(ctx) {
			h(ctx)
		}
	}
}
