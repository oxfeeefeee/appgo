package userSystem

import (
	//log "github.com/Sirupsen/logrus"
	"github.com/oxfeeefeee/appgo"
)

func (u *UserSystem) ReadUser(id appgo.Id) (*UserData, error) {
	m := &UserModel{Id: id}
	if db := U.db.First(m); db.Error != nil {
		if db.RecordNotFound() {
			return nil, nil
		} else {
			return nil, db.Error
		}
	}
	return dbModelToData(m), nil
}

func (u *UserSystem) CreateUser(user *UserData) (appgo.Id, error) {
	if err := U.db.Save(dataToDbModel(user)).Error; err != nil {
		return 0, err
	} else {
		return user.Id, nil
	}
}

func (u *UserSystem) UpdateUser(id appgo.Id, user *UserData) error {
	um := &UserModel{Id: id}
	if err := U.db.Model(um).Updates(dataToDbModel(user)).Error; err != nil {
		return err
	}
	return nil
}

func (u *UserSystem) DeleteUser(id appgo.Id) error {
	return U.db.Delete(&UserModel{Id: id}).Error
}

func (u *UserSystem) UserCount() (int, error) {
	var count int
	if err := U.db.Model(&UserModel{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (u *UserSystem) Search(keyword string, limit int) ([]*UserData, error) {
	var users []*UserModel
	where := "`id` LIKE ? OR `nickname` LIKE ?"
	kw := "%" + keyword + "%"
	db := U.db.Where(where, kw, kw).Limit(limit).Find(&users)
	if db.Error != nil {
		return nil, db.Error
	} else {
		return dbModelsToData(users), nil
	}
}

func (u *UserSystem) ListUsers(offset, limit int) ([]*UserData, error) {
	var users []*UserModel
	if err := U.db.Order("id desc").
		Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, err
	} else {
		return dbModelsToData(users), nil
	}
}

func dbModelsToData(models []*UserModel) []*UserData {
	ret := make([]*UserData, 0, len(models))
	for _, m := range models {
		ret = append(ret, dbModelToData(m))
	}
	return ret
}
