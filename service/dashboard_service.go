package service

import (
	"library-management-backend/dto"
	"library-management-backend/repository"

	"github.com/sirupsen/logrus"
)

type DashboardService interface {
	GetDashboardStats() (*dto.AdminDashboardResponse, error)
}

type dashboardService struct {
	dashboardRepo repository.DashboardRepository
}

func NewDashboardService(dashboardRepo repository.DashboardRepository) DashboardService {
	return &dashboardService{
		dashboardRepo: dashboardRepo,
	}
}

func (s *dashboardService) GetDashboardStats() (*dto.AdminDashboardResponse, error) {
	logrus.Info("GetDashboardStats@Service Started")

	stats, err := s.dashboardRepo.GetDashboardStats()
	if err != nil {
		logrus.Errorf("GetDashboardStats@Service Error: %v", err)
		return nil, err
	}

	logrus.Infof("GetDashboardStats@Service Completed successfully: %+v", stats)
	return stats, nil
}
