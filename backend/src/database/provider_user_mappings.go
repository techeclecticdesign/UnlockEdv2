package database

import (
	"UnlockEdv2/src/models"
	"errors"
)

func (db *DB) CreateProviderUserMapping(providerUserMapping *models.ProviderUserMapping) error {
	return LogDbError(db.Conn.Create(providerUserMapping).Error, "Failed to create provider-user mapping.")
}

func (db *DB) GetProviderUserMappingByExternalUserID(externalUserID string, providerId uint) (*models.ProviderUserMapping, error) {
	var providerUserMapping models.ProviderUserMapping
	if err := db.Conn.Where("external_user_id = ? AND provider_platform_id = ?", externalUserID, providerId).First(&providerUserMapping).Error; err != nil {
		return nil, LogDbError(err, "Failed to get provider-user mapping with external user ID.")
	}
	return &providerUserMapping, nil
}

func (db *DB) GetUserMappingsForProvider(providerId uint) ([]models.ProviderUserMapping, error) {
	var users []models.ProviderUserMapping
	if err := db.Conn.Find(&users).Where("provider_platform_id = ?", providerId).Error; err != nil {
		return nil, LogDbError(err, "Failed to get mappings of users for provider.")
	}
	return users, nil
}

func (db *DB) GetProviderUserMapping(userID, providerID int) (*models.ProviderUserMapping, error) {
	var providerUserMapping models.ProviderUserMapping
	if err := db.Conn.Where("user_id = ? AND provider_platform_id = ?", userID, providerID).First(&providerUserMapping).Error; err != nil {
		return nil, LogDbError(err, "Failed to get provider-user mapping.")
	}
	return &providerUserMapping, nil
}

func (db *DB) UpdateProviderUserMapping(providerUserMapping *models.ProviderUserMapping) error {
	result := db.Conn.Model(&models.ProviderUserMapping{}).Where("id = ?", providerUserMapping.ID).Updates(providerUserMapping)
	if result.Error != nil {
		return LogDbError(result.Error, "Failed to update provider-user mapping.")
	}
	if result.RowsAffected == 0 {
		return LogDbError(errors.New("no matching record"), "Failed to update provider-user mapping.")
	}
	return nil
}

func (db *DB) GetAllProviderMappingsForUser(userID int) ([]models.ProviderUserMapping, error) {
	var providerUserMappings []models.ProviderUserMapping
	if err := db.Conn.Where("user_id = ?", userID).Find(&providerUserMappings).Error; err != nil {
		return nil, LogDbError(err, "Failed to get provider mappings for user.")
	}
	return providerUserMappings, nil
}

func (db *DB) DeleteProviderUserMappingByUserID(userID, providerID int) error {
	result := db.Conn.Model(&models.ProviderUserMapping{}).Where("user_id = ? AND provider_platform_id = ?", userID, providerID).Delete(&models.ProviderUserMapping{})
	if result.Error != nil {
		return LogDbError(result.Error, "Failed to delete provider-user mapping.")
	}
	if result.RowsAffected == 0 {
		return LogDbError(errors.New("no matching record"), "Failed to delete provider-user mapping.")
	}
	return nil
}
