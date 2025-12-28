package webrtc

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
)

type Manager struct {
	connections map[string]*Connection
	mu          sync.RWMutex
}

type Connection struct {
	ID     string
	Path   string
	Status string
}

func NewManager() *Manager {
	return &Manager{
		connections: make(map[string]*Connection),
	}
}

func (m *Manager) HandleOffer(offer string, path string) (string, error) {
	// 生成连接ID
	connID := generateConnectionID()
	
	m.mu.Lock()
	m.connections[connID] = &Connection{
		ID:     connID,
		Path:   path,
		Status: "connecting",
	}
	m.mu.Unlock()

	// 这里应该实现真正的 WebRTC offer/answer 交换
	// 为了简化，我们返回一个模拟的 answer
	answer := generateAnswer(offer)
	
	return answer, nil
}

func (m *Manager) GetConnection(id string) (*Connection, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, ok := m.connections[id]
	return conn, ok
}

func generateConnectionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

func generateAnswer(offer string) string {
	// 这里应该实现真正的 WebRTC answer 生成
	// 为了简化，返回一个模拟的 answer
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

