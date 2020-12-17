package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/morgine/pkg/admin"
	"github.com/morgine/songs/src/message"
)

type Admin struct {
	h *admin.Handler
}

func NewAdmin(opts *admin.Options) (*Admin, error) {
	h, err := admin.NewHandler(opts)
	if err != nil {
		return nil, err
	}
	return &Admin{h: h}, nil
}

func (a *Admin) getTokenFromHeader(ctx *gin.Context) string {
	return ctx.Request.Header.Get("Authorization")
}

func (a *Admin) Auth(ctx *gin.Context) {
	token := a.getTokenFromHeader(ctx)
	adminID, err := a.h.CheckAndRefreshToken(token)
	if err != nil {
		SendError(ctx, err)
	} else {
		ctx.Set("admin_id", adminID)
	}
}

func (a *Admin) GetLoginAdminID(ctx *gin.Context) (adminID int, ok bool) {
	id, ok := ctx.Get("admin_id")
	if ok {
		return id.(int), true
	}
	return 0, false
}

func (a *Admin) Info(ctx *gin.Context) {
	adminID, ok := a.GetLoginAdminID(ctx)
	if ok {
		info, err := a.h.GetAdmin(adminID)
		if err != nil {
			SendError(ctx, err)
		} else {
			SendJSON(ctx, info)
		}
	} else {
		SendError(ctx, message.ErrAdminUnauthorized)
	}
}

func (a *Admin) Login() gin.HandlerFunc {
	type params struct {
		Username string
		Password string
	}
	return func(ctx *gin.Context) {
		ps := &params{}
		err := ctx.Bind(ps)
		if err != nil {
			SendError(ctx, err)
		} else {
			token, err := a.h.Login(ps.Username, ps.Password)
			if err != nil {
				SendError(ctx, err)
			} else {
				SendJSON(ctx, token)
			}
		}
	}
}

func (a *Admin) Logout(ctx *gin.Context) {
	adminID, ok := a.GetLoginAdminID(ctx)
	if ok {
		token := a.getTokenFromHeader(ctx)
		err := a.h.Logout(adminID, token)
		if err != nil {
			SendError(ctx, err)
		} else {
			SendMessage(ctx, message.StatusOK, "已退出")
		}
	}
}

func (a *Admin) Reset() gin.HandlerFunc {
	type params struct {
		NewPassword string
	}
	return func(ctx *gin.Context) {
		adminID, ok := a.GetLoginAdminID(ctx)
		if ok {
			ps := &params{}
			err := ctx.Bind(ps)
			if err != nil {
				SendError(ctx, err)
			} else {
				err = a.h.ResetPassword(adminID, ps.NewPassword)
				if err != nil {
					SendError(ctx, err)
				} else {
					SendMessage(ctx, message.StatusOK, "已保存")
				}
			}
		}
	}
}

// 注册账号，如果账号已存在则返回 ErrUsernameAlreadyExist 错误
func (a *Admin) Register(username, password string) error {
	return a.h.RegisterAdmin(username, password)
}
