package repo

import (
	"database/sql"
	"go-forum-project/auth-service/internal/entity"
)

type AuthRepository interface {
	CreateUser(user *entity.User) error
	GetUserByUsername(username string) (*entity.User, error)
	GetUserByID(id int) (entity.User, error)
	UserExists(username string) (bool, error)
}

type UserRepo struct {
	DB *sql.DB
}

func (r *UserRepo) CreateUser(user *entity.User) error {
	_, err := r.DB.Exec(
		"INSERT INTO users (username, password, role) VALUES ($1, $2, $3)",
		user.Username, user.Password, user.Role,
	)
	return err
}

func (r *UserRepo) GetUserByUsername(username string) (*entity.User, error) {
	var user entity.User
	err := r.DB.QueryRow("SELECT username, password, role FROM users WHERE username = $1", username).
		Scan(&user.Username, &user.Password, &user.Role)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) GetUserByID(id int) (*entity.User, error) {
	var user entity.User
	err := r.DB.QueryRow("SELECT username, password, role FROM users WHERE id = $1", id).
		Scan(&user.Username, &user.Password, &user.Role)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) UserExists(username string) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(
		"SELECT EXISTS(FROM users WHERE username = $1)",
		username,
	).Scan(&exists)
	return exists, err
}
