package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/morgine/songs/src/ad"
	"github.com/morgine/songs/src/message"
	"github.com/morgine/songs/src/platform"
)

type AdvertPlatform struct {
	pt   *platform.AdvertPlatform
	host string
}

func NewAdvertPlatform(pt *platform.AdvertPlatform, host string) *AdvertPlatform {
	return &AdvertPlatform{pt: pt, host: host}
}

// 目前授权信息保存在服务器端，该方法用于检测授权是否过期，授权过期应该提示用户重新授权
func (ap *AdvertPlatform) CheckAdvertAuthorize(ctx *gin.Context) {
	_, err := ap.pt.GetAccessToken()
	if err != nil {
		SendError(ctx, err)
	} else {
		SendJSON(ctx, "已授权")
	}
}

// 获得授权地址
func (ap *AdvertPlatform) GetAdvertAuthorizerUrl(redirectRoute string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		SendJSON(ctx, ap.pt.AuthUrl(resetHostUrl(ap.host, redirectRoute), ""))
	}
}

// 授权跳转地址，正常逻辑应当将 access token 返回给客户端，这里暂时保存在服务端
func (ap *AdvertPlatform) ListenAdvertAuthorizerCode(thisRoute string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		code, _ := ap.pt.AuthCode(ctx.Request)
		err := ap.pt.SaveAccessToken(code, resetHostUrl(ap.host, thisRoute))
		if err != nil {
			SendMessage(ctx, message.StatusOK, "授权失败："+err.Error())
		} else {
			SendMessage(ctx, message.StatusOK, "授权成功")
		}
	}
}

// 获得日报 level 可用字段
func (ap *AdvertPlatform) GetDailyReportsLevelFields(ctx *gin.Context) {
	type params struct {
		Level string
	}
	ps := &params{}
	err := ctx.Bind(ps)
	if err != nil {
		SendError(ctx, err)
	} else {
		fields := ap.pt.GetLevelFields(ps.Level)
		SendJSON(ctx, fields)
	}
}

// 获得日报
func (ap *AdvertPlatform) GetDailyReports(ctx *gin.Context) {
	query := &ad.GetDailyReportsOptions{}
	err := ctx.BindQuery(query)
	if err != nil {
		SendError(ctx, err)
	} else {
		reports, err := ap.pt.GetDailyReports(query)
		if err != nil {
			SendError(ctx, err)
		} else {
			SendJSON(ctx, reports)
		}
	}
}
