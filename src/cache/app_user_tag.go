package cache

import (
	"encoding/json"
	"github.com/morgine/wechat_sdk/pkg/users"
)

type AppUserTagClient struct {
	keyPrefix string
	engine    Engine
}

func NewAppUserTagClient(keyPrefix string, engine Engine) *AppUserTagClient {
	return &AppUserTagClient{
		keyPrefix: keyPrefix,
		engine:    engine,
	}
}

func (tm *AppUserTagClient) Get(appid, tagName string) (*users.Tag, error) {
	data, err := tm.engine.Get(tm.key(appid, tagName))
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		tm := &users.Tag{}
		err = json.Unmarshal(data, tm)
		if err != nil {
			return nil, err
		} else {
			return tm, nil
		}
	} else {
		return nil, nil
	}
}

func (tm *AppUserTagClient) key(appid, tagName string) string {
	return tm.keyPrefix + appid + "_" + tagName
}

func (tm *AppUserTagClient) Set(appid, tagName string, m *users.Tag) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return tm.engine.Set(tm.key(appid, tagName), data, 0)
}
