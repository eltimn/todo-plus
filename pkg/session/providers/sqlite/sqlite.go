package sqlite

import (
	"container/list"
	"context"

	"eltimn/todo-plus/models"
	"eltimn/todo-plus/pkg/session"
)

var pder = &Provider{list: list.New()}

type SessionStore struct {
	session *models.Session
}

func (st *SessionStore) Set(key string, value interface{}) error {
	currentValue := st.session.Value
	currentValue[key] = value
	pder.sessionUpdate(st.session.Id, currentValue)
	return nil
}

func (st *SessionStore) Get(key string) interface{} {
	currentValue := st.session.Value

	if v, ok := currentValue[key]; ok {
		pder.sessionUpdate(st.session.Id, currentValue)
		return v
	}

	return nil
}

func (st *SessionStore) Delete(key string) error {
	currentValue := st.session.Value
	delete(currentValue, key)
	pder.sessionUpdate(st.session.Id, currentValue)
	return nil
}

func (st *SessionStore) SessionID() string {
	return st.session.Id
}

type Provider struct {
	sessions *models.SessionModel
	list     *list.List // gc
}

func (pder *Provider) SessionInit(sid string) (session.Session, error) {
	sess, err := pder.sessions.CreateNewSession(context.TODO(), sid)
	if err != nil {
		return &SessionStore{}, err
	}

	return &SessionStore{session: sess}, nil
}

func (pder *Provider) SessionRead(sid string) (session.Session, error) {
	sess, err := pder.sessions.GetById(context.TODO(), sid)
	if err == nil {
		return &SessionStore{session: sess}, nil
	}

	return pder.SessionInit(sid)
}

func (pder *Provider) SessionDestroy(sid string) error {
	return pder.sessions.DeleteById(context.TODO(), sid)
}

func (pder *Provider) SessionGC(maxlifetime int64) {
	// TODO: implement this
	// pder.lock.Lock()
	// defer pder.lock.Unlock()

	// for {
	// 	element := pder.list.Back()
	// 	if element == nil {
	// 		break
	// 	}
	// 	if (element.Value.(*SessionStore).timeAccessed.Unix() + maxlifetime) < time.Now().Unix() {
	// 		pder.list.Remove(element)
	// 		delete(pder.sessions, element.Value.(*SessionStore).Id)
	// 	} else {
	// 		break
	// 	}
	// }
}

func (pder *Provider) sessionUpdate(sid string, value map[string]interface{}) error {
	return pder.sessions.UpdateSession(context.TODO(), sid, value)
}

func Init(sessions *models.SessionModel) {
	pder.sessions = sessions
	session.Register("sqlite", pder)
}
