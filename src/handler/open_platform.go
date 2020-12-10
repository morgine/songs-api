package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/morgine/log"
	"github.com/morgine/pkg/upload"
	model2 "github.com/morgine/pkg/upload/model"
	"github.com/morgine/songs/pkg/xtime"
	"github.com/morgine/songs/src/cache"
	"github.com/morgine/songs/src/message"
	"github.com/morgine/songs/src/model"
	"github.com/morgine/wechat_sdk/pkg/material"
	message3 "github.com/morgine/wechat_sdk/pkg/message"
	"github.com/morgine/wechat_sdk/pkg/users"
	"github.com/morgine/wechat_sdk/src"
	"gorm.io/gorm"
	"net/http"
	"path/filepath"
	"time"
)

const (
	MsgPictureBig model2.Kind = iota + 900
	MsgPictureSmall
)

type OpenPlatform struct {
	pt              *src.OpenClient
	appModel        *model.AppModel
	tempMaterials   *cache.TempMaterialClient
	miniProgramCard *cache.MiniProgramCardClient
	userTag         *cache.UserTagClient
	appUserTag      *cache.AppUserTagClient
	uploader        *upload.MultiFileHandlers
	host            string
	imageDir        string
}

func NewOpenPlatform(openClientConfigs *src.OpenClientConfigs, db *gorm.DB, rds *redis.Client, host, imageDir string, uploader *upload.MultiFileHandlers) (*OpenPlatform, error) {
	err := db.AutoMigrate(
		&model.App{},
	)
	if err != nil {
		return nil, err
	}
	var (
		tempMaterialClient    *cache.TempMaterialClient
		miniProgramCardClient *cache.MiniProgramCardClient
		userTagClient         *cache.UserTagClient
		appUserTagClient      *cache.AppUserTagClient
	)
	engine := cache.NewRedisEngine(rds)
	tempMaterialClient = cache.NewTempMaterialClient("temp_material_", engine)
	miniProgramCardClient = cache.NewMiniProgramCardClient("mini_p_c_rsp_", engine)
	userTagClient = cache.NewUserTagClient("cpm_user_tag_", engine)
	appUserTagClient = cache.NewAppUserTagClient("app_user_tag_", engine)

	openClient, err := src.NewOpenClient(openClientConfigs)
	if err != nil {
		return nil, err
	}
	op := &OpenPlatform{
		pt:              openClient,
		appModel:        model.NewModel(db).App(),
		tempMaterials:   tempMaterialClient,
		miniProgramCard: miniProgramCardClient,
		userTag:         userTagClient,
		appUserTag:      appUserTagClient,
		host:            host,
		uploader:        uploader,
		imageDir:        imageDir,
	}
	openClient.SubscribeEvent(message3.EvtUserSubscribe, func(msg *message3.EventMessage, ctx *src.Context) {
		appInfo, err := ctx.AppInfo()
		if err != nil {
			log.Error.Println(err)
		} else {
			// 发送小程序卡片消息(通过客服消息接口发送)
			// 图片使用公众号临时素材，3 天后过期
			miniRsp, err := miniProgramCardClient.Get(openClientConfigs.Appid)
			if err != nil {
				log.Error.Println(err)
			} else {
				if miniRsp != nil && miniRsp.ThumbMediaFilename != "" {
					mediaID, err := op.getMaterialID(appInfo.Appid, miniRsp.ThumbMediaFilename)
					if err != nil {
						log.Error.Println(err)
					} else {
						page := &message3.MiniProgramPage{
							Title:        miniRsp.Title,
							Appid:        appInfo.Appid,
							PagePath:     miniRsp.PagePath,
							ThumbMediaID: mediaID,
						}
						responser := ctx.CustomerMsgResponser()
						err = responser.ResponseMiniProgramPage(page)
						if err != nil {
							log.Error.Println(err)
						}
					}
				}
			}
			// 设置用户标签
			tag, err := op.getUserTag(openClientConfigs.Appid, appInfo.Appid)
			if err != nil {
				log.Error.Println(err)
			} else {
				if tag != nil {
					client, err := op.pt.GetClient(appInfo.Appid)
					if err != nil {
						log.Error.Println(err)
					} else {
						err = client.BatchTagging(tag.ID, []string{ctx.Openid})
						if err != nil {
							log.Error.Println(err)
						}
					}
				}
			}
		}
	})
	return op, nil
}

func (op *OpenPlatform) ListenMessage(appidGetter func(ctx *gin.Context) string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取 APPID
		appid := appidGetter(ctx)
		if appid == "" {
			log.Error.Println("appid not provided")
		} else {
			op.pt.ListenMessage(appid, ctx.Writer, ctx.Request)
		}
	}
}

func (op *OpenPlatform) ListenVerifyTicket(ctx *gin.Context) {
	op.pt.ListenVerifyTicket(ctx.Writer, ctx.Request)
}

func (op *OpenPlatform) ComponentLoginPage(redirectUrl string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		uri, err := op.pt.ComponentLoginPage(resetHostUrl(op.host, redirectUrl))
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

func (op *OpenPlatform) ListenLoginPage(ctx *gin.Context) {
	err := op.pt.ListenLoginPage(ctx.Request)
	if err != nil {
		SendMessage(ctx, message.StatusOK, "授权失败："+err.Error())
	} else {
		SendMessage(ctx, message.StatusOK, "授权成功")
	}
}

func (op *OpenPlatform) MigrateApps(ctx *gin.Context) {
	err := op.pt.MigrateApps()
	if err != nil {
		SendMessage(ctx, message.StatusOK, "操作失败："+err.Error())
	} else {
		SendMessage(ctx, message.StatusOK, "操作成功")
	}
}

func (op *OpenPlatform) CountApps(ctx *gin.Context) {
	total, err := op.appModel.CountApps()
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
			apps, err := op.appModel.GetApps(ps.OrderBy, ps.Pagination)
			if err != nil {
				SendError(ctx, err)
			} else {
				SendJSON(ctx, apps)
			}
		}
	}
}

func (op *OpenPlatform) DelApps() gin.HandlerFunc {
	type params struct {
		IDs []int
	}
	return func(ctx *gin.Context) {
		ps := &params{}
		err := ctx.Bind(ps)
		if err != nil {
			SendError(ctx, err)
		} else {
			err := op.appModel.DelApps(model.InIDs(ps.IDs))
			if err != nil {
				SendError(ctx, err)
			} else {
				SendMessage(ctx, message.StatusOK, "已删除")
			}
		}
	}
}

func (op *OpenPlatform) GetUserStatistics() gin.HandlerFunc {
	type params struct {
		model.Pagination
		BeginDate string
		EndDate   string
	}
	return func(ctx *gin.Context) {
		ps := &params{}
		err := ctx.Bind(ps)
		if err != nil {
			SendError(ctx, err)
		} else {
			apps, err := op.appModel.GetApps(ps.Pagination)
			if err != nil {
				SendError(ctx, err)
			} else {
				var appInfos = make([]*src.AppInfo, len(apps))
				for i, app := range apps {
					appInfos[i] = &src.AppInfo{
						Appid:    app.Appid,
						NickName: app.NickName,
					}
				}
				beginDate, err := time.Parse("2006/01/02", ps.BeginDate)
				if err != nil {
					SendError(ctx, err)
					return
				}
				endDate, err := time.Parse("2006/01/02", ps.EndDate)
				if err != nil {
					SendError(ctx, err)
					return
				}
				statistics := op.pt.GetUserStatistics(appInfos, beginDate, endDate)
				SendJSON(ctx, statistics)
			}
		}
	}
}

func (op *OpenPlatform) UploadImages(kind model2.Kind) gin.HandlerFunc {
	opts := upload.CreateMultiFileOpts{
		Kind:    kind,
		PostKey: "picture",
		Success: func(fs []*model2.MultiFile, c *gin.Context) {
			SendJSON(c, fs)
		},
	}
	return op.uploader.CreateMultiFile(opts)
}

func (op *OpenPlatform) CountImages(kind model2.Kind) gin.HandlerFunc {
	opts := upload.CountMultiFilesOpts{
		Kind: kind,
		Success: func(total int64, ctx *gin.Context) {
			SendJSON(ctx, total)
		},
	}
	return op.uploader.CountMultiFiles(opts)
}

func (op *OpenPlatform) GetImages(kind model2.Kind) gin.HandlerFunc {
	type params struct {
		model2.OrderBy
		model2.Pagination
	}
	opts := upload.GetMultiFilesOpts{
		Kind: kind,
		Params: func(ctx *gin.Context) (model2.OrderBy, model2.Pagination, error) {
			ps := &params{}
			err := ctx.Bind(ps)
			return ps.OrderBy, ps.Pagination, err
		},
		Success: func(fs []*model2.MultiFile, ctx *gin.Context) {
			SendJSON(ctx, fs)
		},
	}
	return op.uploader.GetMultiFiles(opts)
}

// 临时文件服务地址
func (op *OpenPlatform) ServeTempImage(
	urlGetter func(host, file string) (url string, err error),
	serveFileParam func(ctx *gin.Context) (file string),
) gin.HandlerFunc {
	err := op.uploader.SetServeUrlGetters(func(file string) (url string, err error) {
		return urlGetter(op.host, file)
	})
	if err != nil {
		panic(err)
	}
	return func(ctx *gin.Context) {
		file := serveFileParam(ctx)
		http.ServeFile(ctx.Writer, ctx.Request, filepath.Join(op.imageDir, file))
	}
}

func (op *OpenPlatform) DelImages(kind model2.Kind) gin.HandlerFunc {
	type params struct {
		IDs []int
	}
	opts := upload.DelMultiFilesOpts{
		Kind: kind,
		Params: func(ctx *gin.Context) (fileIDs []int, err error) {
			ps := &params{}
			err = ctx.Bind(ps)
			if err != nil {
				return nil, err
			} else {
				return ps.IDs, nil
			}
		},
		Success: func(leftTotal int64, ctx *gin.Context) {
			SendJSON(ctx, leftTotal)
		},
	}
	return op.uploader.DelMultiFiles(opts)
}

func (op *OpenPlatform) GetMiniProgramCard() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		card, err := op.miniProgramCard.Get(op.pt.Configs().Appid)
		if err != nil {
			SendError(ctx, err)
		} else {
			if card == nil {
				card = &cache.MiniProgramCard{}
			}
			SendJSON(ctx, card)
		}
	}
}

func (op *OpenPlatform) SaveMiniProgramCard() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		card := &cache.MiniProgramCard{}
		err := ctx.Bind(card)
		if err != nil {
			SendError(ctx, err)
		} else {
			err := op.miniProgramCard.Set(op.pt.Configs().Appid, card)
			if err != nil {
				SendError(ctx, err)
			} else {
				SendMessage(ctx, message.StatusOK, "已保存")
			}
		}
	}
}

func (op *OpenPlatform) DelMiniProgramCard() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		err := op.miniProgramCard.Del(op.pt.Configs().Appid)
		if err != nil {
			SendError(ctx, err)
		} else {
			SendMessage(ctx, message.StatusOK, "已删除")
		}
	}
}

func (op *OpenPlatform) GetUserTag() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		card, err := op.userTag.Get(op.pt.Configs().Appid)
		if err != nil {
			SendError(ctx, err)
		} else {
			if card == nil {
				card = &cache.UserTag{}
			}
			SendJSON(ctx, card)
		}
	}
}

func (op *OpenPlatform) SaveUserTag() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		card := &cache.UserTag{}
		err := ctx.Bind(card)
		if err != nil {
			SendError(ctx, err)
		} else {
			err := op.userTag.Set(op.pt.Configs().Appid, card)
			if err != nil {
				SendError(ctx, err)
			} else {
				SendMessage(ctx, message.StatusOK, "已保存")
			}
		}
	}
}

func (op *OpenPlatform) getMaterialID(appid, filename string) (string, error) {
	tempMedia, err := op.tempMaterials.Get(appid, filename)
	if err != nil {
		return "", err
	}
	now := xtime.Now().Unix()
	if tempMedia == nil || tempMedia.CreatedAt+tempMedia.ExpireIn < now {
		data, err := op.uploader.GetFile(filename)
		if err != nil {
			return "", err
		}
		publicClient, err := op.pt.GetClient(appid)
		if err != nil {
			return "", err
		}
		tempMedia, err = publicClient.UploadTempMaterial(material.IMAGE, data, filename)
		if err != nil {
			return "", err
		} else {
			tempMedia.ExpireIn -= tempMedia.ExpireIn >> 3
			err = op.tempMaterials.Set(appid, filename, tempMedia)
			if err != nil {
				return "", err
			}
		}
	}
	return tempMedia.MediaID, nil
}

func (op *OpenPlatform) getUserTag(componentAppid, appid string) (*users.Tag, error) {
	userTag, err := op.userTag.Get(componentAppid)
	if err != nil {
		return nil, err
	}
	if userTag.Name == "" {
		return nil, nil
	}
	if userTag != nil {
		appTag, err := op.appUserTag.Get(appid)
		if err != nil {
			return nil, err
		}
		if appTag == nil {
			publicClient, err := op.pt.GetClient(appid)
			if err != nil {
				return nil, err
			}
			tags, err := publicClient.GetAppUserTags()
			if err != nil {
				return nil, err
			} else {
				var tag *users.Tag
				err = func() error {
					for _, tag = range tags {
						if tag.Name == userTag.Name {
							return nil
						}
					}
					tag, err = publicClient.CreateAppUserTag(userTag.Name)
					return err
				}()
				if err != nil {
					return nil, err
				}
				err = op.appUserTag.Set(appid, tag)
				if err != nil {
					return nil, err
				}
				return tag, nil
			}
		} else {
			return appTag, nil
		}
	}
	return nil, nil
}
