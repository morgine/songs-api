package model

//
//type AppDateStatics struct {
//	ID                int
//	Appid             string  `gorm:"index"`               // 公众号 appid
//	Date              int64   `gorm:"index"`               // 统计日期, 所有数据均为历史数据
//	CumulateUser      int     `json:"cumulate_user"`       // 用户总量
//	NewUser           int     `json:"new_user"`            // 新增的用户数量
//	CancelUser        int     `json:"cancel_user"`         // 取消关注的用户数量，new_user减去cancel_user即为净增用户数量
//	PositiveUser      int     `json:"positive_user"`       // 净增用户
//	CancelRate        float64 `json:"cancel_rate"`         // 取关率, cancel_user/cumulate_user
//	ReqSuccCount      int     `json:"req_succ_count"`      // 拉取量
//	ExposureCount     int     `json:"exposure_count"`      // 曝光量
//	ExposureRate      float64 `json:"exposure_rate"`       // 曝光率, exposure_count/req_succ_count
//	ClickCount        int     `json:"click_count"`         // 点击量
//	ClickRate         float64 `json:"click_rate"`          // 点击率, click_count/exposure_count
//	Outcome           int     `json:"outcome"`             // 当日支出(分)
//	TotalOutcome      int     `json:"total_outcome"`       // 总支出(分), 前一日总支出加当日支出
//	Income            int     `json:"income"`              // 当日收入(分)
//	TotalIncome       int     `json:"total_income"`        // 总收入(分), 前一日总收入加当日收入
//	IncomeOutcomeRate float64 `json:"income_outcome_rate"` // 收入支出比率, income/outcome
//	Ecpm              float64 `json:"ecpm"`                // 广告千次曝光收益(分), 1000/exposure_count*income
//}
//
//type AppDateStaticsDB struct {
//	db *gorm.DB
//}
//
//func NewAppDateStaticsDB(db *gorm.DB) *AppDateStaticsDB {
//	return &AppDateStaticsDB{db: db}
//}
//
//func (db *AppDateStaticsDB) getPrevDateStatics(date int64) (*AppDateStatics, error) {
//	db.db
//}
//
//func (db *AppDateStaticsDB) Create(ads *AppDateStatics) error {
//	return db.db.Create(ads).Error
//}
