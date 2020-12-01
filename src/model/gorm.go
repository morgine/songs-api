package model

import "gorm.io/gorm"

type Gorm struct {
	db *gorm.DB
}

func NewGorm(db *gorm.DB) *Gorm {
	return &Gorm{db: db}
}

func (g *Gorm) First(model interface{}, conditions ...Condition) error {
	err := HandleConditions(g.db, conditions...).First(model).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return nil
}

func (g *Gorm) Create(model interface{}) error {
	return g.db.Create(model).Error
}

func (g *Gorm) Save(model interface{}, c ...Condition) error {
	err := HandleConditions(g.db, c...).Save(model).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return nil
}

func (g *Gorm) Updates(model interface{}, conditions ...Condition) error {
	err := HandleConditions(g.db, conditions...).Updates(model).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	} else {
		return nil
	}
}

func (g *Gorm) Count(model interface{}, c ...Condition) (total int64, err error) {
	err = HandleConditions(g.db.Model(model), c...).Count(&total).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	} else {
		return total, nil
	}
}

func (g *Gorm) Find(models interface{}, conditions ...Condition) error {
	err := HandleConditions(g.db, conditions...).Find(models).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return nil
}

func (g *Gorm) Delete(model interface{}, conditions ...Condition) (err error) {
	err = HandleConditions(g.db, conditions...).Delete(model).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return nil
}
