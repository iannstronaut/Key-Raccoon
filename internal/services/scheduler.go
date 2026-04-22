package services

import (
	"time"

	"keyraccoon/pkg/logger"
)

type SchedulerService struct {
	proxyService *ProxyService
}

func NewSchedulerService(proxyService *ProxyService) *SchedulerService {
	return &SchedulerService{
		proxyService: proxyService,
	}
}

func (s *SchedulerService) Start() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	logger.Info("Proxy health checker scheduler started")

	for range ticker.C {
		logger.Info("Running proxy health check...")
		if err := s.proxyService.CheckAllProxies(); err != nil {
			logger.Error("Proxy health check failed", "error", err)
		}
	}
}

func (s *SchedulerService) StartInBackground() {
	go s.Start()
}
