package model

import "gorm.io/gorm"

type App struct {
	ID int
	// 公众号 appID
	Appid string `gorm:"uniqueIndex"`
	// 授权方昵称
	NickName string
	// 授权方头像
	HeadImg string
	// 授权方公众号的原始ID
	UserName string
	// 公众号的主体名称
	PrincipalName string
	// 授权方公众号所设置的微信号，可能为空
	Alias string
	// 二维码图片的URL，开发者最好自行也进行保存
	QrcodeUrl string
	// APP 简介
	Signature string
}

type AppGorm struct {
	g *Gorm
}

func (g *Gorm) App() *AppGorm {
	return &AppGorm{g: g}
}

func (g *AppGorm) SaveAPP(appid string, app *App) error {
	app.ID = 0
	return g.g.Save(app, Where("appid=?", appid))
}

func (g *AppGorm) CountApps(c ...Condition) (total int64, err error) {
	return g.g.Count(&App{}, c...)
}

func (g *AppGorm) GetAppsByAppids(appids []string) (apps []*App, err error) {
	return g.GetApps(ConditionFunc(func(db *gorm.DB) *gorm.DB {
		return db.Where("appid in (?)", appids)
	}))
}

func (g *AppGorm) GetApps(c ...Condition) (apps []*App, err error) {
	err = g.g.Find(&apps, c...)
	return
}

func (g *AppGorm) ResetApps(apps []*App) (err error) {
	var appids []string
	for _, app := range apps {
		appids = append(appids, app.Appid)
	}
	err = g.DelApps(ConditionFunc(func(db *gorm.DB) *gorm.DB {
		return db.Not("appid in (?)", appids)
	}))
	if err != nil {
		return err
	}
	for _, app := range apps {
		err = g.SaveAPP(app.Appid, app)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *AppGorm) DelAppByAppid(appid string) (err error) {
	return g.DelApps(Where("appid=?", appid))
}

func (g *AppGorm) DelApps(c ...Condition) (err error) {
	return g.g.Delete(&App{}, c...)
}
