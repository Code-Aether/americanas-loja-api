package services

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"sync"
	"time"
)

type JWTKeyManager struct {
	currentKey   string
	previousKey  string
	keyID        string
	rotationTime time.Time
	mutex        sync.RWMutex
}

func NewJWTKeyManager() *JWTKeyManager {
	manager := &JWTKeyManager{}
	manager.generateNewKey()
	return manager
}

func (m *JWTKeyManager) GetCurrentKey() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.currentKey
}

func (m *JWTKeyManager) GetKeyForVerification(keyID string) string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if keyID == m.keyID {
		return m.currentKey
	}

	return m.previousKey
}

func (m *JWTKeyManager) RotateKey() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.previousKey = m.currentKey

	m.generateNewKey()

	log.Printf("JWT key rotated. New key ID : %s", m.keyID[:8])
}

func (m *JWTKeyManager) ShouldRotate() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return time.Since(m.rotationTime) > 7*24*time.Hour
}

func (m *JWTKeyManager) generateNewKey() {
	keyBytes := make([]byte, 64)
	if _, err := rand.Read(keyBytes); err != nil {
		panic("Error generation JWT Key:" + err.Error())
	}

	idBytes := make([]byte, 16)
	_, err := rand.Read(idBytes)
	if err != nil {
		panic("Error generating random bytes: " + err.Error())
	}

	m.currentKey = base64.URLEncoding.EncodeToString(keyBytes)
	m.keyID = base64.URLEncoding.EncodeToString(idBytes)
	m.rotationTime = time.Now()
}

func (m *JWTKeyManager) StartAutoRotation() {
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			if m.ShouldRotate() {
				m.RotateKey()
			}
		}
	}()
}
