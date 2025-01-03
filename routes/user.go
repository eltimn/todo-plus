package routes

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"eltimn/todo-plus/models"
	"eltimn/todo-plus/pkg/errs"
	"eltimn/todo-plus/pkg/router"
	"eltimn/todo-plus/pkg/session"
	"eltimn/todo-plus/web/pages/user"

	datastar "github.com/starfederation/datastar/sdk/go"
)

const SessionCookieName = "sessionId"
const SessionUserIdKey = "UserId"

type userEnv struct {
	users interface {
		Signup(c context.Context, req *models.CreateUserInput) (*models.User, error)
		Login(c context.Context, email string, password string) (*models.User, error)
		GetById(c context.Context, userId int64) (*models.User, error)
	}

	sessions interface {
		UpdateSession(c context.Context, sessionId string, value map[string]interface{}) error
	}

	isSecure bool
}

func (env *userEnv) loginPage(rw http.ResponseWriter, req *http.Request) error {
	usr := contextUser(req)
	signals := &user.LoginSignals{}
	return user.Login(usr, signals).Render(req.Context(), rw)
}

func (env *userEnv) loginSubmit(rw http.ResponseWriter, req *http.Request) error {

	err := req.ParseForm()
	if err != nil {
		return handleLoginError(rw, err)
	}

	email := req.PostFormValue("email")
	slog.Debug("email", slog.String("email", email))
	password := req.PostFormValue("password")
	slog.Debug("password", slog.String("password", password))

	usr, err := env.users.Login(req.Context(), email, password)
	if err != nil {
		return handleLoginError(rw, err)
	}

	sess := contextSession(req)
	sess.Set(SessionUserIdKey, usr.Id)

	slog.Info("User logged in", slog.String("username", usr.Username))

	// util.HxRedirect(rw, "/")
	// http.Redirect(rw, req, "/", http.StatusSeeOther)
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	jsonData := []byte(`{"status":"OK"}`)
	rw.Write(jsonData)

	return nil
}

type formResponse struct {
	Error string `json:"error"`
}

func handleLoginError(rw http.ResponseWriter, err error) error {
	ret := formResponse{
		err.Error(),
	}

	rw.Header().Set("Content-Type", "application/json")
	retJson, err := json.Marshal(ret)
	if err != nil {
		return err
	}

	rw.WriteHeader(http.StatusBadRequest)
	rw.Write(retJson)
	return nil
}

func (env *userEnv) loginDSPost(rw http.ResponseWriter, req *http.Request) error {
	if err := req.ParseMultipartForm(1 << 20); err != nil {
		return errs.BadRequestError("Failed to parse multipart form")
	}

	email := req.PostFormValue("email")
	slog.Debug("email", slog.String("email", email))
	password := req.PostFormValue("password")
	slog.Debug("password", slog.String("password", password))

	// try logging the user in
	usr, err := env.users.Login(req.Context(), email, password)
	switch {
	case err == sql.ErrNoRows:
		return errs.UserNotFoundError
	case err != nil:
		return errs.BadRequestError(err.Error())
	default:
	}

	sess := contextSession(req)
	sess.Set(SessionUserIdKey, usr.Id)

	slog.Debug("User logged in", slog.String("username", usr.Username))

	sse := datastar.NewSSE(rw, req)
	return sse.ExecuteScript("location.replace('/')")
}

func (env *userEnv) loginDSGet(rw http.ResponseWriter, req *http.Request) error {
	err := req.ParseForm()
	if err != nil {
		return errs.BadRequestError("Failed to parse form")
	}

	formData := req.Form
	jsonData, err := json.Marshal(formData)
	if err != nil {
		return errs.InternalServerError("Failed to encode form data as JSON")
	}

	sse := datastar.NewSSE(rw, req)
	return sse.ExecuteScript(fmt.Sprintf(`alert('Form data received via GET request: %s')`, jsonData))
}

func (env *userEnv) logout(rw http.ResponseWriter, req *http.Request) error {
	// destroy the session
	session.Mgr.SessionDestroy(rw, req)
	// util.HxRedirect(rw, "/")
	http.Redirect(rw, req, "/", http.StatusSeeOther)
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

	sess := contextSession(req)
	sess.Set(SessionUserIdKey, user.Id)

	slog.Info("User logged in", slog.String("username", user.Username))

	// util.HxRedirect(rw, "/")
	http.Redirect(rw, req, "/", http.StatusSeeOther)
	return nil
}

func userRoutes(rtr *router.Router, env *userEnv) {

	rtr.Group(func(r *router.Router) {
		r.Get("/user/login", env.loginPage)
		r.Post("/user/login", env.loginSubmit)
		r.Get("/user/login-ds", env.loginDSGet)
		r.Post("/user/login-ds", env.loginDSPost)
		r.Get("/user/logout", env.logout)
		r.Get("/user/signup", env.signupPage)
		r.Post("/user/signup", env.signupSubmit)
	})
}

func contextSession(req *http.Request) session.Session {
	sess, ok := req.Context().Value(ContextSessionKey).(session.Session)
	if !ok {
		return nil
	}
	return sess
}

func contextUser(req *http.Request) *models.User {
	user, ok := req.Context().Value(ContextUserKey).(*models.User)
	if !ok {
		return &models.User{}
	}
	return user
}

func sessionMiddleware(env *userEnv) router.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			// start a session
			sess := session.Mgr.SessionStart(rw, req)

			// add the session to the request context
			ctx := context.WithValue(req.Context(), ContextSessionKey, sess)

			// check if there's a UserId in the session
			var userId int64
			userIdFloat, ok := sess.Get(SessionUserIdKey).(float64)
			if ok {
				userId = int64(userIdFloat)
			}

			if userId != 0 {
				// get the user from the database
				user, err := env.users.GetById(req.Context(), userId)
				if err != nil {
					slog.Debug("Error fetching user", errs.ErrAttr(err))
				} else {
					// add the user to the request context
					ctx = context.WithValue(ctx, ContextUserKey, user)
					ctx = context.WithValue(ctx, ContextIsLoggedInKey, true)
				}
			}

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
