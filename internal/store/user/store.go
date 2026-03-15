package user

import (
	usermodel "coffeebase-api/internal/models/user"
	"context"
	"database/sql"
)

type Store interface {
	Create(requestContext context.Context, userInstance *usermodel.User) error
	GetByEmail(requestContext context.Context, email string) (usermodel.User, error)
	GetByID(requestContext context.Context, id int) (usermodel.User, error)
	UpdateAvatar(requestContext context.Context, id int, avatarURL string) error
	UpdateLanguage(requestContext context.Context, id int, language string) error
	GetTotalCount(requestContext context.Context) (int, error)
	GetAll(requestContext context.Context) ([]usermodel.User, error)
	UpdateRole(requestContext context.Context, id int, role usermodel.UserRole) error
	Update(requestContext context.Context, userInstance *usermodel.User) error
}

type postgresStore struct {
	databaseConnection *sql.DB
}

// --- Public ---

func NewStore(databaseConnection *sql.DB) Store {
	return &postgresStore{databaseConnection: databaseConnection}
}
