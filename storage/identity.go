package storage

import (
	"context"

	"uuid"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func BindIdentity(ctx context.Context, db *gorm.DB, identity *KaemanIdentity) error {
	return db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(identity).Error
}

func GetIdentity(ctx context.Context, db *gorm.DB, platform Platform, identity string) (*KaemanIdentity, error) {
	var found KaemanIdentity
	if err := db.WithContext(ctx).First(&found, "platform = ? AND identity = ?", platform, identity).Error; err != nil {
		return nil, err
	}
	return &found, nil
}

func ListIdentitiesByUID(ctx context.Context, db *gorm.DB, uid uuid.UUID) ([]KaemanIdentity, error) {
	var identities []KaemanIdentity
	if err := db.WithContext(ctx).Find(&identities, "uid = ?", uid).Error; err != nil {
		return nil, err
	}
	return identities, nil
}

func UnbindIdentity(ctx context.Context, db *gorm.DB, platform Platform, identity string) error {
	return db.WithContext(ctx).
		Where("platform = ? AND identity = ?", platform, identity).
		Delete(&KaemanIdentity{}).Error
}
