package models

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

const SessionDuration = 24 * time.Hour

type Session struct {
	Id           string
	TimeAccessed time.Time `sql:"time_accessed"`
	Value        map[string]interface{}
}

type SessionModel struct {
	db      *sql.DB
	timeout time.Duration
}

func NewSessionModel(db *sql.DB, timeout time.Duration) *SessionModel {
	return &SessionModel{db: db, timeout: timeout}
}

func (model *SessionModel) CreateNewSession(c context.Context, sessionId string) (*Session, error) {
	slog.Info("timeout", slog.Any("timeout", model.timeout))
	ctx, cancel := context.WithTimeout(c, model.timeout)
	defer cancel()

	now := time.Now()
	value := make(map[string]interface{})
	var err error

	jsonValue, err := json.Marshal(value)
	if err != nil {
		return &Session{}, err
	}

	result, err := model.db.ExecContext(
		ctx,
		"INSERT INTO sessions (id, time_accessed, value) VALUES (?, ?, ?)",
		sessionId,
		now.Format(time.RFC3339),
		string(jsonValue),
	)

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
		Id:           sessionId,
		TimeAccessed: now,
		Value:        value,
	}

	slog.Info("newSession", slog.Any("newSession", newSession))

	return &newSession, nil
}

func (model *SessionModel) GetById(c context.Context, sessionId string) (*Session, error) {
	ctx, cancel := context.WithTimeout(c, model.timeout)
	defer cancel()

	var session Session
	var timeAccessed string
	var value string
	err := model.db.
		QueryRowContext(ctx, "SELECT id, time_accessed, value FROM sessions WHERE id = ? AND is_deleted = 0", sessionId).
		Scan(&session.Id, &timeAccessed, &value)

	if err != nil {
		return &Session{}, err
	}

	t, err := time.Parse(time.RFC3339, timeAccessed)
	if err != nil {
		return &Session{}, fmt.Errorf("error converting string into time: %w", err)
	}

	// TODO: check if it's expired
	// TODO: check if it's active

	var valueMap map[string]interface{}
	err = json.Unmarshal([]byte(value), &valueMap)
	if err != nil {
		return &Session{}, err
	}

	session.TimeAccessed = t
	session.Value = valueMap

	return &session, nil
}

func (model *SessionModel) DeleteById(c context.Context, sessionId string) error {
	ctx, cancel := context.WithTimeout(c, model.timeout)
	defer cancel()

	return ExecOneContext(ctx, "UPDATE sessions SET is_deleted = true WHERE id = ?", sessionId)
}

func (model *SessionModel) UpdateSession(c context.Context, sessionId string, value map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(c, model.timeout)
	defer cancel()

	now := time.Now()

	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	slog.Debug("update session", slog.Any("jsonValue", jsonValue))

	return ExecOneContext(ctx, "UPDATE sessions SET time_accessed = ?, value = ? WHERE id = ?", now.Format(time.RFC3339), string(jsonValue), sessionId)
}
