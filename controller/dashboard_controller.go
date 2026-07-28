package controller

import (
	"fmt"

	"library-management-backend/service"
	"library-management-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type DashboardController interface {
	GetDashboardStats(c *gin.Context)
}

type dashboardController struct {
	dashboardService service.DashboardService
}

func NewDashboardController(dashboardService service.DashboardService) DashboardController {
	return &dashboardController{
		dashboardService: dashboardService,
	}
}

func (ctl *dashboardController) GetDashboardStats(c *gin.Context) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			logrus.Errorf("GetDashboardStats@Controller panic: %v", panicInfo)
			utils.InternalServerErrorResponse(c, fmt.Errorf("%v", panicInfo))
		}
	}()

	UserId := c.GetHeader("auth_user_id")
	if UserId == "" {
		utils.UnauthorizedResponse(c, "Authorization failed: User ID is missing")
		return
	}

	stats, err := ctl.dashboardService.GetDashboardStats()
	if err != nil {
		logrus.Error("GetDashboardStats@Error:", err)
		utils.InternalServerErrorResponse(c, err)
		return
	}

	utils.SuccessResponse(c, "Admin dashboard stats fetched successfully", stats)
}
