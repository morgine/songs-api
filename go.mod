module github.com/morgine/songs

go 1.15

require (
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-gonic/gin v1.6.3
	github.com/go-redis/redis/v8 v8.4.0
	github.com/google/go-querystring v1.0.0
	github.com/jackc/pgx/v4 v4.9.2 // indirect
	github.com/morgine/log v0.0.0-20200723085359-3eb4c2be1006
	github.com/morgine/pkg v0.0.0-20201210141122-1eaea814a846
	github.com/morgine/wechat_sdk v0.0.0-20201210141225-4198e3ac4f37
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/orivil/wechat v0.0.0-20200318075742-9daadcc7fab3
	golang.org/x/text v0.3.4 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gorm.io/gorm v1.20.7
)

replace github.com/morgine/wechat_sdk => ../wechat_sdk

replace github.com/morgine/pkg => ../morgine/pkg
