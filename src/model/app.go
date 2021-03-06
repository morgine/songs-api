package model

import "gorm.io/gorm"

type App struct {
	ID int
	// 公众号 appID
	Appid string `gorm:"index"`
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

type AppModel struct {
	g *Model
}

func (m *Model) App() *AppModel {
	return &AppModel{g: m}
}

func (g *AppModel) SaveAPP(appid string, app *App) error {
	app.ID = 0
	exist := &App{}
	err := g.g.First(exist, Where("appid=?", appid))
	if err != nil {
		return err
	}
	if exist.ID > 0 {
		return g.g.Updates(app, Where("appid=?", appid))
	} else {
		return g.g.Create(app)
	}
}

func (g *AppModel) CountApps(c ...Condition) (total int64, err error) {
	return g.g.Count(&App{}, c...)
}

func (g *AppModel) GetAppsByAppids(appids []string) (apps []*App, err error) {
	return g.GetApps(ConditionFunc(func(db *gorm.DB) *gorm.DB {
		return db.Where("appid in (?)", appids)
	}))
}

func (g *AppModel) GetApps(c ...Condition) (apps []*App, err error) {
	err = g.g.Find(&apps, c...)
	return
}

// 重置所有 APP, 参数 apps 必须一次性包含所有已授权的 APP 信息，该操作会删除多余的 APP，并尝试更新或创建新的 APP 信息
func (g *AppModel) ResetApps(apps []*App) (err error) {
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

func (g *AppModel) DelAppByAppid(appid string) (err error) {
	return g.DelApps(Where("appid=?", appid))
}

type InIDs []int

func (i InIDs) AppendCondition(db *gorm.DB) *gorm.DB {
	return db.Where("id in (?)", i)
}

func (g *AppModel) DelApps(c ...Condition) (err error) {
	return g.g.Delete(&App{}, c...)
}
