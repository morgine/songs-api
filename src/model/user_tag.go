package model

import "gorm.io/gorm"

type UserTag struct {
	ID   int
	Name string `gorm:"uniqueIndex"`
}

type UserTagModel struct {
	db *gorm.DB
}

func (m *Model) UserTagModel() *UserTagModel {
	return &UserTagModel{db: m.db}
}

func (m *UserTagModel) Count() (total int64, err error) {
	err = m.db.Model(&UserTag{}).Count(&total).Error
	return
}

func (m *UserTagModel) GetUserTags(limit, offset int) (tags []*UserTag, err error) {
	err = m.db.Order("id asc").Limit(limit).Offset(offset).Find(&tags).Error
	return
}

func (m *UserTagModel) Create(name string) (*UserTag, error) {
	tag := &UserTag{Name: name}
	err := m.db.Create(tag).Error
	if err != nil {
		return nil, err
	} else {
		return tag, nil
	}
}

func (m *UserTagModel) Update(tag *UserTag) error {
	return m.db.Updates(tag).Error
}

func (m *UserTagModel) Delete(id int) error {
	return m.db.Delete(&UserTag{ID: id}).Error
}
