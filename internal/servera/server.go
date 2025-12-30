package servera

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"

	"l2h/internal/crypto"
	"l2h/internal/utils"
	"l2h/internal/webrtc"
)

//go:embed all:static
var adminFS embed.FS

type Server struct {
	port       int
	db         *Database
	webrtc     *webrtc.Manager
	configFile string
}

func NewServer(port int, dbPath string, configFile string) *Server {
	db, err := NewDatabase(dbPath)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	return &Server{
		port:       port,
		db:         db,
		webrtc:     webrtc.NewManager(),
		configFile: configFile,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// 静态文件服务
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/api/", s.handleAPI)

	log.Printf("服务器A启动在端口 %d", s.port)
	return http.ListenAndServe(":"+strconv.Itoa(s.port), mux)
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "index"
	}

	// 检查是否是管理路径
	settings, err := s.db.GetSettings()
	if err == nil && settings != nil {
		// 精确匹配 adminPath，重定向到 adminPath/
		if path == settings.AdminPath {
			http.Redirect(w, r, "/"+path+"/", http.StatusMovedPermanently)
			return
		}

		// 匹配 adminPath/ 前缀
		if strings.HasPrefix(path, settings.AdminPath+"/") {
			s.serveAdminPage(w, r)
			return
		}
	}

	// 检查是否是配置的路径
	dbPath, err := s.db.GetPathByPath(path)
	if err == nil && dbPath != nil {
		// 检查是否需要密码
		if dbPath.Password != "" {
			// 检查是否已认证
			cookie, err := r.Cookie("l2h_auth_" + path)
			if err != nil || cookie.Value == "" {
				// 显示密码输入页面
				s.servePasswordPage(w, r, path)
				return
			}
			// 验证cookie中的密码（支持哈希和明文）
			valid := false
			if crypto.IsHashed(dbPath.Password) {
				valid, _ = crypto.VerifyPassword(cookie.Value, dbPath.Password)
			} else {
				// 向后兼容
				valid = cookie.Value == dbPath.Password
			}
			if !valid {
				s.servePasswordPage(w, r, path)
				return
			}
		}

		// 通过 WebRTC 连接到服务器B
		s.handleWebRTCPath(w, r, path, dbPath.ServerBPort)
		return
	}

	// 默认首页
	s.serveIndexPage(w, r)
}

func (s *Server) handleAPI(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/")

	switch {
	case path == "settings" && r.Method == "GET":
		s.handleGetSettings(w, r)
	case path == "settings" && r.Method == "POST":
		s.handleSetSettings(w, r)
	case path == "paths" && r.Method == "GET":
		s.handleGetPaths(w, r)
	case path == "paths" && r.Method == "POST":
		s.handleAddPath(w, r)
	case strings.HasPrefix(path, "paths/") && r.Method == "DELETE":
		s.handleDeletePath(w, r)
	case path == "api-keys" && r.Method == "GET":
		s.handleGetAPIKeys(w, r)
	case path == "api-keys" && r.Method == "POST":
		s.handleGenerateAPIKey(w, r)
	case strings.HasPrefix(path, "api-keys/") && r.Method == "DELETE":
		s.handleDeleteAPIKey(w, r)
	case path == "webrtc/offer" && r.Method == "POST":
		s.handleWebRTCOffer(w, r)
	case path == "auth" && r.Method == "POST":
		s.handleAuth(w, r)
	default:
		utils.WriteError(w, http.StatusNotFound, "Not found")
	}
}

func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := s.db.GetSettings()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusOK, settings)
}

func (s *Server) handleSetSettings(w http.ResponseWriter, r *http.Request) {
	var settings Settings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.db.SetSettings(&settings); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleGetPaths(w http.ResponseWriter, r *http.Request) {
	paths, err := s.db.GetPaths()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusOK, paths)
}

func (s *Server) handleAddPath(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path        string `json:"path"`
		Password    string `json:"password"`
		ServerBPort int    `json:"server_b_port"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.db.AddPath(req.Path, req.Password, req.ServerBPort); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleDeletePath(w http.ResponseWriter, r *http.Request) {
	// 从路径中提取ID
	path := strings.TrimPrefix(r.URL.Path, "/api/paths/")
	var id int
	if _, err := fmt.Sscanf(path, "%d", &id); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := s.db.DeletePath(id); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleGetAPIKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := s.db.GetAPIKeys()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJSON(w, http.StatusOK, keys)
}

func (s *Server) handleGenerateAPIKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name          string `json:"name"`
		ExpiresInDays int    `json:"expires_in_days"` // 0 表示永不过期
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	key, err := s.db.GenerateAPIKey(req.Name, req.ExpiresInDays)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"key": key})
}

func (s *Server) handleDeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/api-keys/")
	var id int
	if _, err := fmt.Sscanf(path, "%d", &id); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := s.db.DeleteAPIKey(id); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleWebRTCOffer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Offer string `json:"offer"`
		Path  string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 处理 WebRTC offer
	answer, err := s.webrtc.HandleOffer(req.Offer, req.Path)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"answer": answer})
}

func (s *Server) handleAuth(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path     string `json:"path"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	dbPath, err := s.db.GetPathByPath(req.Path)
	if err != nil || dbPath == nil {
		utils.WriteError(w, http.StatusNotFound, "Path not found")
		return
	}

	// 验证密码（支持哈希和明文，用于向后兼容）
	valid := false
	if crypto.IsHashed(dbPath.Password) {
		// 使用哈希验证
		valid, _ = crypto.VerifyPassword(req.Password, dbPath.Password)
	} else {
		// 向后兼容：如果是明文，直接比较
		valid = dbPath.Password == req.Password
		// 如果匹配，更新为哈希格式
		if valid {
			hashed, err := crypto.HashPassword(req.Password)
			if err == nil {
				// 更新数据库中的密码为哈希格式
				s.db.UpdatePathPassword(dbPath.ID, hashed)
			}
		}
	}

	if !valid {
		utils.WriteError(w, http.StatusUnauthorized, "Invalid password")
		return
	}

	// 设置认证cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "l2h_auth_" + req.Path,
		Value:    req.Password,
		Path:     "/",
		MaxAge:   86400 * 7, // 7天
		HttpOnly: true,
	})

	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) serveIndexPage(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
	<title>L2H - WebRTC Proxy</title>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<body>
	<h1>L2H WebRTC Proxy</h1>
	<p>欢迎使用 L2H WebRTC 代理服务</p>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (s *Server) serveAdminPage(w http.ResponseWriter, r *http.Request) {
	// 获取 dist 子目录
	distFS, err := fs.Sub(adminFS, "static")
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to load admin assets")
		return
	}

	// 处理请求路径
	// r.URL.Path 是完整的路径，例如 /admin/assets/main.css
	// 我们需要去掉 /admin 前缀（假设 adminPath 是 admin）
	// 但是这里我们是在 handleRoot 中调用的，handleRoot 已经解析了 path

	// 为了简单起见，我们直接看 accept header 或者 path 后缀
	// 如果请求的是静态资源 (assets/*)，直接服务
	// 否则返回 index.html (SPA)

	settings, _ := s.db.GetSettings()
	adminPath := "admin"
	if settings != nil {
		adminPath = settings.AdminPath
	}

	// 构建相对于 adminPath 的路径
	// r.URL.Path: /admin/login -> rel: /login
	// r.URL.Path: /admin/assets/main.css -> rel: /assets/main.css

	// 注意：handleRoot 里面 path = strings.TrimPrefix(r.URL.Path, "/")
	// 且 logic 是: if path == settings.AdminPath
	// 这意味着只有精确匹配 /admin 时才会调用 serveAdminPage
	// 这对于 SPA 是不够的，因为 SPA 有子路由 /admin/login, /admin/assets/...

	// 目前 handleRoot 的逻辑：
	// path := strings.TrimPrefix(r.URL.Path, "/")
	// if path == settings.AdminPath { serveAdminPage }

	// 这意味着 /admin/foo 不会匹配！
	// 后端的 handleRoot 需要修改以支持 /admin/* 前缀匹配。

	// 鉴于 handleRoot 已经很大，我们先修改 serveAdminPage，稍后修改 handleRoot 或者路由逻辑。
	// 这里假设 serveAdminPage 接收到的 r 是针对 /admin/* 的请求。

	fpath := r.URL.Path
	// 移除 /admin 前缀
	if strings.HasPrefix(fpath, "/"+adminPath) {
		fpath = strings.TrimPrefix(fpath, "/"+adminPath)
	}

	// 如果是空或 /，由 index.html 处理
	if fpath == "" || fpath == "/" {
		fpath = "index.html"
	}

	// 尝试打开文件
	f, err := distFS.Open(strings.TrimPrefix(fpath, "/"))
	if err != nil {
		// 文件不存在，如果是 assets，返回 404
		if strings.HasPrefix(fpath, "/assets/") {
			http.NotFound(w, r)
			return
		}
		// 否则返回 index.html (SPA history mode)
		fpath = "index.html"
		f, err = distFS.Open("index.html")
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Admin interface not found")
			return
		}
	}
	defer f.Close()

	// 特殊处理 index.html，注入 L2H_ADMIN_BASE
	if fpath == "index.html" {
		content, err := io.ReadAll(f)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, "Failed to read admin index")
			return
		}

		// 动态注入 base path 脚本
		script := fmt.Sprintf("<script>window.L2H_ADMIN_BASE = '/%s/';</script>", adminPath)
		html := strings.Replace(string(content), "<head>", "<head>"+script, 1)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Write([]byte(html))
		return
	}

	// 获取文件信息以设置 Content-Type
	stat, _ := f.Stat()

	// 使用 http.ServeContent 服务文件
	http.ServeContent(w, r, fpath, stat.ModTime(), f.(io.ReadSeeker))
}

func (s *Server) servePasswordPage(w http.ResponseWriter, r *http.Request, path string) {
	html := `<!DOCTYPE html>
<html>
<head>
	<title>需要密码</title>
	<meta charset="utf-8">
</head>
<body>
	<h1>此路径需要密码</h1>
	<form id="authForm">
		<input type="password" id="password" placeholder="请输入密码" required>
		<button type="submit">提交</button>
	</form>
	<script>
		document.getElementById('authForm').addEventListener('submit', async (e) => {
			e.preventDefault();
			const password = document.getElementById('password').value;
			const response = await fetch('/api/auth', {
				method: 'POST',
				headers: {'Content-Type': 'application/json'},
				body: JSON.stringify({path: '` + path + `', password: password})
			});
			if (response.ok) {
				window.location.reload();
			} else {
				alert('密码错误');
			}
		});
	</script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (s *Server) handleWebRTCPath(w http.ResponseWriter, r *http.Request, path string, port int) {
	// 这里应该实现 WebRTC 连接逻辑
	// 暂时返回一个简单的页面
	html := `<!DOCTYPE html>
<html>
<head>
	<title>WebRTC 连接</title>
	<meta charset="utf-8">
</head>
<body>
	<h1>正在连接到服务器B的端口 ` + strconv.Itoa(port) + `</h1>
	<div id="status">初始化中...</div>
	<script>
		// WebRTC 连接逻辑
	</script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}
