package ad

import (
	"encoding/json"
	"github.com/morgine/log"
	"math/rand"
	"net/url"
	"strconv"
	"time"
)

var Now = time.Now

type DateRange struct {
	// 开始日期，日期格式：YYYY-MM-DD，且小于等于 end_date
	//	字段长度为 10 字节
	StartDate string `json:"start_date"`
	// 结束日期，日期格式：YYYY-MM-DD，且大于等于 begin_date
	//	字段长度为 10 字节
	EndDate string `json:"end_date"`
}

type Filtering struct {
	// 过滤字段，腾讯广告平台广告主：当 level 为 REPORT_LEVEL_CAMPAIGN 时，field 仅可选择 campaign_id; 当 level 为 REPORT_LEVEL_ADGROUP 时，field 可选择 adgroup_id, campaign_id; 当 level 为 REPORT_LEVEL_AD 时，field 仅可选择为 ad_id, adgroup_id, campaign_id ；当 level 为 REPORT_LEVEL_PROMOTED_OBJECT 时，field 可选择 promoted_object_type 和 promoted_object_id;当 level 为 REPORT_LEVEL_UNION_POSITION 时，field 仅可选择为 promoted_object_type , promoted_object_id , union_position_id;当 level 为 REPORT_LEVEL_CREATIVE_TEMPLATE 时，field 仅可选择为 template_id;当 level 为 REPORT_LEVEL_EXPAND_TARGETING_ADGROUP 时，field 仅可选择为 adgroup_id;当 level 为 REPORT_LEVEL_PRODUCT_CATELOG 时，field 必须选择 product_catalog_id
	//	可选值：{ adgroup_id, campaign_id, ad_id, promoted_object_type, promoted_object_id, union_position_id, template_id, bid_type, product_catalog_id, material_id }
	//	微信公众账号逻辑
	//
	//	当 level 为 REPORT_LEVEL_CAMPAIGN_WECHAT 时，field 仅可选择 campaign_id; 当 level 为 REPORT_LEVEL_ADGROUP_WECHAT 时，field 可选择 adgroup_id, campaign_id; 当 level 为 REPORT_LEVEL_AD_WECHAT 时，field 仅可选择为 ad_id, adgroup_id, campaign_id
	Field string `json:"field"`

	// 操作符
	//	当 field 取值 adgroup_id 时，枚举列表：{ EQUALS, IN }
	//	当 field 取值 campaign_id 时，枚举列表：{ EQUALS, IN }
	//	当 field 取值 ad_id 时，枚举列表：{ EQUALS, IN }
	//	当 field 取值 promoted_object_type 时，枚举列表：{ EQUALS, IN }
	//	当 field 取值 promoted_object_id 时，枚举列表：{ EQUALS, IN }
	//	当 field 取值 union_position_id 时，枚举列表：{ EQUALS }
	//	当 field 取值 template_id 时，枚举列表：{ EQUALS, IN }
	//	当 field 取值 bid_type 时，枚举列表：{ EQUALS }
	//	当 field 取值 product_catalog_id 时，枚举列表：{ EQUALS }
	//	当 field 取值 material_id 时，枚举列表：{ EQUALS, IN }
	Operator string `json:"operator"`

	// 字段取值，values 数组元素的个数限制与 operator 的取值相关，当 level 为 REPORT_LEVEL_UNION_POSITION 时，promoted_object_type , promoted_object_id , union_position_id , template_id, bid_type 的 values 数组元素个数限制为 1 个，详见 [过滤条件]
	//
	//	当 field 取值 promoted_object_type 时，
	//	枚举列表：{ PROMOTED_OBJECT_TYPE_APP_ANDROID, PROMOTED_OBJECT_TYPE_APP_IOS, PROMOTED_OBJECT_TYPE_ECOMMERCE, PROMOTED_OBJECT_TYPE_LINK_WECHAT, PROMOTED_OBJECT_TYPE_APP_ANDROID_MYAPP, PROMOTED_OBJECT_TYPE_APP_ANDROID_UNION, PROMOTED_OBJECT_TYPE_LOCAL_ADS_WECHAT, PROMOTED_OBJECT_TYPE_QQ_BROWSER_MINI_PROGRAM, PROMOTED_OBJECT_TYPE_LINK, PROMOTED_OBJECT_TYPE_QQ_MESSAGE, PROMOTED_OBJECT_TYPE_QZONE_VIDEO_PAGE, PROMOTED_OBJECT_TYPE_LOCAL_ADS, PROMOTED_OBJECT_TYPE_ARTICLE, PROMOTED_OBJECT_TYPE_LEAD_AD, PROMOTED_OBJECT_TYPE_TENCENT_KE, PROMOTED_OBJECT_TYPE_EXCHANGE_APP_ANDROID_MYAPP, PROMOTED_OBJECT_TYPE_QZONE_PAGE_ARTICLE, PROMOTED_OBJECT_TYPE_QZONE_PAGE_IFRAMED, PROMOTED_OBJECT_TYPE_QZONE_PAGE, PROMOTED_OBJECT_TYPE_APP_PC, PROMOTED_OBJECT_TYPE_MINI_GAME_WECHAT, PROMOTED_OBJECT_TYPE_MINI_GAME_QQ }
	//
	//	当 field 取值 promoted_object_type 且 operator 取值 IN 时，
	//	枚举列表：{ PROMOTED_OBJECT_TYPE_APP_ANDROID, PROMOTED_OBJECT_TYPE_APP_IOS, PROMOTED_OBJECT_TYPE_ECOMMERCE, PROMOTED_OBJECT_TYPE_LINK_WECHAT, PROMOTED_OBJECT_TYPE_APP_ANDROID_MYAPP, PROMOTED_OBJECT_TYPE_APP_ANDROID_UNION, PROMOTED_OBJECT_TYPE_LOCAL_ADS_WECHAT, PROMOTED_OBJECT_TYPE_QQ_BROWSER_MINI_PROGRAM, PROMOTED_OBJECT_TYPE_LINK, PROMOTED_OBJECT_TYPE_QQ_MESSAGE, PROMOTED_OBJECT_TYPE_QZONE_VIDEO_PAGE, PROMOTED_OBJECT_TYPE_LOCAL_ADS, PROMOTED_OBJECT_TYPE_ARTICLE, PROMOTED_OBJECT_TYPE_LEAD_AD, PROMOTED_OBJECT_TYPE_TENCENT_KE, PROMOTED_OBJECT_TYPE_EXCHANGE_APP_ANDROID_MYAPP, PROMOTED_OBJECT_TYPE_QZONE_PAGE_ARTICLE, PROMOTED_OBJECT_TYPE_QZONE_PAGE_IFRAMED, PROMOTED_OBJECT_TYPE_QZONE_PAGE, PROMOTED_OBJECT_TYPE_APP_PC, PROMOTED_OBJECT_TYPE_MINI_GAME_WECHAT, PROMOTED_OBJECT_TYPE_MINI_GAME_QQ }
	//
	//
	//	当 field 取值 bid_type 时，
	//	枚举列表：{ BID_TYPE_CPC, BID_TYPE_CPM }
	Values []string `json:"values"`
}

type OrderBy struct {
	// 排序字段，需为 fields 字段中指定返回字段的子集，字段类型为数值类的指标均支持排序
	SortField string `json:"sort_field"`

	// 排序方式，[枚举详情]
	//	枚举列表：{ ASCENDING, DESCENDING }
	SortType string `json:"sort_type"`
}

type GetDailyReportsOptions struct {

	// 广告主帐号 id，有操作权限的帐号 id，不支持代理商 id
	AccountID string `json:"account_id,omitempty"`

	// 获取报表类型级别，获取报表类型级别，腾讯广告平台广告主仅可使用：{REPORT_LEVEL_ADVERTISER, REPORT_LEVEL_CAMPAIGN, REPORT_LEVEL_ADGROUP, REPORT_LEVEL_AD, REPORT_LEVEL_PROMOTED_OBJECT, REPORT_LEVEL_UNION_POSITION, REPORT_LEVEL_CREATIVE_TEMPLATE, REPORT_LEVEL_EXPAND_TARGETING_ADGROUP, REPORT_LEVEL_MATERIAL_VIDEO, REPORT_LEVEL_MATERIAL_IMAGE, REPORT_LEVEL_PRODUCT_CATELOG}，[枚举详情]
	//	枚举列表：{ REPORT_LEVEL_ADVERTISER, REPORT_LEVEL_CAMPAIGN, REPORT_LEVEL_ADGROUP, REPORT_LEVEL_AD, REPORT_LEVEL_PROMOTED_OBJECT, REPORT_LEVEL_UNION_POSITION, REPORT_LEVEL_CREATIVE_TEMPLATE, REPORT_LEVEL_EXPAND_TARGETING_ADGROUP, REPORT_LEVEL_MATERIAL_VIDEO, REPORT_LEVEL_MATERIAL_IMAGE, REPORT_LEVEL_PRODUCT_CATELOG, REPORT_LEVEL_ADVERTISER_WECHAT, REPORT_LEVEL_CAMPAIGN_WECHAT, REPORT_LEVEL_ADGROUP_WECHAT, REPORT_LEVEL_AD_WECHAT }
	//	微信公众账号逻辑
	//
	//	仅可使用：{REPORT_LEVEL_ADVERTISER_WECHAT, REPORT_LEVEL_CAMPAIGN_WECHAT, REPORT_LEVEL_ADGROUP_WECHAT, REPORT_LEVEL_AD_WECHAT}
	Level string `json:"level,omitempty"`

	// 日期范围，最早支持查询 1 年内（365 天）的数据
	DateRange *DateRange `json:"date_range,omitempty"`

	// 过滤条件，若此字段不传，或传空则视为无限制条件，若获取联盟广告位信息此字段必填，详见 [过滤条件]
	//	数组最小长度 1，最大长度 5
	Filtering []Filtering `json:"filtering,omitempty"`

	// 聚合参数，见 [聚合规则]
	//	数组最小长度 1，最大长度 4
	//	字段长度最大 255 字节
	//	微信公众账号逻辑
	//
	//	不支持
	GroupBy []string `json:"group_by,omitempty"`

	// 排序字段
	//	数组长度为 1
	OrderBy []OrderBy `json:"order_by,omitempty"`

	// 	搜索页码，默认值：1
	//	最小值 1，最大值 99999
	Page string `json:"page,omitempty"`

	// 	一页显示的数据条数，默认值：10
	//	最小值 1，最大值 1000
	PageSize string `json:"page_size,omitempty"`

	// 时间口径，[枚举详情]
	//	枚举列表：{ REQUEST_TIME, REPORTING_TIME, ACTIVE_TIME }
	//	微信公众账号逻辑
	//
	//	仅支持 REPORTING_TIME
	TimeLine string `json:"time_line,omitempty"`

	// 	指定返回的字段列表
	//	数组最小长度 1，最大长度 256
	//	字段长度最小 1 字节，长度最大 64 字节
	Fields []string `json:"fields,omitempty"`
}

func (o *GetDailyReportsOptions) uri(accessToken, timestamp, nonce string) (string, error) {
	data, err := json.Marshal(o)
	if err != nil {
		return "", err
	}
	mp := make(map[string]interface{}, 10)
	err = json.Unmarshal(data, &mp)
	if err != nil {
		return "", err
	}
	vs := make(url.Values, len(mp))
	for field, v := range mp {
		switch tv := v.(type) {
		case string:
			vs.Set(field, tv)
		default:
			data, _ = json.Marshal(v)
			vs.Set(field, string(data))
		}
	}
	vs.Set("access_token", accessToken)
	vs.Set("timestamp", timestamp)
	vs.Set("nonce", nonce)
	for field, vues := range vs {
		log.Error.Printf("%s: %v\n", field, vues)
	}
	return "https://api.e.qq.com/v1.3/daily_reports/get?" + vs.Encode(), nil
}

// 分页配置信息
type PageInfo struct {
	Page        int `json:"page"`         // 搜索页码
	PageSize    int `json:"page_size"`    // 一页显示的数据条数
	TotalNumber int `json:"total_number"` // 总条数
	TotalPage   int `json:"total_page"`   // 总页数
}

type DailyReports struct {
	List     []map[string]interface{} `json:"list"`
	PageInfo PageInfo                 `json:"page_info"`
}

// 获取日报表，see: https://developers.e.qq.com/docs/api/insights/ad_insights/daily_reports_get?version=1.1&_preview=1
func GetDailyReports(accessToken string, opts *GetDailyReportsOptions) (*DailyReports, error) {
	timestamp := strconv.FormatInt(Now().Unix(), 10)
	nonce := timestamp + strconv.Itoa(rand.Intn(999999))
	uri, err := opts.uri(accessToken, timestamp, nonce)
	reports := &DailyReports{}
	err = HttpGet(uri, reports)
	if err != nil {
		return nil, err
	}
	return reports, nil
}
