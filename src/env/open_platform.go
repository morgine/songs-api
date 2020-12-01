package env

import (
	"github.com/morgine/pkg/config"
	"github.com/morgine/songs/src/model"
	"github.com/morgine/songs/src/platform"
	"github.com/morgine/songs/src/wpt"
)

func NewOpenPlatform(namespace string, accessStore platform.AccessTokenStorage, m *model.Gorm, configs config.Configs) (*platform.OpenPlatform, error) {
	env := &OpenPlatformEnv{}
	err := configs.UnmarshalSub(namespace, env)
	if err != nil {
		return nil, err
	}
	msgCrypt, err := wpt.NewWXBizMsgCrypt(env.AesToken, env.EncodingAesKey, env.Appid)
	if err != nil {
		return nil, err
	}
	appStore := newOpenPlatformStorage(accessStore, m.App())
	return platform.NewOpenPlatform(env.Appid, env.AppSecret, appStore, msgCrypt), nil
}