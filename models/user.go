package models

import (
	"context"
	"database/sql"
	"eltimn/todo-plus/pkg/util"
	"time"
)

type UserModel struct {
	db      *sql.DB
	timeout time.Duration
}

func NewUserModel(db *sql.DB, timeout time.Duration) *UserModel {
	return &UserModel{db: db, timeout: timeout}
}

type User struct {
	ID       int64
	Username string
	FullName string
	Email    string
	Password string
}

type CreateUserInput struct {
	Username string
	FullName string
	Email    string
	Password string
}

type BasicUser struct {
	ID       int64
	Username string
	FullName string
	Email    string
}

func (model *UserModel) Signup(c context.Context, req *CreateUserInput) (*BasicUser, error) {
	ctx, cancel := context.WithTimeout(c, model.timeout)
	defer cancel()

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	var lastInsertId int
	query := "INSERT INTO users (username, full_name, email, password) VALUES ($1, $2, $3, $4) RETURNING id"
	err = model.db.QueryRowContext(ctx, query, req.Username, req.FullName, req.Email, hashedPassword).Scan(&lastInsertId)
	if err != nil {
		return &BasicUser{}, err
	}

	res := &BasicUser{
		ID:       int64(lastInsertId),
		Username: req.Username,
		FullName: req.FullName,
		Email:    req.Email,
	}

	return res, nil
}

func (model *UserModel) Login(c context.Context, email string, password string) (*BasicUser, error) {
	ctx, cancel := context.WithTimeout(c, model.timeout)
	defer cancel()

	user := User{}
	query := "SELECT id, username, full_name, email, password FROM users WHERE email = $1"
	err := model.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Username, &user.FullName, &user.Email, &user.Password)
	if err != nil {
		return &BasicUser{}, err
	}

	err = util.CheckPassword(user.Password, password)
	if err != nil {
		return &BasicUser{}, err
	}

	return &BasicUser{
		ID:       user.ID,
		Username: user.Username,
		FullName: user.FullName,
		Email:    user.Email,
	}, nil
}
