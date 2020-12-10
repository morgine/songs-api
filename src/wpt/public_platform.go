package wpt

type UserSummary struct {
	RefDate    string `json:"ref_date"`    // 数据的日期
	UserSource int    `json:"user_source"` // 用户的渠道，数值代表的含义如下： 0代表其他合计 1代表公众号搜索 17代表名片分享 30代表扫描二维码 51代表支付后关注（在支付完成页） 57代表文章内账号名称 100微信广告 161他人转载 176 专辑页内账号名称
	NewUser    int    `json:"new_user"`    // 新增的用户数量
	CancelUser int    `json:"cancel_user"` // 取消关注的用户数量，new_user减去cancel_user即为净增用户数量
	//CumulateUser int    `json:"cumulate_user"` // 总用户量
}

// 获取用户增减数据
// access_token	是	调用接口凭证
// begin_date	是	获取数据的起始日期，begin_date和end_date的差值需小于“最大时间跨度”（比如最大时间跨度为1时，begin_date和end_date的差值只能为0，才能小于1），否则会报错
// end_date	是	获取数据的结束日期，end_date允许设置的最大值为昨日
func GetUserSummary(accessToken string, beginDate, endDate string) (summaries []*UserSummary, err error) {
	var params = &struct {
		BeginDate string `json:"begin_date"`
		EndDate   string `json:"end_date"`
	}{
		BeginDate: beginDate,
		EndDate:   endDate,
	}
	res := &struct {
		List []*UserSummary `json:"list"`
	}{}
	url := "https://api.weixin.qq.com/datacube/getusersummary?access_token=" + accessToken
	err = PostSchema(KindJson, url, params, res)
	if err != nil {
		return nil, err
	} else {
		return res.List, nil
	}
}

type UserCumulate struct {
	RefDate      string // 数据的日期
	CumulateUser int    // 总用户量
}

// 获取累计用户数据
func GetUserCumulate(accessToken string, beginDate, endDate string) (summaries []*UserCumulate, err error) {
	data := map[string]string{
		"begin_date": beginDate,
		"end_date":   endDate,
	}
	res := &struct {
		List []*UserCumulate `json:"list"`
	}{}
	url := "https://api.weixin.qq.com/datacube/getusercumulate?access_token=" + accessToken
	err = PostSchema(KindJson, url, data, res)
	if err != nil {
		return nil, err
	} else {
		return res.List, nil
	}
}

type UserTag struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

func CreateAppUserTag(accessToken, name string) (*UserTag, error) {
	data := map[string]interface{}{
		"tag": &UserTag{Name: name},
	}
	res := &struct {
		Tag *UserTag `json:"tag"`
	}{}
	url := "https://api.weixin.qq.com/cgi-bin/tags/create?access_token=" + accessToken
	err := PostSchema(KindJson, url, data, res)
	if err != nil {
		return nil, err
	} else {
		return res.Tag, nil
	}
}

func GetAppUserTags(accessToken string) ([]*UserTag, error) {
	res := &struct {
		Tags []*UserTag `json:"tags"`
	}{}
	url := "https://api.weixin.qq.com/cgi-bin/tags/get?access_token=" + accessToken
	err := PostSchema(KindJson, url, nil, res)
	if err != nil {
		return nil, err
	} else {
		return res.Tags, nil
	}
}

func UpdateAppUserTag(accessToken string, tag *UserTag) error {
	data := &struct {
		Tag *UserTag `json:"tag"`
	}{}
	url := "https://api.weixin.qq.com/cgi-bin/tags/update?access_token=" + accessToken
	return PostSchema(KindJson, url, data, nil)
}

func DeleteAppUserTag(accessToken string, tagID int) error {
	data := &struct {
		Tag *UserTag `json:"tag"`
	}{
		Tag: &UserTag{ID: tagID},
	}
	url := "https://api.weixin.qq.com/cgi-bin/tags/delete?access_token=" + accessToken
	return PostSchema(KindJson, url, data, nil)
}

type TagUsers struct {
	Count int `json:"count"`
	Data  struct {
		Openid []string `json:"openid"`
	} `json:"data"`
	NextOpenid string `json:"next_openid"`
}

// 获得对应标签下的用户列表
func GetAppTagUsers(accessToken string, tagID int, nextOpenid string) (*TagUsers, error) {
	data := map[string]interface{}{
		"tagid":       tagID,
		"next_openid": nextOpenid,
	}
	res := &TagUsers{}
	url := "https://api.weixin.qq.com/cgi-bin/user/tag/get?access_token=" + accessToken
	err := PostSchema(KindJson, url, data, res)
	if err != nil {
		return nil, err
	} else {
		return res, nil
	}
}

// 批量为用户打标签
func BatchTagging(accessToken string, tagID int, openids []string) error {
	data := map[string]interface{}{
		"tagid":       tagID,
		"openid_list": openids,
	}
	url := "https://api.weixin.qq.com/cgi-bin/tags/members/batchtagging?access_token=" + accessToken
	return PostSchema(KindJson, url, data, nil)
}

// 批量为用户取消标签
func BatchUntagging(accessToken string, tagID int, openids []string) error {
	data := map[string]interface{}{
		"tagid":       tagID,
		"openid_list": openids,
	}
	url := "https://api.weixin.qq.com/cgi-bin/tags/members/batchuntagging?access_token=" + accessToken
	return PostSchema(KindJson, url, data, nil)
}

// 获取用户身上的标签列表
func GetUserTags(accessToken string, openid string) (ids []int, err error) {
	data := map[string]interface{}{
		"openid": openid,
	}
	res := &struct {
		TagidList []int `json:"tagid_list"`
	}{}
	url := "https://api.weixin.qq.com/cgi-bin/tags/getidlist?access_token=" + accessToken
	err = PostSchema(KindJson, url, data, res)
	if err != nil {
		return nil, err
	} else {
		return res.TagidList, nil
	}
}
