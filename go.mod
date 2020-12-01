module github.com/morgine/songs

go 1.15

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-gonic/gin v1.6.3
	github.com/go-redis/redis/v8 v8.4.0
	github.com/google/go-querystring v1.0.0
	github.com/morgine/cfg v0.0.0-20200804152015-cd175a04c4d8
	github.com/morgine/database v0.0.0-20201112022355-001078b66b4d
	github.com/morgine/log v0.0.0-20200723085359-3eb4c2be1006
	github.com/morgine/pkg v0.0.0-20201128121612-22825333d18f
	github.com/morgine/redis v0.0.0-20201112085733-0090621c3b52 // indirect
	github.com/morgine/service v0.0.0-20200716030345-bd68903c522c
	github.com/tencentad/marketing-api-go-sdk v1.4.0 // indirect
	golang.org/x/crypto v0.0.0-20201117144127-c1f2f97bffc9
	gorm.io/gorm v1.20.7
)

replace github.com/morgine/pkg v0.0.0-20201128121612-22825333d18f => ../pkg
