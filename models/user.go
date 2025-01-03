package models

import (
	"context"
	"database/sql"
	"eltimn/todo-plus/pkg/util"
	"fmt"
	"time"
)

// This is the full user model, that represents all fields stored in the database.
type FullUser struct {
	Id       int64
	Username string
	FullName string
	Email    string
	Password string
}

type CreateUserInput struct {
	Username  string
	FullName  string
	Email     string
	Password  string
	Password2 string
}

// This is the basic user model, which does not include the password.
type User struct {
	Id       int64
	Username string
	FullName string
	Email    string
}

func (user *User) IsLoggedIn() bool {
	return user.Id > 0
}

type UserModel struct {
	db      *sql.DB
	timeout time.Duration
}

func NewUserModel(db *sql.DB, timeout time.Duration) *UserModel {
	return &UserModel{db: db, timeout: timeout}
}

func (model *UserModel) Signup(c context.Context, req *CreateUserInput) (*User, error) {
	ctx, cancel := context.WithTimeout(c, model.timeout)
	defer cancel()

	// check passwords match
	if req.Password != req.Password2 {
		return &User{}, fmt.Errorf("passwords do not match")
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	var lastInsertId int
	query := "INSERT INTO users (username, full_name, email, password) VALUES (?, ?, ?, ?) RETURNING id"
	err = model.db.QueryRowContext(ctx, query, req.Username, req.FullName, req.Email, hashedPassword).Scan(&lastInsertId)
	if err != nil {
		return &User{}, err
	}

	res := &User{
		Id:       int64(lastInsertId),
		Username: req.Username,
		FullName: req.FullName,
		Email:    req.Email,
	}

	return res, nil
}

func (model *UserModel) Login(c context.Context, email string, password string) (*User, error) {
	ctx, cancel := context.WithTimeout(c, model.timeout)
	defer cancel()

	user := FullUser{}
	query := "SELECT id, username, full_name, email, password FROM users WHERE email = ?"
	err := model.db.QueryRowContext(ctx, query, email).Scan(&user.Id, &user.Username, &user.FullName, &user.Email, &user.Password)
	if err != nil {
		return &User{}, err
	}

	err = util.CheckPassword(user.Password, password)
	if err != nil {
		return &User{}, err
	}

	return &User{
		Id:       user.Id,
		Username: user.Username,
		FullName: user.FullName,
		Email:    user.Email,
	}, nil
}

func (model *UserModel) GetById(c context.Context, userId int64) (*User, error) {
	ctx, cancel := context.WithTimeout(c, model.timeout)
	defer cancel()

	user := User{}
	query := "SELECT id, username, full_name, email FROM users WHERE id = ?"
	err := model.db.QueryRowContext(ctx, query, userId).Scan(&user.Id, &user.Username, &user.FullName, &user.Email)
	if err != nil {
		return &User{}, err
	}

	return &user, nil
}
