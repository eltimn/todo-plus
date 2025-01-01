package session

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/segmentio/ksuid"
)

type Manager struct {
	cookieName  string     // private cookiename
	lock        sync.Mutex // protects session
	provider    Provider
	maxlifetime int64
	isSecure    bool
}

// the global session manager
var Mgr *Manager

func NewManager(providerName, cookieName string, maxlifetime int64, isSecure bool) (*Manager, error) {
	provider, ok := provides[providerName]
	if !ok {
		return nil, fmt.Errorf("session: unknown provider %q (forgotten import?)", providerName)
	}
	return &Manager{provider: provider, cookieName: cookieName, maxlifetime: maxlifetime}, nil
}

func Init(providerName string, cookieName string, isSecure bool) error {
	// initialize the session manager
	var err error
	Mgr, err = NewManager(providerName, cookieName, 3600, isSecure)
	if err != nil {
		return err
	}

	go Mgr.GC()
	return nil
}

type Provider interface {
	SessionInit(sid string) (Session, error)
	SessionRead(sid string) (Session, error)
	SessionDestroy(sid string) error
	SessionGC(maxLifeTime int64)
}

type Session interface {
	Set(key string, value interface{}) error //set session value
	Get(key string) interface{}              //get session value
	Delete(key string) error                 //delete session value
	SessionID() string                       //back current sessionID
}

var provides = make(map[string]Provider)

// Register makes a session provider available by the provided name.
// If a Register is called twice with the same name or if the driver is nil,
// it panics.
func Register(name string, provider Provider) {
	if provider == nil {
		panic("session: Register provider is nil")
	}
	if _, dup := provides[name]; dup {
		panic("session: Register called twice for provider " + name)
	}
	provides[name] = provider
}

func (manager *Manager) sessionId() string {
	// b := make([]byte, 32)
	// if _, err := io.ReadFull(rand.Reader, b); err != nil {
	// 	return ""
	// }
	// return base64.URLEncoding.EncodeToString(b)

	return ksuid.New().String()
}

func (manager *Manager) SessionStart(w http.ResponseWriter, r *http.Request) Session {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	cookie, err := r.Cookie(manager.cookieName)
	var sess Session
	if err != nil || cookie.Value == "" {
		sid := manager.sessionId()
		sess, _ = manager.provider.SessionInit(sid)
		cookie := http.Cookie{
			Name:     manager.cookieName,
			Value:    url.QueryEscape(sid),
			Path:     "/",
			HttpOnly: true,
			Secure:   manager.isSecure,
			MaxAge:   int(manager.maxlifetime),
		}
		http.SetCookie(w, &cookie)
	} else {
		sid, _ := url.QueryUnescape(cookie.Value)
		sess, err = manager.provider.SessionRead(sid)
		if err != nil {
			slog.Warn("error reading session", slog.Any("err", err))
		}
	}

	return sess
}

func (manager *Manager) SessionDestroy(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		return
	} else {
		manager.lock.Lock()
		defer manager.lock.Unlock()
		manager.provider.SessionDestroy(cookie.Value)
		expiration := time.Now()
		cookie := http.Cookie{
			Name:     manager.cookieName,
			Path:     "/",
			HttpOnly: true,
			Secure:   manager.isSecure,
			Expires:  expiration,
			MaxAge:   -1,
		}
		http.SetCookie(w, &cookie)
	}
}

func (manager *Manager) GC() {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	manager.provider.SessionGC(manager.maxlifetime)
	time.AfterFunc(time.Duration(manager.maxlifetime), func() { manager.GC() })
}

// https://astaxie.gitbooks.io/build-web-application-with-golang/content/en/06.2.html
