package storage

import (
	"context"

	"uuid"

	"gorm.io/gorm"
)

func CreateUser(ctx context.Context, db *gorm.DB, user *KaemanUser) error {
	return db.WithContext(ctx).Create(user).Error
}

func GetUser(ctx context.Context, db *gorm.DB, uid uuid.UUID) (*KaemanUser, error) {
	var user KaemanUser
	if err := db.WithContext(ctx).First(&user, "uid = ?", uid).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserWithIdentities(ctx context.Context, db *gorm.DB, uid uuid.UUID) (*KaemanUser, error) {
	var user KaemanUser
	if err := db.WithContext(ctx).Preload("Identities").First(&user, "uid = ?", uid).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func UpdateUser(ctx context.Context, db *gorm.DB, user *KaemanUser) error {
	return db.WithContext(ctx).Model(user).Update("nickname", user.Nickname).Error
}

func DeleteUser(ctx context.Context, db *gorm.DB, uid uuid.UUID) error {
	return db.WithContext(ctx).Delete(&KaemanUser{}, "uid = ?", uid).Error
}
