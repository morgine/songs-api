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
	"github.com/morgine/wechat_sdk/pkg/statistics"
	"github.com/morgine/wechat_sdk/pkg/users"
	"github.com/morgine/wechat_sdk/src"
	"gorm.io/gorm"
	"net/http"
	"path/filepath"
	"sync"
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
	materialClient  *cache.MaterialClient
	miniProgramCard *cache.MiniProgramCardClient
	userTag         *cache.UserTagClient
	appUserTag      *cache.AppUserTagClient
	articleClient   *cache.ArticleClient
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
		materialClient        *cache.MaterialClient
		miniProgramCardClient *cache.MiniProgramCardClient
		userTagClient         *cache.UserTagClient
		appUserTagClient      *cache.AppUserTagClient
		articleClient         *cache.ArticleClient
	)
	engine := cache.NewRedisEngine(rds)
	tempMaterialClient = cache.NewTempMaterialClient("temp_material_", engine)
	materialClient = cache.NewMaterialClient("material_", engine)
	miniProgramCardClient = cache.NewMiniProgramCardClient("mini_p_c_rsp_", engine)
	userTagClient = cache.NewUserTagClient("cpm_user_tag_", engine)
	appUserTagClient = cache.NewAppUserTagClient("app_user_tag_", engine)
	articleClient = cache.NewArticleClient("article_", engine)

	openClient, err := src.NewOpenClient(openClientConfigs)
	if err != nil {
		return nil, err
	}
	op := &OpenPlatform{
		pt:              openClient,
		appModel:        model.NewModel(db).App(),
		tempMaterials:   tempMaterialClient,
		materialClient:  materialClient,
		miniProgramCard: miniProgramCardClient,
		userTag:         userTagClient,
		appUserTag:      appUserTagClient,
		articleClient:   articleClient,
		host:            host,
		uploader:        uploader,
		imageDir:        imageDir,
	}
	openClient.SubscribeEvent(message3.EvtUserSubscribe, func(msg *message3.EventMessage, ctx *src.Context) {
		client := ctx.Client()
		articles, err := articleClient.Get(openClientConfigs.Appid)
		if err != nil {
			log.Error.Println(err)
		} else if len(articles) > 0 {
			var arts []src.Article
			for _, article := range articles {
				media, err := op.getMaterial(client, article.PicFile)
				if err != nil {
					log.Error.Println(err)
				} else {
					arts = append(arts, src.Article{
						Title:       article.Title,
						Description: article.Description,
						Url:         article.Url,
						PicUrl:      media.Url,
					})
				}
			}
			err = ctx.ResponseArticles(arts)
			if err != nil {
				log.Error.Println(err)
			}
		}
		// 发送小程序卡片消息(通过客服消息接口发送)
		// 图片使用公众号临时素材，3 天后过期
		cards, err := miniProgramCardClient.Get(openClientConfigs.Appid)
		if err != nil {
			log.Error.Println(err)
		} else {
			for _, card := range cards {
				if card != nil && card.ThumbMediaFilename != "" {
					mediaID, err := op.getTempMaterialID(client, card.ThumbMediaFilename)
					if err != nil {
						log.Error.Println(err)
					} else {
						page := &message3.MiniProgramPage{
							Title:        card.Title,
							Appid:        card.Appid,
							PagePath:     card.PagePath,
							ThumbMediaID: mediaID,
						}
						err = client.SendMiniProgramPage([]string{ctx.Openid}, page)
						if err != nil {
							log.Error.Println(err)
						}
					}
				}
			}
		}
		// 设置用户标签
		tag, err := op.getUserTag(openClientConfigs.Appid, client)
		if err != nil {
			log.Error.Println(err)
		} else {
			if tag != nil {
				err = client.BatchTagging(tag.ID, []string{ctx.Openid})
				if err != nil {
					log.Error.Println(err)
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

// 统计
type Statistics struct {
	CumulateUser      int     `json:"cumulate_user"`       // 用户总量
	NewUser           int     `json:"new_user"`            // 新增的用户数量
	CancelUser        int     `json:"cancel_user"`         // 取消关注的用户数量，new_user减去cancel_user即为净增用户数量
	PositiveUser      int     `json:"positive_user"`       // 净增用户
	CancelRate        float64 `json:"cancel_rate"`         // 取关率, cancel_user/cumulate_user
	ReqSuccCount      int     `json:"req_succ_count"`      // 拉取量
	ExposureCount     int     `json:"exposure_count"`      // 曝光量
	ExposureRate      float64 `json:"exposure_rate"`       // 曝光率, exposure_count/req_succ_count
	ClickCount        int     `json:"click_count"`         // 点击量
	ClickRate         float64 `json:"click_rate"`          // 点击率, click_count/exposure_count
	Outcome           int     `json:"outcome"`             // 支出(分)
	Income            int     `json:"income"`              // 收入(分)
	IncomeOutcomeRate float64 `json:"income_outcome_rate"` // 收入支出比率, income/outcome
	Ecpm              float64 `json:"ecpm"`                // 广告千次曝光收益(分), 1000/exposure_count*income
}

func (s *Statistics) initRate() {
	if s.CumulateUser > 0 {
		s.CancelRate = float64(s.CancelUser) / float64(s.CumulateUser)
	}
	if s.ReqSuccCount > 0 {
		s.ExposureRate = float64(s.ExposureCount) / float64(s.ReqSuccCount)
	}
	if s.ExposureCount > 0 {
		s.ClickRate = float64(s.ClickCount) / float64(s.ExposureCount)
	}
	if s.Outcome > 0 {
		s.IncomeOutcomeRate = float64(s.Income) / float64(s.Outcome)
	}
	if s.ExposureCount > 0 && s.Income > 0 {
		s.Ecpm = 1000 * float64(s.Income) / float64(s.ExposureCount)
	}
}

// 公众号统计
type AppStatistics struct {
	Appid      string   `json:"appid"`    // 公众号 appid
	Nickname   string   `json:"nickname"` // 公众号昵称
	Errs       []string `json:"errs"`     // 错误
	Statistics          // 统计数据
}

// 日期统计数据
type DateStatistics struct {
	Date               string           `json:"date"`                 // 统计日期
	Data               Statistics       `json:"data"`                 // 总体统计
	TotalExposureCount float64          `json:"total_exposure_count"` // 已曝光量加未曝光量，用于计算曝光率
	TotalClickCount    float64          `json:"total_click_count"`    // 已点击量加未点击量，用于计算点击率
	Apps               []*AppStatistics `json:"apps"`                 // APP 统计
}

func (ds *DateStatistics) initAppData(appid, nickname, err string) *AppStatistics {
	var as *AppStatistics
	for _, app := range ds.Apps {
		if app.Appid == appid {
			as = app
		}
	}
	if as == nil {
		as = &AppStatistics{
			Appid:    appid,
			Nickname: nickname,
		}
		ds.Apps = append(ds.Apps, as)
	}
	if err != "" {
		as.Errs = append(as.Errs, err)
	}
	return as
}

type DatesStatistics []*DateStatistics // 多日期统计数据

func (ds *DatesStatistics) initDateData(date string) *DateStatistics {
	for _, dateStatistics := range *ds {
		if dateStatistics.Date == date {
			return dateStatistics
		}
	}
	d := &DateStatistics{
		Date: date,
		Data: Statistics{},
		Apps: nil,
	}
	*ds = append(*ds, d)
	return d
}

func (ds *DatesStatistics) AppendAppError(appid, nickname string, err error, date time.Time) {
	var dateStatistics *DateStatistics = ds.initDateData(date.Format("2006-01-02"))
	_ = dateStatistics.initAppData(appid, nickname, err.Error())
}

func (ds *DatesStatistics) AddUserStatistics(appid, nickname string, us *src.UserStatistics) {
	if us != nil {
		for _, cumulate := range us.Cumulates {
			var dateStatistics *DateStatistics = ds.initDateData(cumulate.RefDate)
			var appStatistics *AppStatistics = dateStatistics.initAppData(appid, nickname, "")
			appStatistics.CumulateUser = cumulate.CumulateUser
			appStatistics.NewUser = cumulate.NewUser
			appStatistics.CancelUser = cumulate.CancelUser
			appStatistics.PositiveUser = appStatistics.NewUser - appStatistics.CancelUser
			appStatistics.CancelRate = float64(appStatistics.CancelUser) / float64(appStatistics.CumulateUser)
			dateStatistics.Data.CumulateUser += cumulate.CumulateUser
			dateStatistics.Data.NewUser += cumulate.NewUser
			dateStatistics.Data.CancelUser += cumulate.CancelUser
			dateStatistics.Data.PositiveUser = dateStatistics.Data.NewUser - dateStatistics.Data.CancelUser
			dateStatistics.Data.CancelRate += float64(dateStatistics.Data.CancelUser) / float64(dateStatistics.Data.CumulateUser)
		}
	}
}

func (ds *DatesStatistics) AddAdvertStatistics(appid, nickname string, us *statistics.PublisherAdPosGeneralResponse) {
	if us != nil {
		for _, l := range us.List {
			var dateStatistics *DateStatistics = ds.initDateData(l.Date)
			var appStatistics *AppStatistics = dateStatistics.initAppData(appid, nickname, "")
			appStatistics.ReqSuccCount += l.ReqSuccCount   // 拉取量
			appStatistics.ExposureCount += l.ExposureCount // 曝光量
			appStatistics.ClickCount += l.ClickCount       // 点击量
			appStatistics.Income += l.Income               // 收入
			appStatistics.initRate()                       // 比率换算

			dateStatistics.Data.ReqSuccCount += l.ReqSuccCount   // 拉取量
			dateStatistics.Data.ExposureCount += l.ExposureCount // 曝光量
			dateStatistics.Data.ClickCount += l.ClickCount       // 点击量
			dateStatistics.Data.Income += l.Income               // 收入
			dateStatistics.Data.initRate()                       // 比率换算
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
				var datesStatistics DatesStatistics
				wg := sync.WaitGroup{}
				for _, app := range apps {
					wg.Add(1)
					app := app
					go func() {
						client, err := op.pt.GetClient(app.Appid)
						if err != nil {
							datesStatistics.AppendAppError(app.Appid, app.NickName, err, beginDate)
						} else if client != nil {
							// 用户统计
							var us *src.UserStatistics
							us, err = client.GetUserStatistics(beginDate, endDate)
							if err != nil {
								datesStatistics.AppendAppError(app.Appid, app.NickName, err, beginDate)
							} else {
								datesStatistics.AddUserStatistics(app.Appid, app.NickName, us)
							}
							// 广告统计
							var as *statistics.PublisherAdPosGeneralResponse
							page, pageSize := 1, 90
							for {
								as, err = client.GetPublisherAdPosGeneral("", statistics.PublisherCommonOptions{
									Page:      page,
									PageSize:  pageSize,
									StartDate: beginDate,
									EndDate:   endDate,
								})
								if err != nil {
									datesStatistics.AppendAppError(app.Appid, app.NickName, err, beginDate)
									break
								} else {
									datesStatistics.AddAdvertStatistics(app.Appid, app.NickName, as)
									if page*pageSize < as.TotalNum {
										page++
									} else {
										break
									}
								}
							}
						}
						wg.Done()
					}()
				}
				wg.Wait()
				SendJSON(ctx, datesStatistics)
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
		cards, err := op.miniProgramCard.Get(op.pt.Configs().Appid)
		if err != nil {
			SendError(ctx, err)
		} else {
			SendJSON(ctx, cards)
		}
	}
}

func (op *OpenPlatform) SaveMiniProgramCard() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var cards []*cache.MiniProgramCard
		err := ctx.Bind(&cards)
		if err != nil {
			SendError(ctx, err)
		} else {
			err := op.miniProgramCard.Set(op.pt.Configs().Appid, cards)
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

func (op *OpenPlatform) GetArticles() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		articles, err := op.articleClient.Get(op.pt.Configs().Appid)
		if err != nil {
			SendError(ctx, err)
		} else {
			SendJSON(ctx, articles)
		}
	}
}

func (op *OpenPlatform) SaveArticles() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var articles []*cache.Article
		err := ctx.Bind(&articles)
		if err != nil {
			SendError(ctx, err)
		} else {
			err := op.articleClient.Set(op.pt.Configs().Appid, articles)
			if err != nil {
				SendError(ctx, err)
			} else {
				SendMessage(ctx, message.StatusOK, "已保存")
			}
		}
	}
}

func (op *OpenPlatform) DelArticles() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		err := op.articleClient.Del(op.pt.Configs().Appid)
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

func (op *OpenPlatform) getMaterial(client *src.PublicClient, filename string) (*material.UploadedMedia, error) {
	appid := client.GetAppid()
	media, err := op.materialClient.Get(appid, filename)
	if err != nil {
		return nil, err
	}
	if media == nil {
		data, err := op.uploader.GetFile(filename)
		if err != nil {
			return nil, err
		}
		media, err = client.UploadMaterial(material.IMAGE, data, filename, nil)
		if err != nil {
			return nil, err
		} else {
			err = op.materialClient.Set(appid, filename, media)
			if err != nil {
				return nil, err
			}
		}
	}
	return media, nil
}

func (op *OpenPlatform) getTempMaterialID(client *src.PublicClient, filename string) (string, error) {
	appid := client.GetAppid()
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
		tempMedia, err = client.UploadTempMaterial(material.IMAGE, data, filename)
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

func (op *OpenPlatform) getUserTag(componentAppid string, client *src.PublicClient) (*users.Tag, error) {
	userTag, err := op.userTag.Get(componentAppid)
	if err != nil {
		return nil, err
	}
	if userTag.Name == "" {
		return nil, nil
	}
	if userTag != nil {
		appid := client.GetAppid()
		appTag, err := op.appUserTag.Get(appid, userTag.Name)
		if err != nil {
			return nil, err
		}
		if appTag == nil {
			tags, err := client.GetAppUserTags()
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
					tag, err = client.CreateAppUserTag(userTag.Name)
					return err
				}()
				if err != nil {
					return nil, err
				}
				err = op.appUserTag.Set(appid, userTag.Name, tag)
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
