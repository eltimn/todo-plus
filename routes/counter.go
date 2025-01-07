package routes

import (
	"net/http"
	"sync/atomic"

	"eltimn/todo-plus/pkg/router"
	"eltimn/todo-plus/pkg/session"
	"eltimn/todo-plus/web/pages"

	"github.com/Jeffail/gabs/v2"
	datastar "github.com/starfederation/datastar/sdk/go"
)

var globalCounter atomic.Uint32

const SessionCountKey = "count"

type counterEnv struct{}

func userCountVal(req *http.Request) (uint32, session.Session, error) {
	sess := contextSession(req)

	var count uint32
	countFloat, ok := sess.Get(SessionCountKey).(float64)
	if ok {
		count = uint32(countFloat)
	}

	return count, sess, nil
}

func (env *counterEnv) index(rw http.ResponseWriter, req *http.Request) error {
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
	sess.Set(SessionCountKey, val)

	update := gabs.New()
	updateGlobal(update)
	update.Set(val, "user")

	datastar.NewSSE(rw, req).MarshalAndMergeSignals(update)
	return nil
}

func (env *counterEnv) renderCounterPage(rw http.ResponseWriter, req *http.Request, isFullPage bool) error {
	usr := contextUser(req)
	sess := contextSession(req)
	nonce := contextNonce(req)

	var count uint32
	countFloat, ok := sess.Get(SessionCountKey).(float64)
	if ok {
		count = uint32(countFloat)
	}

	signals := pages.TemplCounterSignals{
		Global: globalCounter.Load(),
		User:   count,
	}

	if isFullPage {
		pages.CounterPage(usr, signals, nonce).Render(req.Context(), rw)
	} else {
		pages.CounterPartial(signals).Render(req.Context(), rw)
	}

	return nil
}

func updateGlobal(signals *gabs.Container) {
	signals.Set(globalCounter.Add(1), "global")
}

func counterRoutes(rtr *router.Router) {
	env := &counterEnv{}

	rtr.Group(func(r *router.Router) {
		r.Use(mustBeLoggedInMiddleware)
		r.Get("/counter", env.index)
		r.Post("/counter/increment/global", env.incrementGlobal)
		r.Post("/counter/increment/user", env.incrementUser)
	})
}
