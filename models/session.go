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

type SessionModel struct {
	db      *sql.DB
	timeout time.Duration
}

func NewSessionModel(db *sql.DB, timeout time.Duration) *SessionModel {
	return &SessionModel{db: db, timeout: timeout}
}

type Session struct {
	Id      string
	UserId  int64
	Expires time.Time
}

func (s Session) IsExpired() bool {
	return s.Expires.Before(time.Now())
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
	err := model.db.QueryRowContext(ctx, "SELECT id, user_id, expires FROM sessions WHERE id = ?", sessionId).Scan(&session.Id, &session.UserId, &expires)
	if err != nil {
		return &Session{}, err
	}

	t, err := time.Parse(time.RFC3339, expires)
	if err != nil {
		return &Session{}, fmt.Errorf("error converting string into time: %w", err)
	}

	session.Expires = t

	return &session, nil
}
