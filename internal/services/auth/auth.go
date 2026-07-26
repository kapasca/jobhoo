package auth

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"golang.org/x/crypto/bcrypt"

	"jobhoo/internal/repository"
)

const SessionCookieName = "jobhoo_session"
const SessionTTL = 7 * 24 * time.Hour

type Service struct {
	sessions *repository.SessionRepo
}

func NewService(sessions *repository.SessionRepo) *Service {
	return &Service{sessions: sessions}
}

func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// CreateSession generates a random session id, persists it, and returns it
// so the caller can set it as a cookie.
func (s *Service) CreateSession(userID int64) (string, time.Time, error) {
	id, err := randomToken(32)
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().Add(SessionTTL)
	if err := s.sessions.Create(id, userID, expiresAt); err != nil {
		return "", time.Time{}, err
	}
	return id, expiresAt, nil
}

func (s *Service) UserIDFromSession(sessionID string) (int64, error) {
	return s.sessions.GetUserID(sessionID)
}

func (s *Service) DestroySession(sessionID string) error {
	return s.sessions.Delete(sessionID)
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
