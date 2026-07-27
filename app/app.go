package app

import (
	"library-management-backend/controller"
	"library-management-backend/repository"
	"library-management-backend/service"
)

type App struct {
	ConsentRequestController *ConsentRequestController
}

type ConsentRequestController struct {
	AuthController *controller.AuthController
}

type Services struct {
	AuthService service.AuthService
}

type Repositories struct {
	UserRepository repository.UserRepository
}

func InitApp() *App {
	repos := initRepositories()
	services := initServices(repos)
	controllers := initControllers(services)

	return &App{
		ConsentRequestController: controllers,
	}
}

func initRepositories() *Repositories {
	return &Repositories{
		UserRepository: repository.NewUserRepository(),
	}
}

func initServices(repos *Repositories) *Services {
	return &Services{
		AuthService: service.NewAuthService(repos.UserRepository),
	}
}

func initControllers(services *Services) *ConsentRequestController {
	return &ConsentRequestController{
		AuthController: controller.NewAuthController(services.AuthService),
	}
}
