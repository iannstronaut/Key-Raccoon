package services

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/models"
)

var HealthCheckURL = "https://www.google.com"

type ProxyService struct {
	proxyRepo *repositories.ProxyRepository
}

func NewProxyService(proxyRepo *repositories.ProxyRepository) *ProxyService {
	return &ProxyService{proxyRepo: proxyRepo}
}

func (s *ProxyService) AddProxy(proxyURL, proxyType, username, password string) (*models.Proxy, error) {
	parsed, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy url: %w", err)
	}

	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("invalid proxy url: scheme and host are required")
	}

	proxy := &models.Proxy{
		URL:      proxyURL,
		Type:     proxyType,
		Username: username,
		Password: password,
		IsActive: true,
		Status:   "unknown",
	}

	if err := s.proxyRepo.Create(proxy); err != nil {
		return nil, fmt.Errorf("failed to create proxy: %w", err)
	}

	go s.CheckProxyHealth(proxy.ID)

	return proxy, nil
}

func (s *ProxyService) GetProxy(proxyID uint) (*models.Proxy, error) {
	return s.proxyRepo.GetByID(proxyID)
}

func (s *ProxyService) GetAllProxies(limit, offset int) ([]models.Proxy, int64, error) {
	return s.proxyRepo.GetAll(limit, offset)
}

func (s *ProxyService) DeleteProxy(proxyID uint) error {
	return s.proxyRepo.Delete(proxyID)
}

func (s *ProxyService) CheckProxyHealth(proxyID uint) error {
	proxy, err := s.proxyRepo.GetByID(proxyID)
	if err != nil {
		return err
	}

	isHealthy := s.isProxyHealthy(proxy)

	status := "unhealthy"
	if isHealthy {
		status = "healthy"
	}

	_ = s.proxyRepo.UpdateStatus(proxyID, status)
	_ = s.proxyRepo.UpdateLastCheck(proxyID)

	return nil
}

func (s *ProxyService) isProxyHealthy(proxy *models.Proxy) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	proxyURL, err := url.Parse(proxy.URL)
	if err != nil {
		return false
	}

	if proxy.Username != "" && proxy.Password != "" {
		proxyURL.User = url.UserPassword(proxy.Username, proxy.Password)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 5 * time.Second,
		}).DialContext,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, HealthCheckURL, nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound
}

func (s *ProxyService) CheckAllProxies() error {
	proxies, _, err := s.proxyRepo.GetAll(1000, 0)
	if err != nil {
		return err
	}

	for _, proxy := range proxies {
		go func(id uint) {
			_ = s.CheckProxyHealth(id)
		}(proxy.ID)
	}

	return nil
}

func (s *ProxyService) GetHealthyProxy() (*models.Proxy, error) {
	return s.proxyRepo.GetHealthy()
}

func (s *ProxyService) GetProxyWithFallback() (*models.Proxy, error) {
	proxy, err := s.GetHealthyProxy()
	if err == nil {
		return proxy, nil
	}

	active, err := s.proxyRepo.GetActive()
	if err != nil || len(active) == 0 {
		return nil, errors.New("no proxy available")
	}

	return &active[0], nil
}

func (s *ProxyService) RotateProxy(unhealthyProxyID uint) (*models.Proxy, error) {
	_ = s.proxyRepo.UpdateStatus(unhealthyProxyID, "unhealthy")
	return s.GetHealthyProxy()
}

func (s *ProxyService) TestProxy(proxyURL, proxyType, username, password string) (bool, error) {
	proxy := &models.Proxy{
		URL:      proxyURL,
		Type:     proxyType,
		Username: username,
		Password: password,
	}

	return s.isProxyHealthy(proxy), nil
}
