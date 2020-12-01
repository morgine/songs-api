package handler
//
//import (
//	"bytes"
//	"fmt"
//	"github.com/gin-gonic/gin"
//	"github.com/morgine/songs/pkg/crypt"
//	"github.com/morgine/songs/pkg/xtime"
//	"github.com/morgine/songs/src/cache"
//	"github.com/morgine/songs/src/message"
//	"github.com/morgine/songs/src/model"
//	"strconv"
//)
//
//type Admin struct {
//	Session     *cache.AdminSession
//	Gorm        *model.AdminGorm
//	AuthExpires int64  // 会话过期世界
//	AesCryptKey []byte // 16 位字符串
//}
//
//func (a *Admin) Auth(ctx *gin.Context) {
//	token := ctx.Request.Header.Get("Authorization")
//	if token == "" {
//		SendError(ctx, message.ErrAdminUnauthorized)
//	} else {
//		adminID, err := a.decryptToken(token)
//		if err != nil {
//			SendError(ctx, err)
//		} else {
//			ok, err := a.Session.CheckAndRefreshToken(adminID, token, a.AuthExpires)
//			if err != nil {
//				SendError(ctx, err)
//			} else {
//				if !ok {
//					SendError(ctx, message.ErrAdminUnauthorized)
//				} else {
//					ctx.Set("auth_admin", adminID)
//				}
//			}
//		}
//	}
//}
//
//func GetAuthAdmin(ctx *gin.Context) (adminID int, ok bool) {
//	v, ok := ctx.Get("auth_admin")
//	if ok {
//		return v.(int), true
//	} else {
//		return 0, false
//	}
//}
//
//func (a *Admin) Login() gin.HandlerFunc {
//	type params struct {
//		Username string
//		Password string
//	}
//	return func(ctx *gin.Context) {
//		ps := &params{}
//		err := ctx.Bind(ps)
//		if err != nil {
//			SendError(ctx, err)
//		} else {
//			admin, err := a.Gorm.LoginAdmin(ps.Username, ps.Password)
//			if err != nil {
//				SendError(ctx, err)
//			} else {
//				token, err := a.encryptToken(admin.ID)
//				if err != nil {
//					SendError(ctx, err)
//				} else {
//					err = a.Session.SaveAuthToken(admin.ID, token, a.AuthExpires)
//					if err != nil {
//						SendError(ctx, err)
//					} else {
//						SendJSON(ctx, token)
//					}
//				}
//			}
//		}
//	}
//}
//
//func (a *Admin) ResetPassword() gin.HandlerFunc {
//
//	type params struct {
//		NewPassword string
//	}
//	return func(ctx *gin.Context) {
//		adminID, ok := GetAuthAdmin(ctx)
//		if !ok {
//			SendError(ctx, message.ErrAdminUnauthorized)
//			return
//		}
//		ps := &params{}
//		err := ctx.Bind(ps)
//		if err != nil {
//			SendError(ctx, err)
//		} else {
//			err = a.Gorm.ResetPassword(adminID, ps.NewPassword)
//			if err != nil {
//				SendError(ctx, err)
//			} else {
//				token, err := a.encryptToken(adminID)
//				if err != nil {
//					SendError(ctx, err)
//				} else {
//					err = a.Session.SaveAuthToken(adminID, token, a.AuthExpires)
//					if err != nil {
//						SendError(ctx, err)
//					} else {
//						SendJSON(ctx, token)
//					}
//				}
//			}
//		}
//	}
//}
//
//func (a *Admin) LogoutAdmin(ctx *gin.Context) {
//	admin, ok := GetAuthAdmin(ctx)
//	if !ok {
//		SendMessage(ctx, message.StatusOK, "已退出")
//	} else {
//		err := a.Session.DelAuthToken(admin)
//		if err != nil {
//			SendError(ctx, err)
//		} else {
//			SendMessage(ctx, message.StatusOK, "已退出")
//		}
//	}
//}
//
//// token 加密
//func (a *Admin) encryptToken(adminID int) (token string, err error) {
//	return crypt.AesCBCEncrypt([]byte(fmt.Sprintf("%d:%10d", adminID, xtime.Now().UnixNano())), a.AesCryptKey)
//}
//
//// token 解密
//func (a *Admin) decryptToken(token string) (adminID int, err error) {
//	data, err := crypt.AesCBCDecrypt(token, a.AesCryptKey)
//	if err != nil {
//		return 0, err
//	} else {
//		sepIdx := bytes.Index(data, []byte(":"))
//		return strconv.Atoi(string(data[:sepIdx]))
//	}
//}
