package src

import (
	"context"
	"flag"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/morgine/log"
	"github.com/morgine/pkg/admin"
	"github.com/morgine/songs/src/env"
	"github.com/morgine/songs/src/handler"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func Run() {
	// 加载配置文件
	configFile := flag.String("c", "config.toml", "配置文件")
	addr := flag.String("a", ":9879", "监听地址")
	flag.Parse()
	handlers, release := env.LoadEnv(*configFile)
	defer release()
	// 注册默认管理员账号
	registerDefaultAdmin(handlers.Admin)
	// 注册路由
	engine := gin.New()

	// 处理共开放平台验证文件
	//engine.NoRoute(func(ctx *gin.Context) {
	//	path := ctx.Request.URL.Path
	//	if ext := filepath.Ext(path); ext == ".txt" {
	//		http.ServeFile(ctx.Writer, ctx.Request, filepath.Join(*appVerifiesDir, path))
	//	} else {
	//		http.FileServer()
	//		http.NotFound(ctx.Writer, ctx.Request)
	//	}
	//})
	v1Path := "/v1"
	listenMsgRoute := "/listen-message/:appid"
	appidGetter := func(ctx *gin.Context) string {
		return ctx.Param("appid")
	}

	engine.Use(WithSkipHandler(func(ctx *gin.Context) (skip bool) {
		return strings.Contains(ctx.Request.URL.Path, "listen-message") && appidGetter(ctx) != "wx2d13afe6dfb82892"
	}, gin.Logger()))

	engine.Use(gin.Recovery())

	engine.Use(cors.New(cors.Config{
		// Set cors and db middleware
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	noAuth := engine.Group(v1Path)
	adminAuth := engine.Group(v1Path).Use(handlers.Admin.Auth)
	{
		noAuth.GET("/proxy", handlers.Proxy.ProxyImage)
	}
	{
		adminAuth.GET("/info", handlers.Admin.Info)
		noAuth.POST("/login", handlers.Admin.Login())
		adminAuth.GET("/logout", handlers.Admin.Logout)
		adminAuth.PUT("/reset", handlers.Admin.Reset())
	}

	{
		noAuth.POST("/listen-verify-ticket", handlers.OpenPlatform.ListenVerifyTicket)
		noAuth.POST(listenMsgRoute, handlers.OpenPlatform.ListenMessage(appidGetter))
		noAuth.GET("/app-authorizer-url", handlers.OpenPlatform.ComponentLoginPage(v1Path+"/listen-authorizer-code"))
		noAuth.GET("/listen-authorizer-code", handlers.OpenPlatform.ListenLoginPage)
		adminAuth.GET("/reset-apps", handlers.OpenPlatform.MigrateApps)
		adminAuth.GET("/count-apps", handlers.OpenPlatform.CountApps)
		adminAuth.GET("/apps", handlers.OpenPlatform.GetApps())
		adminAuth.DELETE("/apps", handlers.OpenPlatform.DelApps())
		adminAuth.GET("/user-statistics", handlers.OpenPlatform.GetUserStatistics())
	}

	{
		adminAuth.GET("/big-picture/count", handlers.OpenPlatform.CountImages(handler.MsgPictureBig))
		adminAuth.GET("/big-pictures", handlers.OpenPlatform.GetImages(handler.MsgPictureBig))
		adminAuth.PUT("/big-pictures", handlers.OpenPlatform.UploadImages(handler.MsgPictureBig))
		adminAuth.DELETE("/big-pictures", handlers.OpenPlatform.DelImages(handler.MsgPictureBig))
	}

	{
		noAuth.GET("/picture/:filename", handlers.OpenPlatform.ServeTempImage(
			func(host, file string) (url string, err error) {
				return v1Path + "/picture/" + file, nil
			},
			func(ctx *gin.Context) (file string) {
				return ctx.Param("filename")
			},
		))
	}

	{
		adminAuth.GET("/mini-program-card", handlers.OpenPlatform.GetMiniProgramCard())
		adminAuth.PUT("/mini-program-card", handlers.OpenPlatform.SaveMiniProgramCard())
		adminAuth.DELETE("/mini-program-card", handlers.OpenPlatform.DelMiniProgramCard())
	}

	{
		adminAuth.GET("/articles", handlers.OpenPlatform.GetArticles())
		adminAuth.PUT("/articles", handlers.OpenPlatform.SaveArticles())
		adminAuth.DELETE("/articles", handlers.OpenPlatform.DelArticles())
	}

	{
		adminAuth.GET("/menus", handlers.OpenPlatform.GetMenus())
		adminAuth.PUT("/menus", handlers.OpenPlatform.SaveMenus())
		adminAuth.POST("/app-menu", handlers.OpenPlatform.GenerateMenu())
		adminAuth.DELETE("/app-menu", handlers.OpenPlatform.RemoveMenu())
	}

	{
		adminAuth.GET("/user-tag", handlers.OpenPlatform.GetUserTag())
		adminAuth.PUT("/user-tag", handlers.OpenPlatform.SaveUserTag())
	}

	{
		adminAuth.GET("/app-groups", handlers.App.GetAppGroups())
		adminAuth.POST("/app-group", handlers.App.CreateAppGroup())
		adminAuth.DELETE("/app-group", handlers.App.DeleteAppGroup())
		adminAuth.GET("/group-apps", handlers.App.GetGroupApps())
		adminAuth.POST("/group-app", handlers.App.SetGroupApp())
		adminAuth.DELETE("/group-app", handlers.App.DeleteGroupApp())

		adminAuth.GET("/apps-payout", handlers.App.GetAppsPayouts())
		adminAuth.POST("/app-payout", handlers.App.SetAppPayout())
	}

	{
		adminAuth.GET("/subscribe/msg/groups", handlers.SubscribeMessage.GetGroups())
		adminAuth.PUT("/subscribe/msg/group", handlers.SubscribeMessage.SaveGroup())
		adminAuth.DELETE("/subscribe/msg/group", handlers.SubscribeMessage.DeleteGroup())

		adminAuth.GET("/subscribe/msg/group/articles", handlers.SubscribeMessage.GetGroupArticles())
		adminAuth.PUT("/subscribe/msg/group/article", handlers.SubscribeMessage.SaveGroupArticle())
		adminAuth.DELETE("/subscribe/msg/group/article", handlers.SubscribeMessage.DeleteGroupArticle())

		adminAuth.GET("/subscribe/msg/group/mini-program-cards", handlers.SubscribeMessage.GetGroupMiniProgramCards())
		adminAuth.PUT("/subscribe/msg/group/mini-program-card", handlers.SubscribeMessage.SaveGroupMiniProgramCard())
		adminAuth.DELETE("/subscribe/msg/group/mini-program-card", handlers.SubscribeMessage.DeleteGroupMiniProgramCard())

		adminAuth.POST("/subscribe/msg/group/apply", handlers.SubscribeMessage.Apply())
		adminAuth.DELETE("/subscribe/msg/group/cancel", handlers.SubscribeMessage.Cancel())
		adminAuth.GET("/subscribe/msg/group/show-applied-msgs", handlers.SubscribeMessage.ShowAppMsgs())
		adminAuth.GET("/subscribe/msg/group/applied-apps", handlers.SubscribeMessage.AppliedApps())
	}

	{
		adminAuth.GET("/check-advert-authorized", handlers.AdvertPlatform.CheckAdvertAuthorize)
		var listenCodeRoute = "/listen-advert-authorizer-code"
		noAuth.GET("/advert-authorizer-url", handlers.AdvertPlatform.GetAdvertAuthorizerUrl(v1Path+listenCodeRoute))
		noAuth.GET(listenCodeRoute, handlers.AdvertPlatform.ListenAdvertAuthorizerCode(v1Path+listenCodeRoute))
		adminAuth.GET("/daily-reports-level-fields", handlers.AdvertPlatform.GetDailyReportsLevelFields)
		adminAuth.POST("/daily-reports", handlers.AdvertPlatform.GetDailyReports)
	}
	serveHttp(*addr, engine)
}

func serveHttp(addr string, engine *gin.Engine) {
	// 开启服务
	srv := &http.Server{
		Addr:         addr,
		Handler:      engine,
		ReadTimeout:  180 * time.Second,
		WriteTimeout: 180 * time.Second,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Emergency.Printf("服务器已停止: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Init.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Init.Println("Server forced to shutdown:", err)
	}
	log.Init.Println("Server exiting")
}

func registerDefaultAdmin(ah *handler.Admin) {
	err := ah.Register("admin", "admin")
	if err != nil && err != admin.ErrUsernameAlreadyExist {
		panic(err)
	}
}
