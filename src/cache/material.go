package cache

import (
	"encoding/json"
	"github.com/morgine/wechat_sdk/pkg/material"
	"time"
)

type TempMaterialClient struct {
	keyPrefix string
	engine    Engine
}

func NewTempMaterialClient(keyPrefix string, engine Engine) *TempMaterialClient {
	return &TempMaterialClient{
		keyPrefix: keyPrefix,
		engine:    engine,
	}
}

func (tm *TempMaterialClient) Get(appid, filename string) (*material.TempMedia, error) {
	data, err := tm.engine.Get(tm.key(appid, filename))
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		tm := &material.TempMedia{}
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

func (tm *TempMaterialClient) key(appid, filename string) string {
	return tm.keyPrefix + appid + "_" + filename
}

func (tm *TempMaterialClient) Set(appid, filename string, m *material.TempMedia) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return tm.engine.Set(tm.key(appid, filename), data, time.Duration(m.ExpireIn)*time.Second)
}
