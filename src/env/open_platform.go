package env

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/morgine/log"
	"github.com/morgine/pkg/config"
	"github.com/morgine/songs/src/model"
	"github.com/morgine/wechat_sdk/src"
	"gorm.io/gorm"
	"time"
)

func initOpenClientConfigs(namespace string, appDB *gorm.DB, accessRds *redis.Client, configs config.Configs) (*src.OpenClientConfigs, error) {
	env := &OpenPlatformEnv{}
	err := configs.UnmarshalSub(namespace, env)
	if err != nil {
		return nil, err
	}
	componentStorage := newComponentStorage(env.Appid, accessRds)
	appStorage := &appStorage{db: appDB}
	return &src.OpenClientConfigs{
		Appid:            env.Appid,
		Secret:           env.AppSecret,
		MsgVerifyToken:   env.MessageVerifyToken,
		AesKey:           env.EncodingAesKey,
		AesToken:         env.AesToken,
		ComponentStorage: componentStorage,
		AppStorage:       appStorage,
		Logger:           log.Error,
	}, nil
	//msgCrypt, err := wpt.NewWXBizMsgCrypt(env.AesToken, env.EncodingAesKey, env.Appid)
	//if err != nil {
	//	return nil, err
	//}
	//appStore := newOpenPlatformStorage(accessStore, m.App())
	//return platform.initOpenClientConfigs(env.Appid, env.AppSecret, appStore, env.MessageVerifyToken, msgCrypt), nil
}

type accessStorage struct {
	client *redis.Client
}

func newComponentStorage(componentAppid string, client *redis.Client) src.ComponentStorage {
	return src.NewComponentStorage(componentAppid, &accessStorage{client: client})
}

var noCtx = context.Background()

func (a *accessStorage) Set(key string, value []byte, expiration time.Duration) error {
	return a.client.Set(noCtx, key, string(value), expiration).Err()
}

func (a *accessStorage) Get(key string) (value []byte, err error) {
	value, err = a.client.Get(noCtx, key).Bytes()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	return value, nil
}

type appStorage struct {
	db *gorm.DB
}

func (a *appStorage) SaveAppInfo(appid string, info *src.AppInfo) error {
	app := &model.App{}
	err := a.db.Where("appid=?", appid).Select("id").First(app).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	app.Appid = appid
	app.NickName = info.NickName
	app.HeadImg = info.HeadImg
	app.UserName = info.UserName
	app.PrincipalName = info.PrincipalName
	app.Alias = info.Alias
	app.QrcodeUrl = info.QrcodeUrl
	app.Signature = info.Signature
	return a.db.Save(app).Error
}

func (a *appStorage) GetAppInfo(appid string) (*src.AppInfo, error) {
	app := &model.App{}
	err := a.db.Where("appid=?", appid).Select("id").First(app).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if app.ID > 0 {
		return &src.AppInfo{
			Appid:         app.Appid,
			NickName:      app.NickName,
			HeadImg:       app.HeadImg,
			UserName:      app.UserName,
			PrincipalName: app.PrincipalName,
			Alias:         app.Alias,
			QrcodeUrl:     app.QrcodeUrl,
			Signature:     app.Signature,
		}, nil
	} else {
		return nil, nil
	}
}

func (a *appStorage) DelAppInfo(appid string) error {
	return a.db.Where("appid=?", appid).Delete(&model.App{}).Error
}

func (a *appStorage) DelAppInfoNotIn(appids []string) error {
	return a.db.Not("appid in (?)", appids).Delete(&model.App{}).Error
}
