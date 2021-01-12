package model

type Advert struct {
	ID    int
	Appid string  `gorm:"index"`
	Paid  float64 `gorm:"comment:总支出"`
}
