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
	AuthController      controller.AuthController
	BookController      controller.BookController
	BorrowController    controller.BorrowController
	DashboardController controller.DashboardController
}

type Services struct {
	AuthService      service.AuthService
	BookService      service.BookService
	BorrowService    service.BorrowService
	DashboardService service.DashboardService
}

type Repositories struct {
	UserRepository      repository.UserRepository
	BookRepository      repository.BookRepository
	BorrowRepository    repository.BorrowRepository
	DashboardRepository repository.DashboardRepository
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
		UserRepository:      repository.NewUserRepository(),
		BookRepository:      repository.NewBookRepository(),
		BorrowRepository:    repository.NewBorrowRepository(),
		DashboardRepository: repository.NewDashboardRepository(),
	}
}

func initServices(repos *Repositories) *Services {
	return &Services{
		AuthService:      service.NewAuthService(repos.UserRepository),
		BookService:      service.NewBookService(repos.BookRepository),
		BorrowService:    service.NewBorrowService(repos.BorrowRepository, repos.BookRepository, repos.UserRepository),
		DashboardService: service.NewDashboardService(repos.DashboardRepository),
	}
}

func initControllers(services *Services) *ConsentRequestController {
	return &ConsentRequestController{
		AuthController:      controller.NewAuthController(services.AuthService),
		BookController:      controller.NewBookController(services.BookService),
		BorrowController:    controller.NewBorrowController(services.BorrowService),
		DashboardController: controller.NewDashboardController(services.DashboardService),
	}
}
