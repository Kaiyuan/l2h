package serverb

import (
	"database/sql"
	"time"

	"l2h/internal/crypto"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	d := &Database{db: db}
	if err := d.initTables(); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Database) initTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS admin (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL,
			password TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS bindings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			path TEXT UNIQUE NOT NULL,
			port INTEGER NOT NULL,
			password TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS server_info (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			server_url TEXT NOT NULL,
			api_key TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range queries {
		if _, err := d.db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

type AdminInfo struct {
	Username string
	Password string
}

func (d *Database) GetAdminInfo() (*AdminInfo, error) {
	var info AdminInfo
	err := d.db.QueryRow("SELECT username, password FROM admin LIMIT 1").Scan(
		&info.Username, &info.Password)
	if err == sql.ErrNoRows {
		// 如果没有设置，创建默认管理员
		defaultUser := "admin"
		defaultPass := generateRandomString(16)
		if err := d.SetAdminInfo(defaultUser, defaultPass); err != nil {
			return nil, err
		}
		return &AdminInfo{Username: defaultUser, Password: defaultPass}, nil
	}
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (d *Database) SetAdminInfo(username, password string) error {
	// 如果密码不是哈希格式，则进行哈希
	hashedPassword := password
	if !crypto.IsHashed(password) {
		hashed, err := crypto.HashPassword(password)
		if err != nil {
			return err
		}
		hashedPassword = hashed
	}

	_, err := d.db.Exec(
		"INSERT OR REPLACE INTO admin (id, username, password) VALUES (1, ?, ?)",
		username, hashedPassword)
	return err
}

type Binding struct {
	ID        int
	Path      string
	Port      int
	Password  string
	CreatedAt time.Time
}

func (d *Database) GetBindings() ([]*Binding, error) {
	rows, err := d.db.Query("SELECT id, path, port, password, created_at FROM bindings ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bindings []*Binding
	for rows.Next() {
		var b Binding
		var createdAt string
		if err := rows.Scan(&b.ID, &b.Path, &b.Port, &b.Password, &createdAt); err != nil {
			return nil, err
		}
		if t, err := time.Parse("2006-01-02 15:04:05", createdAt); err == nil {
			b.CreatedAt = t
		}
		bindings = append(bindings, &b)
	}
	return bindings, rows.Err()
}

func (d *Database) AddBinding(path string, port int, password string) error {
	// 如果密码不是哈希格式，则进行哈希
	hashedPassword := password
	if password != "" && !crypto.IsHashed(password) {
		hashed, err := crypto.HashPassword(password)
		if err != nil {
			return err
		}
		hashedPassword = hashed
	}

	_, err := d.db.Exec(
		"INSERT INTO bindings (path, port, password) VALUES (?, ?, ?)",
		path, port, hashedPassword)
	return err
}

func (d *Database) DeleteBinding(id int) error {
	_, err := d.db.Exec("DELETE FROM bindings WHERE id = ?", id)
	return err
}

func (d *Database) GetBindingByPath(path string) (*Binding, error) {
	var b Binding
	var createdAt string
	err := d.db.QueryRow(
		"SELECT id, path, port, password, created_at FROM bindings WHERE path = ?",
		path).Scan(&b.ID, &b.Path, &b.Port, &b.Password, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if t, err := time.Parse("2006-01-02 15:04:05", createdAt); err == nil {
		b.CreatedAt = t
	}
	return &b, nil
}

func (d *Database) GetBindingByPort(port int) (*Binding, error) {
	var b Binding
	var createdAt string
	err := d.db.QueryRow(
		"SELECT id, path, port, password, created_at FROM bindings WHERE port = ?",
		port).Scan(&b.ID, &b.Path, &b.Port, &b.Password, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if t, err := time.Parse("2006-01-02 15:04:05", createdAt); err == nil {
		b.CreatedAt = t
	}
	return &b, nil
}

type ServerInfo struct {
	ServerURL string
	APIKey    string
}

func (d *Database) GetServerInfo() (*ServerInfo, error) {
	var info ServerInfo
	err := d.db.QueryRow("SELECT server_url, api_key FROM server_info LIMIT 1").Scan(
		&info.ServerURL, &info.APIKey)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (d *Database) SetServerInfo(serverURL, apiKey string) error {
	_, err := d.db.Exec(
		"INSERT OR REPLACE INTO server_info (id, server_url, api_key) VALUES (1, ?, ?)",
		serverURL, apiKey)
	return err
}

