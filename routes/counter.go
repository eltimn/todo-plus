package routes

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"

	"eltimn/todo-plus/models"
	"eltimn/todo-plus/pkg/router"
	"eltimn/todo-plus/web/pages"

	"github.com/Jeffail/gabs/v2"
	datastar "github.com/starfederation/datastar/sdk/go"
)

var globalCounter atomic.Uint32

type counterEnv struct {
	sessions interface {
		SetCount(c context.Context, sessionId string, count uint32) error
	}
}

func userCountVal(req *http.Request) (uint32, *models.Session, error) {
	session := contextSession(req)

	if session.IsActive {
		return session.Count, session, nil
	}

	return 0, session, fmt.Errorf("session is not active")
}

func (env *counterEnv) index(rw http.ResponseWriter, req *http.Request) error {
	slog.Debug("Request", slog.Any("request", req.URL.Path))
	return env.renderCounterPage(rw, req, true)
}

func (env *counterEnv) incrementGlobal(rw http.ResponseWriter, req *http.Request) error {
	update := gabs.New()
	updateGlobal(update)

	datastar.NewSSE(rw, req).MarshalAndMergeSignals(update)
	return nil
}

func (env *counterEnv) incrementUser(rw http.ResponseWriter, req *http.Request) error {
	val, sess, err := userCountVal(req)
	if err != nil {
		return err
	}

	val++
	sess.Count = val
	if err := env.sessions.SetCount(req.Context(), sess.Id, val); err != nil {
		return err
	}

	update := gabs.New()
	updateGlobal(update)
	update.Set(val, "user")

	datastar.NewSSE(rw, req).MarshalAndMergeSignals(update)
	return nil
}

func (env *counterEnv) renderCounterPage(rw http.ResponseWriter, req *http.Request, isFullPage bool) error {
	usr := contextUser(req)
	session := contextSession(req)

	signals := pages.TemplCounterSignals{
		Global: globalCounter.Load(),
		User:   session.Count,
	}

	if isFullPage {
		pages.CounterPage(usr, signals).Render(req.Context(), rw)
	} else {
		pages.CounterPartial(signals).Render(req.Context(), rw)
	}

	return nil
}

func updateGlobal(signals *gabs.Container) {
	signals.Set(globalCounter.Add(1), "global")
}

func counterRoutes(rtr *router.Router, env *counterEnv) {
	rtr.Group(func(r *router.Router) {
		r.Use(mustBeLoggedInMiddleware)
		r.Get("/counter", env.index)
		r.Post("/counter/increment/global", env.incrementGlobal)
		r.Post("/counter/increment/user", env.incrementUser)
	})
}
