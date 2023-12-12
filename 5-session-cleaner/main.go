//////////////////////////////////////////////////////////////////////
//
// Given is a SessionManager that stores session information in
// memory. The SessionManager itself is working, however, since we
// keep on adding new sessions to the manager our program will
// eventually run out of memory.
//
// Your task is to implement a session cleaner routine that runs
// concurrently in the background and cleans every session that
// hasn't been updated for more than 5 seconds (of course usually
// session times are much longer).
//
// Note that we expect the session to be removed anytime between 5 and
// 7 seconds after the last update. Also, note that you have to be
// very careful in order to prevent race conditions.
//

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

const sessionTTL time.Duration = 5 * time.Second

// SessionManager keeps track of all sessions from creation, updating
// to destroying.
type SessionManager struct {
	sessions map[string]Session
	sync     sync.Mutex
}

// Session stores the session's data
type Session struct {
	Data           map[string]interface{}
	ExpirationTime time.Time
	ctx            context.Context
	cancelFunc     context.CancelFunc
}

// NewSessionManager creates a new sessionManager
func NewSessionManager() *SessionManager {
	m := &SessionManager{
		sessions: make(map[string]Session),
	}

	return m
}

// CreateSession creates a new session and returns the sessionID
func (m *SessionManager) CreateSession() (string, error) {
	m.sync.Lock()
	defer m.sync.Unlock()

	sessionID, err := MakeSessionID()
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithCancel(context.Background())
	session := Session{
		Data:           make(map[string]interface{}),
		ExpirationTime: time.Now().Add(sessionTTL),
		ctx:            ctx,
		cancelFunc:     cancel,
	}

	m.sessions[sessionID] = session
	go m.CleanSession(sessionID, session)

	return sessionID, nil
}

func (m *SessionManager) removeSession(sessionID string) {
	m.sync.Lock()
	defer m.sync.Unlock()

	delete(m.sessions, sessionID)
}

func (m *SessionManager) CleanSession(sessionID string, session Session) {
	ctx := session.ctx

	select {
	case <-time.After(time.Until(session.ExpirationTime)):
		select {
		case <-ctx.Done():
			fmt.Println("canceling by update")
		default:
			fmt.Println("removing")
			m.removeSession(sessionID)
		}
	case <-ctx.Done():
		fmt.Println("canceling by update")
	}

}

// ErrSessionNotFound returned when sessionID not listed in
// SessionManager
var ErrSessionNotFound = errors.New("SessionID does not exists")

// GetSessionData returns data related to session if sessionID is
// found, errors otherwise
func (m *SessionManager) GetSessionData(sessionID string) (map[string]interface{}, error) {
	m.sync.Lock()
	defer m.sync.Unlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}
	return session.Data, nil
}

// UpdateSessionData overwrites the old session data with the new one
func (m *SessionManager) UpdateSessionData(sessionID string, data map[string]interface{}) error {
	m.sync.Lock()
	defer m.sync.Unlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return ErrSessionNotFound
	}

	// Hint: you should renew expiry of the session here

	session.cancelFunc()

	ctx, cancel := context.WithCancel(context.Background())

	session.ExpirationTime = time.Now().Add(sessionTTL)
	session.ctx = ctx
	session.cancelFunc = cancel

	go m.CleanSession(sessionID, session)

	return nil
}

func main() {
	// Create new sessionManager and new session
	m := NewSessionManager()
	sID, err := m.CreateSession()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Created new session with ID", sID)

	// Update session data
	data := make(map[string]interface{})
	data["website"] = "longhoang.de"

	err = m.UpdateSessionData(sID, data)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Update session data, set website to longhoang.de")

	// Retrieve data from manager again
	updatedData, err := m.GetSessionData(sID)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Get session data:", updatedData)
}
