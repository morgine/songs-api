package main

import (
	"encoding/json"
	"github.com/morgine/wechat_sdk/pkg/statistics"
	"github.com/morgine/wechat_sdk/src"
	"io/ioutil"
	"os"
)

func main() {
	s := src.MultipleAppUserStatistics{
		AppCount:     0,
		CumulateUser: 0,
		NewUser:      0,
		CancelUser:   0,
		AppUserStatistics: []*src.AppUserStatistics{
			{
				Statistics: &src.UserStatistics{
					CumulateUser: 0,
					NewUser:      0,
					CancelUser:   0,
					Cumulates: []*src.Cumulate{
						{
							Summaries: []*statistics.Summary{
								{},
							},
						},
					},
				},
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
