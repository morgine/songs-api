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

func (tm *AppUserTagClient) Get(appid string) (*users.Tag, error) {
	data, err := tm.engine.Get(tm.key(appid))
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

func (tm *AppUserTagClient) key(appid string) string {
	return tm.keyPrefix + appid
}

func (tm *AppUserTagClient) Set(appid string, m *users.Tag) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return tm.engine.Set(tm.key(appid), data, 0)
}
