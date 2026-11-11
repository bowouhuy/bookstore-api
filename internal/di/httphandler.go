package di

import (
	"github.com/bowouhuy/bookstore-api/config"
	"github.com/bowouhuy/bookstore-api/internal/adapters/httphandler"
)

func GetHealthHandler() httphandler.Handler {
	return httphandler.NewHealthHandler(GetHealthUsecase(), config.Version)
}

func GetRegisterHandler() httphandler.Handler {
	return httphandler.NewRegisterHandler(GetAuthUsecase())
}

func GetLoginHandler() httphandler.Handler {
	return httphandler.NewLoginHandler(GetAuthUsecase())
}

func GetMeHandler() httphandler.Handler {
	return httphandler.NewMeHandler(GetAuthUsecase())
}

func GetCreateBookHandler() httphandler.Handler {
	return httphandler.NewCreateBookHandler(GetBookUsecase())
}

func GetListBooksHandler() httphandler.Handler {
	return httphandler.NewListBooksHandler(GetBookUsecase())
}

func GetGetBookHandler() httphandler.Handler {
	return httphandler.NewGetBookHandler(GetBookUsecase())
}

func GetUpdateBookHandler() httphandler.Handler {
	return httphandler.NewUpdateBookHandler(GetBookUsecase())
}

func GetDeleteBookHandler() httphandler.Handler {
	return httphandler.NewDeleteBookHandler(GetBookUsecase())
}
