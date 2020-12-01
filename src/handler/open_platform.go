package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/morgine/log"
	"github.com/morgine/songs/src/message"
	"github.com/morgine/songs/src/model"
	"github.com/morgine/songs/src/platform"
	"github.com/morgine/songs/src/wpt"
	"gorm.io/gorm"
	"math/rand"
	"time"
)

type OpenPlatform struct {
	pt   *platform.OpenPlatform
	gorm *model.AppGorm
	host string
}

func NewOpenPlatform(pt *platform.OpenPlatform, db *gorm.DB, host string) (*OpenPlatform, error) {
	err := db.AutoMigrate(&model.App{})
	if err != nil {
		return nil, err
	}
	return &OpenPlatform{
		pt:   pt,
		gorm: model.NewGorm(db).App(),
		host: host,
	}, nil
}

func (op *OpenPlatform) ListenMessage(appidGetter func(ctx *gin.Context) string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取 APPID
		_ = appidGetter(ctx)
		_, echoStr, err := op.pt.ListenMessage(ctx.Request)
		if err != nil {
			log.Emergency.Println(err)
		} else {
			if echoStr != "" {
				ctx.Writer.WriteString(echoStr)
			} else {
				// 处理 message
			}
		}
	}
}

func (op *OpenPlatform) ListenVerifyTicket(ctx *gin.Context) {
	err := op.pt.ListenTicket(ctx.Request)
	if err != nil {
		// 不返回 success, 表示该请求暂未处理, 后续可继续处理
		log.Emergency.Println(err)
	} else {
		ctx.Writer.WriteString("success")
	}
}

func (op *OpenPlatform) GetAppAuthorizerUrl(redirectUrl string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		uri, err := op.pt.OpenPlatformAuthRedirectUrl(resetHostUrl(op.host, redirectUrl))
		if err != nil {
			SendError(ctx, err)
		} else {
			SendJSON(ctx, uri)
		}
	}
}

func resetHostUrl(host, route string) string {
	return host + route
}

func (op *OpenPlatform) ListenAppAuthorizerCode(ctx *gin.Context) {
	err := op.pt.ListenAuthorized(ctx.Request)
	if err != nil {
		SendMessage(ctx, message.StatusOK, "授权失败："+err.Error())
	} else {
		SendMessage(ctx, message.StatusOK, "授权成功")
	}
}

func (op *OpenPlatform) ResetAppAuthorizers(ctx *gin.Context) {
	err := op.pt.ResetAuthorizers()
	if err != nil {
		SendMessage(ctx, message.StatusOK, "操作失败："+err.Error())
	} else {
		SendMessage(ctx, message.StatusOK, "操作成功")
	}
}

func (op *OpenPlatform) CountApps(ctx *gin.Context) {
	total, err := op.gorm.CountApps()
	if err != nil {
		SendError(ctx, err)
	} else {
		SendJSON(ctx, total)
	}
}

func (op *OpenPlatform) GetApps() gin.HandlerFunc {
	type params struct {
		model.OrderBy
		model.Pagination
	}
	return func(ctx *gin.Context) {
		ps := &params{}
		err := ctx.Bind(ps)
		if err != nil {
			SendError(ctx, err)
		} else {
			apps, err := op.gorm.GetApps(ps.OrderBy, ps.Pagination)
			if err != nil {
				SendError(ctx, err)
			} else {
				SendJSON(ctx, apps)
			}
		}
	}
}

//
//var GetAppTags Handler = func() gin.HandlerFunc {
//	type app struct {
//		Appid string
//	}
//	return func(ctx *gin.Context) {
//		query := &app{}
//		err := ctx.BindQuery(query)
//		if err != nil {
//			ctx.AbortWithError(http.StatusBadRequest, err)
//		} else {
//			tags, err := OpenPlatform.GetAppUserTags(query.Appid)
//			if err != nil {
//				SendMessage(ctx, Error, err.Error())
//			} else {
//				SendJSON(ctx, tags)
//			}
//		}
//	}
//}
//
//var CrateAppTag Handler = func() gin.HandlerFunc {
//	type app struct {
//		Appid string
//		Tag   string
//	}
//	return func(ctx *gin.Context) {
//		form := &app{}
//		err := ctx.Bind(form)
//		if err != nil {
//			ctx.AbortWithError(http.StatusBadRequest, err)
//		} else {
//			tag, err := OpenPlatform.CreateAppUserTag(form.Appid, form.Tag)
//			if err != nil {
//				SendMessage(ctx, Error, err.Error())
//			} else {
//				SendJSON(ctx, tag)
//			}
//		}
//	}
//}
//
//var DelAppTag Handler = func() gin.HandlerFunc {
//	type app struct {
//		Appid string
//		TagID int
//	}
//	return func(ctx *gin.Context) {
//		query := &app{}
//		err := ctx.BindQuery(query)
//		if err != nil {
//			ctx.AbortWithError(http.StatusBadRequest, err)
//		} else {
//			err := OpenPlatform.DeleteAppUserTag(query.Appid, query.TagID)
//			if err != nil {
//				SendMessage(ctx, Error, err.Error())
//			} else {
//				SendMessage(ctx, Success, "已删除")
//			}
//		}
//	}
//}
//
//var UpdateAppTag Handler = func() gin.HandlerFunc {
//	type app struct {
//		Appid string
//		wpt.UserTag
//	}
//	return func(ctx *gin.Context) {
//		form := &app{}
//		err := ctx.Bind(form)
//		if err != nil {
//			ctx.AbortWithError(http.StatusBadRequest, err)
//		} else {
//			err := OpenPlatform.UpdateAppUserTag(form.Appid, &form.UserTag)
//			if err != nil {
//				SendMessage(ctx, Error, err.Error())
//			} else {
//				SendMessage(ctx, Success, "操作成功")
//			}
//		}
//	}
//}

func (op *OpenPlatform) GetUserSummary() gin.HandlerFunc {
	type params struct {
		Appids    []string
		BeginDate string
		EndDate   string
	}
	return func(ctx *gin.Context) {
		ps := &params{}
		err := ctx.Bind(ps)
		if err != nil {
			SendError(ctx, err)
		} else {
			//summary, err := randomTotalSummaries(ps.BeginDate, ps.EndDate)
			//if err != nil {
			//	SendError(ctx, err)
			//} else {
			//	SendJSON(ctx, summary)
			//}
			summary, err := op.pt.GetUserSummary(nil, ps.BeginDate, ps.EndDate)
			if err != nil {
				SendError(ctx, err)
			} else {
				SendJSON(ctx, summary)
			}
		}
	}
}

// 生成随机统计数据，便于测试
func randomTotalSummaries(dateFrom, dateTo string) (summary *platform.Summary, err error) {
	appSummaries, err := randomAppSummaries(20+rand.Intn(80), dateFrom, dateTo)
	if err != nil {
		return nil, err
	}
	return platform.CountTotalSummary(appSummaries), nil
}

func randomAppSummaries(appNum int, dateFrom, dateTo string) (appSummaries []*platform.AppSummary, err error) {
	for i := 1; i <= appNum; i++ {
		userSummaries, err := randomSummaries(dateFrom, dateTo)
		if err != nil {
			return nil, err
		}
		appSummaries = append(appSummaries, &platform.AppSummary{
			Appid:     fmt.Sprintf("appid_%4d", i),
			Nickname:  fmt.Sprintf("nickname_%4d", i),
			Summaries: userSummaries,
		})
	}
	return appSummaries, nil
}

func randomSummaries(dateFrom, dateTo string) ([]*wpt.UserSummary, error) {
	startDate, err := time.Parse("2006-01-02", dateFrom)
	if err != nil {
		return nil, err
	}
	endDate, err := time.Parse("2006-01-02", dateTo)
	if err != nil {
		return nil, err
	}
	var summaries []*wpt.UserSummary
	for {
		if startDate.Before(endDate) || startDate.Equal(endDate) {
			summaries = append(summaries, &wpt.UserSummary{
				RefDate:      startDate.Format("2006-01-02"),
				UserSource:   0,                        // 用户的渠道，数值代表的含义如下： 0代表其他合计 1代表公众号搜索 17代表名片分享 30代表扫描二维码 51代表支付后关注（在支付完成页） 57代表文章内账号名称 100微信广告 161他人转载 176 专辑页内账号名称
				NewUser:      rand.Intn(10000),         // 新增的用户数量
				CancelUser:   rand.Intn(1000),          // 取消关注的用户数量，new_user减去cancel_user即为净增用户数量
				CumulateUser: 10000 + rand.Intn(10000), // 总用户量
			})
			startDate = startDate.AddDate(0, 0, 1)
		} else {
			break
		}
	}
	return summaries, nil
}

func (op *OpenPlatform) GetUserCumulate() gin.HandlerFunc {
	type params struct {
		BeginDate string
		EndDate   string
	}
	return func(ctx *gin.Context) {
		ps := &params{}
		err := ctx.BindQuery(ps)
		if err != nil {
			SendError(ctx, err)
		} else {
			summary, err := op.pt.GetUserCumulate(ps.BeginDate, ps.EndDate)
			if err != nil {
				SendError(ctx, err)
			} else {
				SendJSON(ctx, summary)
			}
		}
	}
}
