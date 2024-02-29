package routes

import (
	"context"
	"log/slog"
	"net/http"

	"eltimn/todo-plus/middleware"
	"eltimn/todo-plus/models"
	"eltimn/todo-plus/pkg/router"
	"eltimn/todo-plus/web/pages/user"
)

type UserEnv struct {
	users interface {
		Signup(c context.Context, req *models.CreateUserInput) (*models.BasicUser, error)
		Login(c context.Context, email string, password string) (*models.BasicUser, error)
	}
}

func (env *UserEnv) loginPage(rw http.ResponseWriter, req *http.Request) error {
	return user.Login().Render(req.Context(), rw)
}

func (env *UserEnv) loginSubmit(rw http.ResponseWriter, req *http.Request) error {
	email := req.PostFormValue("email")
	slog.Debug("email", slog.String("email", email))
	password := req.PostFormValue("password")
	slog.Debug("password", slog.String("password", password))

	user, err := env.users.Login(req.Context(), email, password)
	if err != nil {
		return err
	}
	slog.Info("User logged in", slog.String("username", user.Username))
	return nil
}

func (env *UserEnv) logout(rw http.ResponseWriter, req *http.Request) error {
	// TODO: clear session variable
	http.Redirect(rw, req, "/", http.StatusSeeOther)
	return nil
}

func (env *UserEnv) signupPage(rw http.ResponseWriter, req *http.Request) error {
	return user.Signup().Render(req.Context(), rw)
}

func (env *UserEnv) signupSubmit(rw http.ResponseWriter, req *http.Request) error {
	// TODO: error checking and validation
	newUser := models.CreateUserInput{
		Email:    req.PostFormValue("email"),
		Username: req.PostFormValue("username"),
		FullName: req.PostFormValue("full_name"),
		Password: req.PostFormValue("password"),
	}

	user, err := env.users.Signup(req.Context(), &newUser)
	if err != nil {
		return err
	}

	slog.Info("User created", slog.String("username", user.Username))

	// TODO: set session variable
	http.Redirect(rw, req, "/", http.StatusSeeOther)
	return nil
}

func userRoutes(rtr *router.Router, users *models.UserModel) {
	env := &UserEnv{
		users: users,
	}

	rtr.Group(func(r *router.Router) {
		r.Use(middleware.Mid(2))
		r.Get("/user/login", env.loginPage)
		r.Post("/user/login", env.loginSubmit)
		r.Get("/user/logout", env.logout)
		r.Get("/user/signup", env.signupPage)
		r.Post("/user/signup", env.signupSubmit)
	})
}
