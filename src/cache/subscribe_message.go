package cache

import (
	"encoding/json"
	"github.com/morgine/wechat_sdk/pkg/message"
	"github.com/morgine/wechat_sdk/src"
)

type SubscribeMessages struct {
	Articles []src.Article
	Cards    []*message.MiniProgramPage
}

type SubscribeMessageClient struct {
	keyPrefix string
	engine    Engine
}

func NewSubscribeMessageClient(keyPrefix string, engine Engine) *SubscribeMessageClient {
	return &SubscribeMessageClient{
		keyPrefix: keyPrefix,
		engine:    engine,
	}
}

func (tm *SubscribeMessageClient) Get(appid string) (msgs *SubscribeMessages, err error) {
	data, err := tm.engine.Get(tm.key(appid))
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		msgs = &SubscribeMessages{}
		err = json.Unmarshal(data, &msgs)
		if err != nil {
			return nil, err
		} else {
			return msgs, nil
		}
	} else {
		return nil, nil
	}
}

func (tm *SubscribeMessageClient) key(appid string) string {
	return tm.keyPrefix + appid
}

func (tm *SubscribeMessageClient) Set(appid string, msgs *SubscribeMessages) error {
	data, err := json.Marshal(msgs)
	if err != nil {
		return err
	}
	return tm.engine.Set(tm.key(appid), data, 0)
}

func (tm *SubscribeMessageClient) Del(appid string) error {
	return tm.engine.Del(tm.key(appid))
}
