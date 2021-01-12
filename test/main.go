package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

func main() {
	s := &Button{
		SubButton: []Button{},
	}
	data, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("data.json", data, os.ModePerm)
	if err != nil {
		panic(err)
	}
}

type Button struct {
	Name      string   `json:"name"`       // 菜单标题，不超过16个字节，子菜单不超过60个字节
	Type      string   `json:"type"`       // 菜单的响应动作类型，view表示网页类型，click表示点击类型，miniprogram表示小程序类型
	SubButton []Button `json:"sub_button"` // 子菜单, 1-5 个
	Url       string   `json:"url"`        // view, miniprogram 类型必须, 网页 链接，用户点击菜单可打开链接，不超过1024字节。 type为miniprogram时，不支持小程序的老版本客户端将打开本url。
	Key       string   `json:"key"`        // click, pic_sysphoto, pic_photo_or_album, pic_weixin, location_select 类型必须, 菜单KEY值，用于消息接口推送，不超过128字节
	Appid     string   `json:"appid"`      // miniprogram 类型必须, 小程序的appid（仅认证公众号可配置）
	Pagepath  string   `json:"pagepath"`   // miniprogram 类型必须, 小程序的页面路径
	MediaId   string   `json:"media_id"`   // media_id, view_limited 类型必须, 调用新增永久素材接口返回的合法media_id
}
