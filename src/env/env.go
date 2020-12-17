package env

import (
	"github.com/gin-gonic/gin"
	admin2 "github.com/morgine/pkg/admin"
	"github.com/morgine/pkg/config"
	"github.com/morgine/pkg/database/orm"
	"github.com/morgine/pkg/redis"
	"github.com/morgine/pkg/session"
	"github.com/morgine/songs/src/handler"
)

type ServerEnv struct {
	Host   string `toml:"host"`
	Upload string `toml:"upload"`
}

type OpenPlatformEnv struct {
	AesToken           string `toml:"aes_token"`
	EncodingAesKey     string `toml:"encoding_aes_key"`
	AppSecret          string `toml:"app_secret"`
	Appid              string `toml:"appid"`
	MessageVerifyToken string `toml:"message_verify_token"`
}

type AdvertPlatformEnv struct {
	ClientID string `toml:"client_id"`
	Secret   string `toml:"secret"`
}

type Handlers struct {
	Admin          *handler.Admin
	OpenPlatform   *handler.OpenPlatform
	AdvertPlatform *handler.AdvertPlatform
	Proxy          *handler.Proxy
}

func LoadEnv(configFile string) (handlers *Handlers, release func() error) {

	// 初始化配置服务
	var configs, err = config.UnmarshalFile(configFile)
	if err != nil {
		panic(err)
	}
	serverEnv := &ServerEnv{}
	err = configs.UnmarshalSub("server", serverEnv)
	if err != nil {
		panic(err)
	}
	orm, err := orm.NewPostgresORM("postgres", "gorm", configs)
	if err != nil {
		panic(err)
	}
	accessRedisClient, err := redis.NewClient("access-token-redis", configs)
	if err != nil {
		panic(err)
	}
	cacheRedisClient, err := redis.NewClient("cache-redis", configs)
	if err != nil {
		panic(err)
	}
	openClientConfigs, err := initOpenClientConfigs("open-platform", orm, accessRedisClient, configs)
	if err != nil {
		panic(err)
	}
	advertAccessStore := newAccessTokenStorage("ad_", accessRedisClient)
	advertPlatform, err := NewAdvertPlatform("advert-platform", advertAccessStore, configs)
	if err != nil {
		panic(err)
	}

	adminRedisClient, err := redis.NewClient("admin-redis", configs)
	if err != nil {
		panic(err)
	}
	admin, err := handler.NewAdmin(&admin2.Options{
		DB:          orm,
		Session:     session.NewRedisStorage("admin_", adminRedisClient),
		AuthExpires: 86400,
		AesCryptKey: []byte("change this pass"),
	})
	if err != nil {
		panic(err)
	}
	uploadHandlers, err := handler.NewMultiFileHandlers(
		orm,
		serverEnv.Upload,
		func(ctx *gin.Context) (userID int, ok bool) {
			return admin.GetLoginAdminID(ctx)
		},
	)
	openPlatformHandler, err := handler.NewOpenPlatform(
		openClientConfigs,
		orm,
		cacheRedisClient,
		serverEnv.Host,
		serverEnv.Upload,
		uploadHandlers,
	)
	if err != nil {
		panic(err)
	}
	advertPlatformHandler := handler.NewAdvertPlatform(advertPlatform, serverEnv.Host)
	return &Handlers{
			Admin:          admin,
			Proxy:          &handler.Proxy{},
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
