package repo

import (
	"database/sql"
	"go-forum-project/auth-service/internal/entity"
)

type AuthRepository interface {
	GetUserByUsername(username string) (entity.User, error)
}

type UserRepo struct {
	DB *sql.DB
}

func (r *UserRepo) CreateUser(user *entity.User) error {
	_, err := r.DB.Exec(
		"INSERT INTO users (username, email, password) VALUES ($1, $2, $3)",
		user.Username, user.Password, user.Role,
	)
	return err
}

func (r *UserRepo) GetUserByUsername(username string) (*entity.User, error) {
	var user entity.User
	err := r.DB.QueryRow("SELECT id, username, email, password, role FROM users WHERE username = $1", username).
		Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
