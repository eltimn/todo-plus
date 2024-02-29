package middleware

import (
	"eltimn/todo-plus/pkg/router"
	"log/slog"
	"net/http"

	"github.com/segmentio/ksuid"
)

type SessionCookieConfig struct {
	Secure   bool
	HTTPOnly bool
}

type SessionCookieOption func(*SessionCookieConfig)

func SessionCookie(opts ...SessionCookieOption) router.Middleware {
	cfg := SessionCookieConfig{
		Secure:   true,
		HTTPOnly: true,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			slog.Info("SessionCookie middleware")
			id := sessionID(req)
			if id == "" {
				id = ksuid.New().String()
				http.SetCookie(w, &http.Cookie{Name: "sessionID", Value: id, Secure: cfg.Secure, HttpOnly: cfg.HTTPOnly})
			}
			next.ServeHTTP(w, req)
		})
	}
}

func SessionCookieWithSecure(secure bool) SessionCookieOption {
	return func(m *SessionCookieConfig) {
		m.Secure = secure
	}
}

func SessionCookieWithHTTPOnly(httpOnly bool) SessionCookieOption {
	return func(m *SessionCookieConfig) {
		m.HTTPOnly = httpOnly
	}
}

func sessionID(r *http.Request) (id string) {
	cookie, err := r.Cookie("sessionID")
	if err != nil {
		return
	}
	slog.Info("SessionID", slog.String("id", cookie.Value))
	return cookie.Value
}
