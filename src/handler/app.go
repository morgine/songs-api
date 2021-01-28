package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/morgine/songs/src/message"
	"github.com/morgine/songs/src/model"
	"gorm.io/gorm"
)

type App struct {
	db        *model.AppGroupDB
	payDB     *model.AppPayOutDB
	managerDB *model.AppManagerDB
}

func NewApp(db *gorm.DB) *App {
	agdb, err := model.NewAppGroup(db)
	if err != nil {
		panic(err)
	}
	paydb := model.NewAppPayOutDB(db)
	managerDB := model.NewAppManagerDB(db)
	return &App{db: agdb, payDB: paydb, managerDB: managerDB}
}

func (a *App) GetAppGroups() gin.HandlerFunc {
	type group struct {
		model.AppGroup
		Subs []group
	}
	return func(ctx *gin.Context) {
		gs, err := a.db.GetAppGroups()
		if err != nil {
			SendError(ctx, err)
		} else {
			SendJSON(ctx, gs)
		}
	}
}

func (a *App) CreateAppGroup() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ps := &model.AppGroup{}
		err := ctx.Bind(ps)
		if err != nil {
			SendError(ctx, err)
		} else {
			err := a.db.CreateAppGroup(ps.ParentID, ps)
			if err != nil {
				SendError(ctx, err)
			} else {
				SendJSON(ctx, ps)
			}
		}
	}
}

func (a *App) DeleteAppGroup() gin.HandlerFunc {
	type params struct {
		ID int
	}
	return func(ctx *gin.Context) {
		ps := &params{}
		err := ctx.Bind(ps)
		if err != nil {
			SendError(ctx, err)
		} else {
			err = a.db.DeleteAppGroup(ps.ID)
			if err != nil {
				SendError(ctx, err)
			} else {
				SendMessage(ctx, message.StatusOK, "删除成功")
			}
		}
	}
}

func (a *App) GetGroupApps() gin.HandlerFunc {
	type params struct {
		GroupID int
		Selects []string
	}
	return func(ctx *gin.Context) {
		ps := &params{}
		err := ctx.Bind(ps)
		if err != nil {
			SendError(ctx, err)
		} else {
			apps, err := a.db.GetGroupApps(ps.GroupID, ps.Selects)
			if err != nil {
				SendError(ctx, err)
			} else {
				SendJSON(ctx, apps)
			}
		}
	}
}

func (a *App) SetGroupApp() gin.HandlerFunc {
	type params struct {
		GroupID int
		Appid   string
	}
	return func(ctx *gin.Context) {
		ps := &params{}
		err := ctx.Bind(ps)
		if err != nil {
			SendError(ctx, err)
		} else {
			err := a.db.SetGroupApp(ps.GroupID, ps.Appid)
			if err != nil {
				SendError(ctx, err)
			} else {
				SendMessage(ctx, message.StatusOK, "设置成功")
			}
		}
	}
}

func (a *App) DeleteGroupApp() gin.HandlerFunc {
	type params struct {
		GroupID int
		Appid   string
	}
	return func(ctx *gin.Context) {
		ps := &params{}
		err := ctx.Bind(ps)
		if err != nil {
			SendError(ctx, err)
		} else {
			err = a.db.DeleteGroupApp(ps.GroupID, ps.Appid)
			if err != nil {
				SendError(ctx, err)
			} else {
				SendMessage(ctx, message.StatusOK, "删除成功")
			}
		}
	}
}

func (a *App) SetAppPayout() gin.HandlerFunc {
	type params struct {
		Appid  string
		Payout float64
	}
	return func(ctx *gin.Context) {
		ps := &params{}
		err := ctx.Bind(ps)
		if err != nil {
			SendError(ctx, err)
		} else {
			err = a.payDB.SetAppPayOut(ps.Appid, ps.Payout)
			if err != nil {
				SendError(ctx, err)
			} else {
				SendMessage(ctx, message.StatusOK, "已保存")
			}
		}
	}
}

func (a *App) GetAppsPayouts() gin.HandlerFunc {
	type params struct {
		Appids []string
	}
	return func(ctx *gin.Context) {
		ps := &params{}
		err := ctx.Bind(ps)
		if err != nil {
			SendError(ctx, err)
		} else {
			pays := a.payDB.GetAppPayOut(ps.Appids)
			SendJSON(ctx, pays)
		}
	}
}

func (a *App) SetAppManager() gin.HandlerFunc {
	type params struct {
		Appid   string
		Manager string
	}
	return func(ctx *gin.Context) {
		ps := &params{}
		err := ctx.Bind(ps)
		if err != nil {
			SendError(ctx, err)
		} else {
			err = a.managerDB.SetAppManager(ps.Appid, ps.Manager)
			if err != nil {
				SendError(ctx, err)
			} else {
				SendMessage(ctx, message.StatusOK, "已保存")
			}
		}
	}
}

func (a *App) GetAppsManagers() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		managers := a.managerDB.GetAppManagers()
		SendJSON(ctx, managers)
	}
}
