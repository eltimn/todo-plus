package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"eltimn/todo-plus/models"
	"eltimn/todo-plus/pkg/errs"
	"eltimn/todo-plus/pkg/router"
	"eltimn/todo-plus/pkg/util"
	"eltimn/todo-plus/web/pages/user"

	datastar "github.com/starfederation/datastar/sdk/go"
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
	signals := &user.LoginSignals{}
	return user.Login(usr, signals).Render(req.Context(), rw)
}

func (env *userEnv) loginSubmit(rw http.ResponseWriter, req *http.Request) error {

	err := req.ParseForm()
	if err != nil {
		return handleLoginError(rw, err)
	}
	fmt.Println("POST")

	email := req.PostFormValue("email")
	slog.Debug("email", slog.String("email", email))
	password := req.PostFormValue("password")
	slog.Debug("password", slog.String("password", password))

	usr, err := env.users.Login(req.Context(), email, password)
	if err != nil {
		return handleLoginError(rw, err)
	}

	session, err := env.sessions.CreateNewSession(req.Context(), usr.Id)
	if err != nil {
		return handleLoginError(rw, err)
	}
	setSessionCookie(rw, session.Id, session.Expires, env.isSecure)

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

	b, err := json.Marshal(req.Form)
	if err != nil {
		return errs.InternalServerError("Failed to encode form data as JSON")
	}
	sse := datastar.NewSSE(rw, req)
	return sse.ExecuteScript(fmt.Sprintf(`alert('Form data received via POST request: %s')`, string(b)))
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
	// delete the cookie
	deleteSessionCookie(rw, env.isSecure)
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
		r.Get("/user/login-ds", env.loginDSGet)
		r.Post("/user/login-ds", env.loginDSPost)
		r.Get("/user/logout", env.logout)
		r.Get("/user/signup", env.signupPage)
		r.Post("/user/signup", env.signupSubmit)
	})
}

func contextSession(req *http.Request) *models.Session {
	user, ok := req.Context().Value(ContextSessionKey).(*models.Session)
	if !ok {
		return &models.Session{}
	}
	return user
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

func sessionMiddleware(env *userEnv) router.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			// get the cookie value from the request
			cookie, err := req.Cookie(SessionCookieName)
			if err != nil {
				slog.Debug("Session cookie not found", errs.ErrAttr(err))
				next.ServeHTTP(rw, req)
				return
			}

			slog.Debug("SessionId", slog.String("sessionId", cookie.Value))

			// get the session from the database
			session, err := env.sessions.GetById(req.Context(), cookie.Value)
			if err != nil {
				slog.Warn("Session not found in the database", errs.ErrAttr(err))
				next.ServeHTTP(rw, req)
				return
			}

			// add the session to the request context
			ctx := context.WithValue(req.Context(), ContextSessionKey, session)

			// get the user from the database
			user, err := env.users.GetById(req.Context(), session.UserId)
			if err != nil {
				slog.Debug("Error fetching user", errs.ErrAttr(err))
			} else {
				// add the user to the request context
				ctx = context.WithValue(ctx, ContextUserKey, user)
				ctx = context.WithValue(ctx, ContextIsLoggedInKey, true)
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
