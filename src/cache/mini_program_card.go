package cache

import (
	"encoding/json"
)

type MiniProgramCard struct {
	Title              string
	PagePath           string
	ThumbMediaFilename string
}

type MiniProgramCardClient struct {
	keyPrefix string
	engine    Engine
}

func NewMiniProgramCardClient(keyPrefix string, engine Engine) *MiniProgramCardClient {
	return &MiniProgramCardClient{
		keyPrefix: keyPrefix,
		engine:    engine,
	}
}

func (tm *MiniProgramCardClient) Get(componentAppid string) (*MiniProgramCard, error) {
	data, err := tm.engine.Get(tm.key(componentAppid))
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		tm := &MiniProgramCard{}
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

func (tm *MiniProgramCardClient) key(appid string) string {
	return tm.keyPrefix + appid
}

func (tm *MiniProgramCardClient) Set(componentAppid string, m *MiniProgramCard) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return tm.engine.Set(tm.key(componentAppid), data, 0)
}

func (tm *MiniProgramCardClient) Del(componentAppid string) error {
	return tm.engine.Del(tm.key(componentAppid))
}
