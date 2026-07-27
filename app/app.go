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
	BookController *controller.BookController
}

type Services struct {
	AuthService service.AuthService
	BookService service.BookService
}

type Repositories struct {
	UserRepository repository.UserRepository
	BookRepository repository.BookRepository
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
		BookRepository: repository.NewBookRepository(),
	}
}

func initServices(repos *Repositories) *Services {
	return &Services{
		AuthService: service.NewAuthService(repos.UserRepository),
		BookService: service.NewBookService(repos.BookRepository),
	}
}

func initControllers(services *Services) *ConsentRequestController {
	return &ConsentRequestController{
		AuthController: controller.NewAuthController(services.AuthService),
		BookController: controller.NewBookController(services.BookService),
	}
}
