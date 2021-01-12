package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/morgine/log"
	"github.com/morgine/songs/src/cache"
	"github.com/morgine/songs/src/message"
	"github.com/morgine/songs/src/model"
	message2 "github.com/morgine/wechat_sdk/pkg/message"
	"github.com/morgine/wechat_sdk/src"
	"gorm.io/gorm"
)

type subscribeMessageStorage struct {
	engine cache.Engine
}

func (s *subscribeMessageStorage) SetDefaultMessages(msgs *model.SubscribeMessages) error {
	return s.setMessages("def", msgs)
}

func (s *subscribeMessageStorage) GetDefaultMessages() (msgs *model.SubscribeMessages, err error) {
	return s.getMessages("def")
}

func (s *subscribeMessageStorage) DelDefaultMessages() error {
	return s.delMessages("def")
}

func (s *subscribeMessageStorage) setMessages(key string, msgs *model.SubscribeMessages) error {
	data, err := json.Marshal(msgs)
	if err != nil {
		return err
	}
	return s.engine.Set(key, data, 0)
}

func (s *subscribeMessageStorage) getMessages(key string) (msgs *model.SubscribeMessages, err error) {
	data, err := s.engine.Get(key)
	if err != nil {
		return nil, err
	}
	if data != nil {
		msgs = &model.SubscribeMessages{}
		err = json.Unmarshal(data, msgs)
		if err != nil {
			return nil, err
		}
	}
	return
}

func (s *subscribeMessageStorage) delMessages(key string) error {
	return s.engine.Del(key)
}

func (s *subscribeMessageStorage) SetAppMessages(appid string, msgs *model.SubscribeMessages) error {
	return s.setMessages("app_"+appid, msgs)
}

func (s *subscribeMessageStorage) GetAppMessages(appid string) (msgs *model.SubscribeMessages, err error) {
	return s.getMessages("app_" + appid)
}

func (s *subscribeMessageStorage) DelAppMessages(appid string) error {
	return s.delMessages("app_" + appid)
}

type SubscribeMessage struct {
	db *model.SubscribeMessageDB
	op *OpenPlatform
}

func NewSubscribeMessage(op *OpenPlatform, rds *redis.Client, db *gorm.DB) *SubscribeMessage {
	cacheEngine := cache.NewRedisEngine(rds)
	msgDB := model.NewSubscribeMessageDB(db, &subscribeMessageStorage{engine: cacheEngine})
	op.pt.SubscribeEvent(message2.EvtUserSubscribe, func(msg *message2.EventMessage, ctx *src.Context) {
		client := ctx.Client()
		appid := client.GetAppid()
		msgs, err := msgDB.GetAppSubscribeMessages(appid)
		if err != nil {
			log.Error.Println(err)
			return
		}
		if msgs == nil {
			msgs, err = msgDB.GetDefaultSubscribeMessages()
			if err != nil {
				log.Error.Println(err)
				return
			}
		}

		if msgs != nil && len(msgs.Articles) > 0 || len(msgs.Cards) > 0 {
			if len(msgs.Articles) > 0 {
				var arts []src.Article
				for _, article := range msgs.Articles {
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
			if len(msgs.Cards) > 0 {
				for _, card := range msgs.Cards {
					if card != nil && card.ThumbMediaFilename != "" {
						mediaID, err := op.getTempMaterialID(client, card.ThumbMediaFilename)
						if err != nil {
							log.Error.Println(err)
						} else {
							page := &message2.MiniProgramPage{
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
		}
	})
	return &SubscribeMessage{
		db: msgDB,
		op: op,
	}
}

func (sm *SubscribeMessage) GetGroups() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		SendJSON(ctx, sm.db.GetGroups())
	}
}

func (sm *SubscribeMessage) SaveGroup() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ps := &model.SubscribeMessageGroup{}
		err := ctx.Bind(ps)
		if err != nil {
			SendError(ctx, err)
		} else {
			err = sm.db.SaveGroup(ps)
			if err != nil {
				SendError(ctx, err)
			} else {
				SendJSON(ctx, ps)
			}
		}
	}
}

func (sm *SubscribeMessage) DeleteGroup() gin.HandlerFunc {
	type params struct {
		ID int
	}
	return func(ctx *gin.Context) {
		ps := &params{}
		err := ctx.Bind(ps)
		if err != nil {
			SendError(ctx, err)
		} else {
			err = sm.db.DeleteGroup(ps.ID)
			if err != nil {
				SendError(ctx, err)
			} else {
				SendMessage(ctx, message.StatusOK, "已删除")
			}
		}
	}
}

func (sm *SubscribeMessage) GetGroupArticles() gin.HandlerFunc {
	type params struct {
		GroupID int
	}
	return func(ctx *gin.Context) {
		ps := &params{}
		err := ctx.Bind(ps)
		if err != nil {
			SendError(ctx, err)
		} else {
			msgs, err := sm.db.GetArticleMessages(ps.GroupID)
			if err != nil {
				SendError(ctx, err)
			} else {
				SendJSON(ctx, msgs)
			}
		}
	}
}

func (sm *SubscribeMessage) SaveGroupArticle() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ps := &model.SubscribeArticleMessage{}
		err := ctx.Bind(ps)
		if err != nil {
			SendError(ctx, err)
		} else {
			err = sm.db.SaveArticleMessage(ps)
			if err != nil {
				SendError(ctx, err)
			} else {
				SendJSON(ctx, ps)
			}
		}
	}
}

func (sm *SubscribeMessage) DeleteGroupArticle() gin.HandlerFunc {
	type params struct {
		ID int
	}
	return func(ctx *gin.Context) {
		ps := &params{}
		err := ctx.Bind(ps)
		if err != nil {
			SendError(ctx, err)
		} else {
			err = sm.db.DeleteArticleMessage(ps.ID)
			if err != nil {
				SendError(ctx, err)
			} else {
				SendMessage(ctx, message.StatusOK, "已删除")
			}
		}
	}
}

func (sm *SubscribeMessage) GetGroupMiniProgramCards() gin.HandlerFunc {
	type params struct {
		GroupID int
	}
	return func(ctx *gin.Context) {
		ps := &params{}
		err := ctx.Bind(ps)
		if err != nil {
			SendError(ctx, err)
		} else {
			msgs, err := sm.db.GetMiniProgramMessages(ps.GroupID)
			if err != nil {
				SendError(ctx, err)
			} else {
				SendJSON(ctx, msgs)
			}
		}
	}
}

func (sm *SubscribeMessage) SaveGroupMiniProgramCard() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ps := &model.SubscribeMiniProgramMessage{}
		err := ctx.Bind(ps)
		if err != nil {
			SendError(ctx, err)
		} else {
			err = sm.db.SaveMiniProgramMessage(ps)
			if err != nil {
				SendError(ctx, err)
			} else {
				SendJSON(ctx, ps)
			}
		}
	}
}

func (sm *SubscribeMessage) DeleteGroupMiniProgramCard() gin.HandlerFunc {
	type params struct {
		ID int
	}
	return func(ctx *gin.Context) {
		ps := &params{}
		err := ctx.Bind(ps)
		if err != nil {
			SendError(ctx, err)
		} else {
			err = sm.db.DeleteMiniProgramMessage(ps.ID)
			if err != nil {
				SendError(ctx, err)
			} else {
				SendMessage(ctx, message.StatusOK, "已删除")
			}
		}
	}
}

func (sm *SubscribeMessage) Apply() gin.HandlerFunc {
	type params struct {
		Appids  []string
		GroupID int
	}
	return func(ctx *gin.Context) {
		ps := &params{}
		err := ctx.Bind(ps)
		if err != nil {
			SendError(ctx, err)
		} else {
			err = sm.db.ApplyApp(ps.GroupID, ps.Appids)
			if err != nil {
				SendError(ctx, err)
			} else {
				SendMessage(ctx, message.StatusOK, "操作成功")
			}
		}
	}
}

func (sm *SubscribeMessage) AppliedApps() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		appids, err := sm.db.GetAppliedApps()
		if err != nil {
			SendError(ctx, err)
		} else {
			SendJSON(ctx, appids)
		}
	}
}

func (sm *SubscribeMessage) getAppSubscribeMsgs(appid string) (msgs *model.SubscribeMessages, isDefault bool, err error) {
	msgs, err = sm.db.GetAppSubscribeMessages(appid)
	if err != nil {
		return nil, false, err
	}
	if msgs != nil && len(msgs.Cards) > 0 || len(msgs.Articles) > 0 {
		return msgs, false, nil
	}
	msgs, err = sm.db.GetDefaultSubscribeMessages()
	if err != nil {
		return nil, false, err
	}
	if msgs != nil && len(msgs.Cards) > 0 || len(msgs.Articles) > 0 {
		return msgs, true, nil
	}
	return nil, false, nil
}

func (sm *SubscribeMessage) ShowAppMsgs() gin.HandlerFunc {
	type params struct {
		Appid string
	}
	return func(ctx *gin.Context) {
		ps := &params{}
		err := ctx.Bind(ps)
		if err != nil {
			SendError(ctx, err)
		} else {
			msgs, err := sm.db.GetAppSubscribeMessages(ps.Appid)
			if err != nil {
				SendError(ctx, err)
			} else {
				SendJSON(ctx, msgs)
			}
		}
	}
}

func (sm *SubscribeMessage) Cancel() gin.HandlerFunc {
	type params struct {
		Appids []string
	}
	return func(ctx *gin.Context) {
		ps := &params{}
		err := ctx.Bind(ps)
		if err != nil {
			SendError(ctx, err)
		} else {
			err = sm.db.CancelApp(ps.Appids)
			if err != nil {
				SendError(ctx, err)
			} else {
				SendMessage(ctx, message.StatusOK, "已取消")
			}
			//var errs []error
			//for _, appid := range ps.Appids {
			//	nerr := sm.cache.Del(appid)
			//	if nerr != nil {
			//		errs = append(errs, nerr)
			//	}
			//}
			//if len(errs) > 0 {
			//	SendErrors(ctx, errs...)
			//} else {
			//	SendMessage(ctx, message.StatusOK, "已取消")
			//}
		}
	}
}
