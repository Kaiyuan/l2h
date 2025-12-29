// Package servera 提供服务器A的数据库操作功能
package servera

import (
	"database/sql"
	"time"

	"l2h/internal/crypto"

	_ "github.com/mattn/go-sqlite3"
)

// Database 数据库结构体
type Database struct {
	db *sql.DB
}

// NewDatabase 创建新的数据库实例
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

// initTables 初始化数据库表结构
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
			expires_at DATETIME,
			last_used_at DATETIME,
			usage_count INTEGER DEFAULT 0,
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

// Close 关闭数据库连接
func (d *Database) Close() error {
	return d.db.Close()
}

// Settings 设置结构体
type Settings struct {
	AdminPath string `json:"admin_path"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Email     string `json:"email"`
}

// GetSettings 获取系统设置
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

// SetSettings 设置系统设置
func (d *Database) SetSettings(s *Settings) error {
	password := s.Password
	if !crypto.IsHashed(password) {
		hashed, err := crypto.HashPassword(password)
		if err != nil {
			return err
		}
		password = hashed
	}

	_, err := d.db.Exec(
		"INSERT OR REPLACE INTO settings (id, admin_path, username, password, email) VALUES (1, ?, ?, ?, ?)",
		s.AdminPath, s.Username, password, s.Email)
	return err
}

// Path 路径结构体
type Path struct {
	ID          int       `json:"id"`
	Path        string    `json:"path"`
	Password    string    `json:"password"`
	ServerBPort int       `json:"server_b_port"`
	CreatedAt   time.Time `json:"created_at"`
}

// GetPaths 获取所有路径配置
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

// AddPath 添加新的路径配置
func (d *Database) AddPath(path string, password string, serverBPort int) error {
	hashedPassword := password
	if password != "" && !crypto.IsHashed(password) {
		hashed, err := crypto.HashPassword(password)
		if err != nil {
			return err
		}
		hashedPassword = hashed
	}

	_, err := d.db.Exec(
		"INSERT INTO paths (path, password, server_b_port) VALUES (?, ?, ?)",
		path, hashedPassword, serverBPort)
	return err
}

// DeletePath 删除路径配置
func (d *Database) DeletePath(id int) error {
	_, err := d.db.Exec("DELETE FROM paths WHERE id = ?", id)
	return err
}

// UpdatePathPassword 更新路径密码
func (d *Database) UpdatePathPassword(id int, password string) error {
	_, err := d.db.Exec("UPDATE paths SET password = ? WHERE id = ?", password, id)
	return err
}

// GetPathByPath 根据路径字符串获取路径配置
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

// APIKey API Key 结构体
type APIKey struct {
	ID         int        `json:"id"`
	Key        string     `json:"key"`
	Name       string     `json:"name"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	UsageCount int        `json:"usage_count"`
	CreatedAt  time.Time  `json:"created_at"`
}

// GenerateAPIKey 生成新的 API Key
func (d *Database) GenerateAPIKey(name string, expiresInDays int) (string, error) {
	key := generateRandomString(32)

	var expiresAt interface{}
	if expiresInDays > 0 {
		expiresAt = time.Now().AddDate(0, 0, expiresInDays).Format("2006-01-02 15:04:05")
	}

	_, err := d.db.Exec("INSERT INTO api_keys (key, name, expires_at) VALUES (?, ?, ?)", key, name, expiresAt)
	if err != nil {
		return "", err
	}
	return key, nil
}

// GetAPIKeys 获取所有 API Key
func (d *Database) GetAPIKeys() ([]*APIKey, error) {
	rows, err := d.db.Query("SELECT id, key, name, expires_at, last_used_at, usage_count, created_at FROM api_keys ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*APIKey
	for rows.Next() {
		var k APIKey
		var createdAt, expiresAt, lastUsedAt sql.NullString
		if err := rows.Scan(&k.ID, &k.Key, &k.Name, &expiresAt, &lastUsedAt, &k.UsageCount, &createdAt); err != nil {
			return nil, err
		}
		if t, err := time.Parse("2006-01-02 15:04:05", createdAt.String); err == nil {
			k.CreatedAt = t
		}
		if expiresAt.Valid {
			if t, err := time.Parse("2006-01-02 15:04:05", expiresAt.String); err == nil {
				k.ExpiresAt = &t
			}
		}
		if lastUsedAt.Valid {
			if t, err := time.Parse("2006-01-02 15:04:05", lastUsedAt.String); err == nil {
				k.LastUsedAt = &t
			}
		}
		keys = append(keys, &k)
	}
	return keys, rows.Err()
}

// ValidateAPIKey 验证 API Key 的有效性
func (d *Database) ValidateAPIKey(key string) (bool, error) {
	var id, usageCount int
	var expiresAt sql.NullString

	err := d.db.QueryRow("SELECT id, expires_at, usage_count FROM api_keys WHERE key = ?", key).Scan(&id, &expiresAt, &usageCount)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if expiresAt.Valid {
		expires, err := time.Parse("2006-01-02 15:04:05", expiresAt.String)
		if err == nil && time.Now().After(expires) {
			return false, nil
		}
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = d.db.Exec("UPDATE api_keys SET last_used_at = ?, usage_count = usage_count + 1 WHERE id = ?", now, id)
	if err != nil {
		// 即使更新失败，也返回验证成功（因为key是有效的）
	}

	return true, nil
}

// DeleteAPIKey 删除 API Key
func (d *Database) DeleteAPIKey(id int) error {
	_, err := d.db.Exec("DELETE FROM api_keys WHERE id = ?", id)
	return err
}
