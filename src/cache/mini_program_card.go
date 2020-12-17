package cache

import (
	"encoding/json"
)

type MiniProgramCard struct {
	Title              string
	Appid              string
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

func (tm *MiniProgramCardClient) Get(componentAppid string) ([]*MiniProgramCard, error) {
	data, err := tm.engine.Get(tm.key(componentAppid))
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		var cards []*MiniProgramCard
		err = json.Unmarshal(data, &cards)
		if err != nil {
			return nil, err
		} else {
			return cards, nil
		}
	} else {
		return nil, nil
	}
}

func (tm *MiniProgramCardClient) key(appid string) string {
	return tm.keyPrefix + appid
}

func (tm *MiniProgramCardClient) Set(componentAppid string, cards []*MiniProgramCard) error {
	data, err := json.Marshal(cards)
	if err != nil {
		return err
	}
	return tm.engine.Set(tm.key(componentAppid), data, 0)
}

func (tm *MiniProgramCardClient) Del(componentAppid string) error {
	return tm.engine.Del(tm.key(componentAppid))
}
