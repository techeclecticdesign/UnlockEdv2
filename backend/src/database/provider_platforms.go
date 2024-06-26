package database

import (
	"UnlockEdv2/src/models"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
)

func (db *DB) GetAllProviderPlatforms(page, perPage int) (int64, []models.ProviderPlatform, error) {
	var platforms []models.ProviderPlatform
	var total int64
	if err := db.Conn.Model(&models.ProviderPlatform{}).Count(&total).Error; err != nil {
		return 0, nil, LogDbError(err, "Failed to count of provider platforms.")
	}
	if err := db.Conn.Offset((page - 1) * perPage).
		Limit(perPage).
		Find(&platforms).Error; err != nil {
		return 0, nil, LogDbError(err, "Failed to get provider platforms.")
	}
	return total, platforms, nil
}

func (db *DB) GetAllActiveProviderPlatforms() ([]models.ProviderPlatform, error) {
	var platforms []models.ProviderPlatform
	if err := db.Conn.Where("state = ?", "active").Find(&platforms).Error; err != nil {
		return nil, LogDbError(err, "Failed to find active provider platforms.")
	}
	return platforms, nil
}

func (db *DB) GetProviderPlatformByID(id int) (*models.ProviderPlatform, error) {
	var platform models.ProviderPlatform
	if err := db.Conn.Where("id = ?", fmt.Sprintf("%d", id)).First(&platform).Error; err != nil {
		return nil, LogDbError(err, "Failed to get provider platform by ID.")
	}
	key, err := platform.DecryptAccessKey()
	if err != nil {
		return nil, err
	}
	platform.AccessKey = key
	return &platform, nil
}

func (db *DB) CreateProviderPlatform(platform *models.ProviderPlatform) (*models.ProviderPlatform, error) {

	key, err := platform.EncryptAccessKey()
	if err != nil {
		return nil, err
	}
	platform.AccessKey = key
	log.Debugf("Creating provider platform: %v", platform)
	if err := db.Conn.Create(&platform).Error; err != nil {
		return nil, LogDbError(err, "Failed to create provider platform.")
	}
	newProv := models.ProviderPlatform{}
	if err := db.Conn.Where("name = ?", platform.Name).First(&newProv).Error; err != nil {
		return nil, LogDbError(err, "Failed to get new provider platform.")
	}
	return &newProv, nil
}

func (db *DB) UpdateProviderPlatform(platform *models.ProviderPlatform, id uint) (*models.ProviderPlatform, error) {
	log.Infof("Updating provider platform with ID: %d", id)
	var existingPlatform models.ProviderPlatform
	if err := db.Conn.First(&existingPlatform, id).Error; err != nil {
		return nil, LogDbError(err, "Failed to find provider platform.")
	}
	models.UpdateStruct(&existingPlatform, platform)
	if platform.AccessKey != "" {
		key, err := platform.EncryptAccessKey()
		if err != nil {
			return nil, errors.New("Error encrypting access key, " + err.Error())
		}
		existingPlatform.AccessKey = key
	}
	if platform.State != "" {
		existingPlatform.State = platform.State
	}
	if err := db.Conn.Save(&existingPlatform).Error; err != nil {
		return nil, LogDbError(err, "Failed to update provider platform.")
	}
	return &existingPlatform, nil
}

func (db *DB) DeleteProviderPlatform(id int) error {
	log.Infof("Deleting provider platform with ID: %d", id)
	if err := db.Conn.Delete(&models.ProviderPlatform{}, fmt.Sprintf("%d", id)).Error; err != nil {
		return LogDbError(err, "Failed to delete provider platform.")
	}
	return nil
}
