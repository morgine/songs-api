package cache

import (
	"encoding/json"
)

type UserTag struct {
	Name string
}

type UserTagClient struct {
	keyPrefix string
	engine    Engine
}

func NewUserTagClient(keyPrefix string, engine Engine) *UserTagClient {
	return &UserTagClient{
		keyPrefix: keyPrefix,
		engine:    engine,
	}
}

func (tm *UserTagClient) Get(componentAppid string) (*UserTag, error) {
	data, err := tm.engine.Get(tm.key(componentAppid))
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		tm := &UserTag{}
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

func (tm *UserTagClient) key(appid string) string {
	return tm.keyPrefix + appid
}

func (tm *UserTagClient) Set(componentAppid string, m *UserTag) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return tm.engine.Set(tm.key(componentAppid), data, 0)
}
