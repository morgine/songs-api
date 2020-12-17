package cache

import (
	"encoding/json"
	"github.com/morgine/wechat_sdk/pkg/material"
)

type MaterialClient struct {
	keyPrefix string
	engine    Engine
}

func NewMaterialClient(keyPrefix string, engine Engine) *MaterialClient {
	return &MaterialClient{
		keyPrefix: keyPrefix,
		engine:    engine,
	}
}

func (tm *MaterialClient) Get(appid, filename string) (*material.UploadedMedia, error) {
	data, err := tm.engine.Get(tm.key(appid, filename))
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		um := &material.UploadedMedia{}
		err = json.Unmarshal(data, um)
		if err != nil {
			return nil, err
		} else {
			return um, nil
		}
	} else {
		return nil, nil
	}
}

func (tm *MaterialClient) key(appid, filename string) string {
	return tm.keyPrefix + appid + "_" + filename
}

func (tm *MaterialClient) Set(appid, filename string, m *material.UploadedMedia) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return tm.engine.Set(tm.key(appid, filename), data, 0)
}
