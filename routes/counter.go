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
	countG, err := sess.Get(SessionCountKey)
	if err != nil {
		return 0, sess, err
	}

	countFloat, ok := countG.(float64)
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
	err := updateGlobal(update)
	if err != nil {
		return err
	}

	return datastar.NewSSE(rw, req).MarshalAndMergeSignals(update)
}

func (env *counterEnv) incrementUser(rw http.ResponseWriter, req *http.Request) error {
	val, sess, err := userCountVal(req)
	if err != nil {
		return err
	}

	val++
	err = sess.Set(SessionCountKey, val)
	if err != nil {
		return err
	}

	update := gabs.New()
	err = updateGlobal(update)
	if err != nil {
		return err
	}

	_, err = update.Set(val, "user")
	if err != nil {
		return err
	}

	return datastar.NewSSE(rw, req).MarshalAndMergeSignals(update)
}

func (env *counterEnv) renderCounterPage(rw http.ResponseWriter, req *http.Request, isFullPage bool) error {
	usr := contextUser(req)
	sess := contextSession(req)
	nonce := contextNonce(req)

	var count uint32
	countG, err := sess.Get(SessionCountKey)
	if err != nil {
		return err
	}

	countFloat, ok := countG.(float64)
	if ok {
		count = uint32(countFloat)
	}

	signals := pages.TemplCounterSignals{
		Global: globalCounter.Load(),
		User:   count,
	}

	if isFullPage {
		return pages.CounterPage(usr, signals, nonce).Render(req.Context(), rw)
	}

	return pages.CounterPartial(signals).Render(req.Context(), rw)
}

func updateGlobal(signals *gabs.Container) error {
	_, err := signals.Set(globalCounter.Add(1), "global")
	return err
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
