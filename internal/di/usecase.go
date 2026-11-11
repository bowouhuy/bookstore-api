package di

import "github.com/bowouhuy/bookstore-api/internal/usecases"

func GetHealthUsecase() usecases.HealthUsecase {
	return usecases.NewHealthUsecase(GetHealthRepo())
}

func GetBookUsecase() usecases.BookUsecase {
	return usecases.NewBookUsecase(GetBookRepo())
}

func GetAuthUsecase() usecases.AuthUsecase {
	return usecases.NewAuthUsecase(GetUserRepo(), GetTokenManager())
}
