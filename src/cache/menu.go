package cache

import (
	"encoding/json"
	"github.com/morgine/wechat_sdk/pkg/custom_menu"
)

type Menus []Menu

type Menu struct {
	Name    string               `json:"name"`
	Buttons []custom_menu.Button `json:"buttons"`
}

type MenusClient struct {
	keyPrefix string
	engine    Engine
}

func NewMenusClient(keyPrefix string, engine Engine) *MenusClient {
	return &MenusClient{
		keyPrefix: keyPrefix,
		engine:    engine,
	}
}

func (mc *MenusClient) Get(componentAppid string) (Menus, error) {
	data, err := mc.engine.Get(mc.key(componentAppid))
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		var menus Menus
		err = json.Unmarshal(data, &menus)
		if err != nil {
			return nil, err
		} else {
			return menus, nil
		}
	} else {
		return nil, nil
	}
}

func (mc *MenusClient) key(appid string) string {
	return mc.keyPrefix + appid
}

func (mc *MenusClient) Set(componentAppid string, menus Menus) error {
	data, err := json.Marshal(menus)
	if err != nil {
		return err
	}
	return mc.engine.Set(mc.key(componentAppid), data, 0)
}

func (mc *MenusClient) Del(componentAppid string) error {
	return mc.engine.Del(mc.key(componentAppid))
}
