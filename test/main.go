package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

func main() {
	s := DatesStatistics{
		{
			Apps: []*AppStatistics{
				{},
			},
		},
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

// 统计
type Statistics struct {
	CumulateUser      int     `json:"cumulate_user"`       // 用户总量
	NewUser           int     `json:"new_user"`            // 新增的用户数量
	CancelUser        int     `json:"cancel_user"`         // 取消关注的用户数量，new_user减去cancel_user即为净增用户数量
	PositiveUser      int     `json:"positive_user"`       // 净增用户
	CancelRate        float64 `json:"cancel_rate"`         // 取关率, cancel_user/cumulate_user
	ReqSuccCount      int     `json:"req_succ_count"`      // 拉取量
	ExposureCount     int     `json:"exposure_count"`      // 曝光量
	ExposureRate      float64 `json:"exposure_rate"`       // 曝光率
	ClickCount        int     `json:"click_count"`         // 点击量
	ClickRate         float64 `json:"click_rate"`          // 点击率
	Outcome           int     `json:"outcome"`             // 支出(分)
	Income            int     `json:"income"`              // 收入(分)
	IncomeOutcomeRate float64 `json:"income_outcome_rate"` // 收入支出比率
	Ecpm              int     `json:"ecpm"`                // 广告千次曝光收益(分)
}

// 公众号统计
type AppStatistics struct {
	Appid    string     `json:"appid"`    // 公众号 appid
	Nickname string     `json:"nickname"` // 公众号昵称
	Errs     []string   `json:"err"`      // 错误
	Data     Statistics `json:"data"`     // 统计数据
}

// 日期统计数据
type DateStatistics struct {
	Date               string           `json:"date"`                 // 统计日期
	Data               Statistics       `json:"data"`                 // 总体统计
	TotalExposureCount float64          `json:"total_exposure_count"` // 已曝光量加未曝光量，用于计算曝光率
	TotalClickCount    float64          `json:"total_click_count"`    // 已点击量加未点击量，用于计算点击率
	Apps               []*AppStatistics `json:"apps"`                 // APP 统计
}
type DatesStatistics []*DateStatistics // 多日期统计数据
