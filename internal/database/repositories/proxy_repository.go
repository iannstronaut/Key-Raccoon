package repositories

import (
	"errors"

	"gorm.io/gorm"

	"keyraccoon/internal/models"
)

var ErrProxyNotFound = errors.New("proxy not found")

type ProxyRepository struct {
	db *gorm.DB
}

func NewProxyRepository(db *gorm.DB) *ProxyRepository {
	return &ProxyRepository{db: db}
}

func (r *ProxyRepository) Create(proxy *models.Proxy) error {
	return r.db.Create(proxy).Error
}

func (r *ProxyRepository) GetByID(id uint) (*models.Proxy, error) {
	var proxy models.Proxy
	err := r.db.First(&proxy, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProxyNotFound
	}
	return &proxy, err
}

func (r *ProxyRepository) GetAll(limit, offset int) ([]models.Proxy, int64, error) {
	var proxies []models.Proxy
	var total int64

	if err := r.db.Model(&models.Proxy{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Limit(limit).Offset(offset).Order("id ASC").Find(&proxies).Error; err != nil {
		return nil, 0, err
	}

	return proxies, total, nil
}

func (r *ProxyRepository) GetActive() ([]models.Proxy, error) {
	var proxies []models.Proxy
	err := r.db.Where("is_active = ? AND status = ?", true, "healthy").
		Order("id ASC").
		Find(&proxies).Error
	return proxies, err
}

func (r *ProxyRepository) GetHealthy() (*models.Proxy, error) {
	var proxy models.Proxy
	err := r.db.Where("is_active = ? AND status = ?", true, "healthy").
		Order("id ASC").
		First(&proxy).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no healthy proxy available")
		}
		return nil, err
	}
	return &proxy, nil
}

func (r *ProxyRepository) Update(proxy *models.Proxy) error {
	return r.db.Save(proxy).Error
}

func (r *ProxyRepository) UpdateStatus(id uint, status string) error {
	result := r.db.Model(&models.Proxy{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrProxyNotFound
	}
	return nil
}

func (r *ProxyRepository) UpdateLastCheck(id uint) error {
	result := r.db.Model(&models.Proxy{}).Where("id = ?", id).
		Update("last_check", gorm.Expr("CURRENT_TIMESTAMP"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrProxyNotFound
	}
	return nil
}

func (r *ProxyRepository) Disable(id uint) error {
	result := r.db.Model(&models.Proxy{}).Where("id = ?", id).Update("is_active", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrProxyNotFound
	}
	return nil
}

func (r *ProxyRepository) Delete(id uint) error {
	result := r.db.Delete(&models.Proxy{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrProxyNotFound
	}
	return nil
}
