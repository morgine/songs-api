package model

import "gorm.io/gorm"

type Model struct {
	db *gorm.DB
}

func NewModel(db *gorm.DB) *Model {
	return &Model{db: db}
}

func (m *Model) First(model interface{}, conditions ...Condition) error {
	err := HandleConditions(m.db, conditions...).First(model).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return nil
}

func (m *Model) Create(model interface{}) error {
	return m.db.Create(model).Error
}

func (m *Model) Save(model interface{}, c ...Condition) error {
	err := HandleConditions(m.db, c...).Save(model).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return nil
}

func (m *Model) Updates(model interface{}, conditions ...Condition) error {
	err := HandleConditions(m.db, conditions...).Updates(model).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	} else {
		return nil
	}
}

func (m *Model) Count(model interface{}, c ...Condition) (total int64, err error) {
	err = HandleConditions(m.db.Model(model), c...).Count(&total).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	} else {
		return total, nil
	}
}

func (m *Model) Find(models interface{}, conditions ...Condition) error {
	err := HandleConditions(m.db, conditions...).Find(models).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return nil
}

func (m *Model) Delete(model interface{}, conditions ...Condition) (err error) {
	err = HandleConditions(m.db, conditions...).Delete(model).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return nil
}
