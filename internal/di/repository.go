package di

import (
	"github.com/bowouhuy/bookstore-api/internal/entities/repository"
	"github.com/bowouhuy/bookstore-api/internal/infrastructure/datastore"
)

// GetBaseRepo returns BaseRepo instance.
func GetBaseRepo() *datastore.BaseRepo {
	cfg := GetConfig()
	return datastore.NewBaseRepo(datastore.GetDatabaseMaster(&cfg.MasterDB), datastore.GetDatabaseSlave(&cfg.SlaveDB))
}

func GetHealthRepo() repository.HealthRepo {
	return datastore.NewHealthRepo(GetBaseRepo())
}

func GetBookRepo() repository.BookRepo {
	return datastore.NewBookRepo(GetBaseRepo())
}

func GetUserRepo() repository.UserRepo {
	return datastore.NewUserRepo(GetBaseRepo())
}
