package env

import (
	"fmt"
	admin2 "github.com/morgine/pkg/admin"
	"github.com/morgine/pkg/config"
	"github.com/morgine/pkg/database/orm"
	"github.com/morgine/pkg/redis"
	"github.com/morgine/pkg/session"
	"github.com/morgine/songs/src/handler"
	"github.com/morgine/songs/src/model"
	"gorm.io/gorm"
)

type OpenPlatformEnv struct {
	AesToken       string `toml:"aes_token"`
	EncodingAesKey string `toml:"encoding_aes_key"`
	AppSecret      string `toml:"app_secret"`
	Appid          string `toml:"appid"`
}

type AdvertPlatformEnv struct {
	ClientID string `toml:"client_id"`
	Secret   string `toml:"secret"`
}

type Handlers struct {
	Admin          *admin2.Handler
	OpenPlatform   *handler.OpenPlatform
	AdvertPlatform *handler.AdvertPlatform
}

func LoadEnv(configFile string) (handlers *Handlers, release func() error) {

	// 初始化配置服务
	var configs, err = config.UnmarshalFile(configFile)
	if err != nil {
		panic(err)
	}
	orm, err := orm.NewPostgresORM("postgres", "gorm", configs)
	if err != nil {
		panic(err)
	}
	//createTestApps(orm)
	accessRedisClient, err := redis.NewClient("access-token-redis", configs)
	if err != nil {
		panic(err)
	}
	//var m = model.NewGorm(orm)
	//openAccessStore := newAccessTokenStorage("app_", accessRedisClient)
	//openPlatform, err := NewOpenPlatform("open-platform", openAccessStore, m, configs)
	//if err != nil {
	//	panic(err)
	//}
	advertAccessStore := newAccessTokenStorage("ad_", accessRedisClient)
	advertPlatform, err := NewAdvertPlatform("advert-platform", advertAccessStore, configs)
	if err != nil {
		panic(err)
	}

	adminRedisClient, err := redis.NewClient("admin-redis", configs)
	if err != nil {
		panic(err)
	}
	admin, err := admin2.NewHandler(&admin2.Options{
		DB:          orm,
		Session:     session.NewRedisStorage("admin_", adminRedisClient),
		AuthExpires: 86400,
		AesCryptKey: []byte("change this pass"),
		Sender:      adminSender(0),
	})
	if err != nil {
		panic(err)
	}
	//openPlatformHandler, err := handler.NewOpenPlatform(openPlatform, orm)
	openPlatformHandler, err := handler.NewOpenPlatform(nil, orm)
	if err != nil {
		panic(err)
	}
	advertPlatformHandler := handler.NewAdvertPlatform(advertPlatform)
	return &Handlers{
			Admin:          admin,
			OpenPlatform:   openPlatformHandler,
			AdvertPlatform: advertPlatformHandler,
		}, func() error {
			db, _ := orm.DB()
			db.Close()
			accessRedisClient.Close()
			adminRedisClient.Close()
			return nil
		}
}

func createTestApps(db *gorm.DB) {
	for i := 0; i < 99; i++ {
		app := &model.App{
			ID:            0,
			Appid:         fmt.Sprintf("appid%3d", i),
			NickName:      fmt.Sprintf("nickname%3d", i),
			HeadImg:       fmt.Sprintf("headImg%3d", i),
			UserName:      fmt.Sprintf("UserName%3d", i),
			PrincipalName: fmt.Sprintf("PrincipalName%3d", i),
			Alias:         fmt.Sprintf("Alias%3d", i),
			QrcodeUrl:     fmt.Sprintf("QrcodeUrl%3d", i),
			Signature:     fmt.Sprintf("Signature%3d", i),
		}
		db.Create(app)
	}
}