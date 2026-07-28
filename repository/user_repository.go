package repository

import (
	"errors"

	"library-management-backend/database"
	"library-management-backend/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserRepository interface {
	StoreUser(tx *gorm.DB, userDetails *model.User) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	FindByUUID(uuid string) (*model.User, error)
	FindByID(id uint) (*model.User, error)
	ExistsByEmail(email string) (bool, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository() UserRepository {
	return &userRepository{db: database.LibraryManagementDB}
}

func (r *userRepository) getDB() *gorm.DB {
	if r.db != nil {
		return r.db
	}
	return database.LibraryManagementDB
}

func (r *userRepository) StoreUser(tx *gorm.DB, userDetails *model.User) (*model.User, error) {
	if tx == nil {
		tx = r.getDB()
	}
	if err := tx.Create(userDetails).Error; err != nil {
		logrus.Error("StoreUserWithtx DB Error: ", err)
		return nil, err
	}
	logrus.Infof("StoreUserWithtx success, UserID: %d", userDetails.ID)
	return userDetails, nil
}

func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.getDB().Where("email = ? AND deleted_at IS NULL", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logrus.Error("FindByEmail DB Error: ", err)
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByUUID(uuid string) (*model.User, error) {
	var user model.User
	err := r.getDB().Where("uuid = ? AND deleted_at IS NULL", uuid).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logrus.Error("FindByUUID DB Error: ", err)
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := r.getDB().Where("id = ? AND deleted_at IS NULL", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logrus.Error("FindByID DB Error: ", err)
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.getDB().Model(&model.User{}).Where("email = ? AND deleted_at IS NULL", email).Count(&count).Error
	if err != nil {
		logrus.Error("ExistsByEmail DB Error: ", err)
		return false, err
	}
	return count > 0, nil
}
