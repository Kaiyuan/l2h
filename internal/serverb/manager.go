package serverb

type Manager struct {
	db *Database
}

func NewManager(dbPath string) *Manager {
	db, err := NewDatabase(dbPath)
	if err != nil {
		panic(err)
	}
	return &Manager{db: db}
}

func (m *Manager) GetAdminInfo() (*AdminInfo, error) {
	return m.db.GetAdminInfo()
}

func (m *Manager) ListBindings() ([]*Binding, error) {
	return m.db.GetBindings()
}

func (m *Manager) AddBinding(path string, port int, password string) error {
	return m.db.AddBinding(path, port, password)
}

func (m *Manager) DeleteBinding(id int) error {
	return m.db.DeleteBinding(id)
}

func (m *Manager) SetServerInfo(serverURL, apiKey string) error {
	return m.db.SetServerInfo(serverURL, apiKey)
}
