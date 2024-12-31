package models

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/ksuid"
)

const SessionDuration = 24 * time.Hour

type Session struct {
	Id       string
	UserId   int64
	Expires  time.Time
	Count    uint32
	IsActive bool
}

func (s Session) IsExpired() bool {
	return s.Expires.Before(time.Now())
}

type SessionModel struct {
	db      *sql.DB
	timeout time.Duration
}

func NewSessionModel(db *sql.DB, timeout time.Duration) *SessionModel {
	return &SessionModel{db: db, timeout: timeout}
}

func (model *SessionModel) CreateNewSession(c context.Context, userId int64) (*Session, error) {
	slog.Info("timeout", slog.Any("timeout", model.timeout))
	ctx, cancel := context.WithTimeout(c, model.timeout)
	defer cancel()

	newId := ksuid.New().String()
	expires := time.Now().Add(SessionDuration)

	result, err := db.ExecContext(ctx, "INSERT INTO sessions (id, user_id, expires) VALUES (?, ?, ?)", newId, userId, expires.Format(time.RFC3339))
	if err != nil {
		return &Session{}, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return &Session{}, err
	}
	if rows != 1 {
		return &Session{}, fmt.Errorf("expected to affect 1 row, affected %d", rows)
	}

	newSession := Session{
		Id:      newId,
		UserId:  userId,
		Expires: expires,
	}

	slog.Info("newSession", slog.Any("newSession", newSession))

	return &newSession, nil
}

func (model *SessionModel) GetById(c context.Context, sessionId string) (*Session, error) {
	ctx, cancel := context.WithTimeout(c, model.timeout)
	defer cancel()

	var session Session
	var expires string
	err := model.db.QueryRowContext(ctx, "SELECT id, user_id, expires, count, is_active FROM sessions WHERE id = ?", sessionId).Scan(&session.Id, &session.UserId, &expires, &session.Count, &session.IsActive)
	if err != nil {
		return &Session{}, err
	}

	t, err := time.Parse(time.RFC3339, expires)
	if err != nil {
		return &Session{}, fmt.Errorf("error converting string into time: %w", err)
	}

	// TODO: check if it's expired
	// TODO: check if it's active

	session.Expires = t

	return &session, nil
}

// func (model *SessionModel) GetCountByUserId(c context.Context, session *Session) (uint32, error) {
// 	ctx, cancel := context.WithTimeout(c, model.timeout)
// 	defer cancel()

// 	var count uint32
// 	err := db.QueryRowContext(ctx, "SELECT count FROM sessions WHERE user_id = ?", session.UserId).Scan(&count)
// 	if err != nil {
// 		return 0, err
// 	}

// 	return count, nil
// }

func (model *SessionModel) SetCount(c context.Context, sessionId string, count uint32) error {
	ctx, cancel := context.WithTimeout(c, model.timeout)
	defer cancel()

	err := ExecOneContext(ctx, "UPDATE sessions SET count = ? WHERE id = ?", count, sessionId)
	if err != nil {
		return err
	}

	return nil
}
