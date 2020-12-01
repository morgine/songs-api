package src

import (
	"context"
	"flag"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/morgine/log"
	"github.com/morgine/pkg/admin"
	"github.com/morgine/songs/src/env"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run() {
	// 加载配置文件
	configFile := flag.String("c", "config.toml", "配置文件")
	addr := flag.String("addr", ":9879", "监听地址")
	flag.Parse()
	handlers, release := env.LoadEnv(*configFile)
	defer release()
	registerDefaultAdmin(handlers.Admin)
	// 注册路由
	engine := gin.New()

	engine.Use(gin.Logger())

	engine.Use(cors.New(cors.Config{
		// Set cors and db middleware
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	noAuth := engine.Group("/")
	adminAuth := engine.Group("/").Use(handlers.Admin.Auth("Authorization"))

	{
		adminAuth.GET("/admin/info", handlers.Admin.GetLoginAdmin)
		noAuth.POST("/login", handlers.Admin.Login())
		adminAuth.GET("/logout", handlers.Admin.Logout)
		adminAuth.PUT("/reset", handlers.Admin.ResetPassword())
	}

	{
		noAuth.GET("/listen-verify-ticket", handlers.OpenPlatform.ListenVerifyTicket)
		noAuth.GET("/app-authorizer-url", handlers.OpenPlatform.GetAppAuthorizerUrl("/listen-authorizer-code"))
		noAuth.GET("/listen-app-authorizer-code", handlers.OpenPlatform.ListenAppAuthorizerCode)
		adminAuth.GET("/reset-apps", handlers.OpenPlatform.ResetAppAuthorizers)
		adminAuth.GET("/count-apps", handlers.OpenPlatform.CountApps)
		adminAuth.GET("/apps", handlers.OpenPlatform.GetApps())
		adminAuth.GET("/user-summary", handlers.OpenPlatform.GetUserSummary())
		adminAuth.GET("/user-cumulate", handlers.OpenPlatform.GetUserCumulate())
	}

	{
		adminAuth.GET("/check-advert-authorized", handlers.AdvertPlatform.CheckAdvertAuthorize)
		var listenCodeRoute = "/listen-advert-authorizer-code"
		noAuth.GET("/advert-authorizer-url", handlers.AdvertPlatform.GetAdvertAuthorizerUrl(listenCodeRoute))
		noAuth.GET(listenCodeRoute, handlers.AdvertPlatform.ListenAdvertAuthorizerCode(listenCodeRoute))
		adminAuth.GET("/daily-reports-level-fields", handlers.AdvertPlatform.GetDailyReportsLevelFields)
		adminAuth.POST("/daily-reports", handlers.AdvertPlatform.GetDailyReports)
	}

	serveHttp(*addr, engine)
}

func serveHttp(addr string, engine *gin.Engine) {
	// 开启服务
	srv := &http.Server{
		Addr:    addr,
		Handler: engine,
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

func registerDefaultAdmin(handler *admin.Handler) {
	err := handler.RegisterAdmin("admin", "admin")
	if err != nil && err != admin.ErrUsernameAlreadyExist {
		panic(err)
	}
}
