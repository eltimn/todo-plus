package routes

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"eltimn/todo-plus/models"
	"eltimn/todo-plus/pkg/errs"
	"eltimn/todo-plus/pkg/router"
	"eltimn/todo-plus/pkg/util"
	"eltimn/todo-plus/web/pages/user"
)

const SessionCookieName = "sessionId"

type userEnv struct {
	users interface {
		Signup(c context.Context, req *models.CreateUserInput) (*models.User, error)
		Login(c context.Context, email string, password string) (*models.User, error)
		GetById(c context.Context, userId int64) (*models.User, error)
	}

	sessions interface {
		CreateNewSession(c context.Context, userId int64) (*models.Session, error)
		GetById(c context.Context, sessionId string) (*models.Session, error)
	}

	isSecure bool
}

func (env *userEnv) loginPage(rw http.ResponseWriter, req *http.Request) error {
	usr := contextUser(req)
	return user.Login(usr).Render(req.Context(), rw)
}

func (env *userEnv) loginSubmit(rw http.ResponseWriter, req *http.Request) error {
	email := req.PostFormValue("email")
	slog.Debug("email", slog.String("email", email))
	password := req.PostFormValue("password")
	slog.Debug("password", slog.String("password", password))

	user, err := env.users.Login(req.Context(), email, password)
	if err != nil {
		return err
	}

	session, err := env.sessions.CreateNewSession(req.Context(), user.Id)
	if err != nil {
		return err
	}
	setSessionCookie(rw, session.Id, session.Expires, env.isSecure)

	slog.Info("User logged in", slog.String("username", user.Username))
	util.HxRedirect(rw, "/")
	return nil
}

func (env *userEnv) logout(rw http.ResponseWriter, req *http.Request) error {
	// delete the cookie
	deleteSessionCookie(rw, env.isSecure)
	util.HxRedirect(rw, "/")
	return nil
}

func (env *userEnv) signupPage(rw http.ResponseWriter, req *http.Request) error {
	usr := contextUser(req)
	return user.Signup(usr).Render(req.Context(), rw)
}

func (env *userEnv) signupSubmit(rw http.ResponseWriter, req *http.Request) error {
	// TODO: validation
	newUser := models.CreateUserInput{
		Email:     req.PostFormValue("email"),
		Username:  req.PostFormValue("username"),
		FullName:  req.PostFormValue("full_name"),
		Password:  req.PostFormValue("password"),
		Password2: req.PostFormValue("password2"),
	}

	user, err := env.users.Signup(req.Context(), &newUser)
	if err != nil {
		return err
	}

	slog.Info("User created", slog.String("username", user.Username))

	session, err := env.sessions.CreateNewSession(req.Context(), user.Id)
	if err != nil {
		return err
	}
	setSessionCookie(rw, session.Id, session.Expires, env.isSecure)

	slog.Info("User logged in", slog.String("username", user.Username))

	util.HxRedirect(rw, "/")
	return nil
}

func userRoutes(rtr *router.Router, env *userEnv) {

	rtr.Group(func(r *router.Router) {
		r.Get("/user/login", env.loginPage)
		r.Post("/user/login", env.loginSubmit)
		r.Get("/user/logout", env.logout)
		r.Get("/user/signup", env.signupPage)
		r.Post("/user/signup", env.signupSubmit)
	})
}

func contextUser(req *http.Request) *models.User {
	user, ok := req.Context().Value(ContextUserKey).(*models.User)
	if !ok {
		return &models.User{}
	}
	return user
}

func setSessionCookie(rw http.ResponseWriter, sessionId string, expires time.Time, isSecure bool) {
	http.SetCookie(rw, &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionId,
		Path:     "/",
		Secure:   isSecure,
		HttpOnly: true,
		Expires:  expires,
	})
}

func deleteSessionCookie(rw http.ResponseWriter, isSecure bool) {
	http.SetCookie(rw, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		Secure:   isSecure,
		HttpOnly: true,
		Expires:  time.Now(),
	})
}

func getUserFromCookie(req *http.Request, env *userEnv) (*models.User, error) {
	// get the cookie value from the request
	cookie, err := req.Cookie(SessionCookieName)
	if err != nil {
		slog.Debug("No session cookie found")
		return &models.User{}, nil
	}

	slog.Debug("SessionId", slog.String("id", cookie.Value))

	// get the session from the database
	session, err := env.sessions.GetById(req.Context(), cookie.Value)
	if err != nil {
		slog.Debug("Session not found", errs.ErrAttr(err))
		return &models.User{}, nil
	}

	// get the user from the database
	return env.users.GetById(req.Context(), session.UserId)
}

func userSessionMiddleware(env *userEnv) router.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			// get the user from the cookie
			user, err := getUserFromCookie(req, env)
			if err != nil {
				slog.Debug("User not found", errs.ErrAttr(err))
				next.ServeHTTP(rw, req)
				return
			}

			// add the user to the request context
			ctx := context.WithValue(req.Context(), ContextUserKey, user)
			next.ServeHTTP(rw, req.WithContext(ctx))
		})
	}
}

func mustBeLoggedInMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		usr := contextUser(req)
		if !usr.IsLoggedIn() {
			http.Redirect(rw, req, "/user/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(rw, req)
	})
}
