package servera

import (
	"database/sql"
	"time"

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
		`CREATE TABLE IF NOT EXISTS settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			admin_path TEXT UNIQUE NOT NULL,
			username TEXT NOT NULL,
			password TEXT NOT NULL,
			email TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS paths (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			path TEXT UNIQUE NOT NULL,
			password TEXT,
			server_b_port INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS api_keys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT UNIQUE NOT NULL,
			name TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
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

type Settings struct {
	AdminPath string
	Username  string
	Password  string
	Email     string
}

func (d *Database) GetSettings() (*Settings, error) {
	var s Settings
	err := d.db.QueryRow("SELECT admin_path, username, password, email FROM settings LIMIT 1").Scan(
		&s.AdminPath, &s.Username, &s.Password, &s.Email)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (d *Database) SetSettings(s *Settings) error {
	_, err := d.db.Exec(
		"INSERT OR REPLACE INTO settings (id, admin_path, username, password, email) VALUES (1, ?, ?, ?, ?)",
		s.AdminPath, s.Username, s.Password, s.Email)
	return err
}

type Path struct {
	ID          int
	Path        string
	Password    string
	ServerBPort int
	CreatedAt   time.Time
}

func (d *Database) GetPaths() ([]*Path, error) {
	rows, err := d.db.Query("SELECT id, path, password, server_b_port, created_at FROM paths ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []*Path
	for rows.Next() {
		var p Path
		var createdAt string
		if err := rows.Scan(&p.ID, &p.Path, &p.Password, &p.ServerBPort, &createdAt); err != nil {
			return nil, err
		}
		if t, err := time.Parse("2006-01-02 15:04:05", createdAt); err == nil {
			p.CreatedAt = t
		}
		paths = append(paths, &p)
	}
	return paths, rows.Err()
}

func (d *Database) AddPath(path string, password string, serverBPort int) error {
	_, err := d.db.Exec(
		"INSERT INTO paths (path, password, server_b_port) VALUES (?, ?, ?)",
		path, password, serverBPort)
	return err
}

func (d *Database) DeletePath(id int) error {
	_, err := d.db.Exec("DELETE FROM paths WHERE id = ?", id)
	return err
}

func (d *Database) GetPathByPath(path string) (*Path, error) {
	var p Path
	var createdAt string
	err := d.db.QueryRow(
		"SELECT id, path, password, server_b_port, created_at FROM paths WHERE path = ?",
		path).Scan(&p.ID, &p.Path, &p.Password, &p.ServerBPort, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if t, err := time.Parse("2006-01-02 15:04:05", createdAt); err == nil {
		p.CreatedAt = t
	}
	return &p, nil
}

type APIKey struct {
	ID        int
	Key       string
	Name      string
	CreatedAt time.Time
}

func (d *Database) GenerateAPIKey(name string) (string, error) {
	key := generateRandomString(32)
	_, err := d.db.Exec("INSERT INTO api_keys (key, name) VALUES (?, ?)", key, name)
	if err != nil {
		return "", err
	}
	return key, nil
}

func (d *Database) GetAPIKeys() ([]*APIKey, error) {
	rows, err := d.db.Query("SELECT id, key, name, created_at FROM api_keys ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*APIKey
	for rows.Next() {
		var k APIKey
		var createdAt string
		if err := rows.Scan(&k.ID, &k.Key, &k.Name, &createdAt); err != nil {
			return nil, err
		}
		if t, err := time.Parse("2006-01-02 15:04:05", createdAt); err == nil {
			k.CreatedAt = t
		}
		keys = append(keys, &k)
	}
	return keys, rows.Err()
}

func (d *Database) ValidateAPIKey(key string) (bool, error) {
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM api_keys WHERE key = ?", key).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (d *Database) DeleteAPIKey(id int) error {
	_, err := d.db.Exec("DELETE FROM api_keys WHERE id = ?", id)
	return err
}

