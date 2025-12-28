package servera

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"l2h/internal/webrtc"
)

type Server struct {
	port     int
	db       *Database
	webrtc   *webrtc.Manager
	configFile string
}

func NewServer(port int, dbPath string, configFile string) *Server {
	db, err := NewDatabase(dbPath)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	return &Server{
		port:     port,
		db:       db,
		webrtc:   webrtc.NewManager(),
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
		if path == settings.AdminPath {
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
			if err != nil || cookie.Value != dbPath.Password {
				// 显示密码输入页面
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
		writeError(w, http.StatusNotFound, "Not found")
	}
}

func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := s.db.GetSettings()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (s *Server) handleSetSettings(w http.ResponseWriter, r *http.Request) {
	var settings Settings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.db.SetSettings(&settings); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleGetPaths(w http.ResponseWriter, r *http.Request) {
	paths, err := s.db.GetPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, paths)
}

func (s *Server) handleAddPath(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path        string `json:"path"`
		Password    string `json:"password"`
		ServerBPort int    `json:"server_b_port"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.db.AddPath(req.Path, req.Password, req.ServerBPort); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleDeletePath(w http.ResponseWriter, r *http.Request) {
	// 从路径中提取ID
	path := strings.TrimPrefix(r.URL.Path, "/api/paths/")
	var id int
	if _, err := fmt.Sscanf(path, "%d", &id); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := s.db.DeletePath(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleGetAPIKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := s.db.GetAPIKeys()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, keys)
}

func (s *Server) handleGenerateAPIKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	key, err := s.db.GenerateAPIKey(req.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"key": key})
}

func (s *Server) handleDeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/api-keys/")
	var id int
	if _, err := fmt.Sscanf(path, "%d", &id); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := s.db.DeleteAPIKey(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleWebRTCOffer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Offer string `json:"offer"`
		Path  string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 处理 WebRTC offer
	answer, err := s.webrtc.HandleOffer(req.Offer, req.Path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"answer": answer})
}

func (s *Server) handleAuth(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path     string `json:"path"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	dbPath, err := s.db.GetPathByPath(req.Path)
	if err != nil || dbPath == nil {
		writeError(w, http.StatusNotFound, "Path not found")
		return
	}

	if dbPath.Password != req.Password {
		writeError(w, http.StatusUnauthorized, "Invalid password")
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

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
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
	// 这里应该返回使用 PrimeVue V4 的管理界面
	// 为了简化，先返回一个基础的管理页面
	html := `<!DOCTYPE html>
<html>
<head>
	<title>L2H 管理后台</title>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<script src="https://cdn.jsdelivr.net/npm/vue@3/dist/vue.global.js"></script>
	<script src="https://unpkg.com/primevue@^4/core/core.min.js"></script>
	<link rel="stylesheet" href="https://unpkg.com/primevue@^4/themes/aura-light-blue/theme.css" />
</head>
<body>
	<div id="app">
		<h1>L2H 管理后台</h1>
		<!-- 这里应该使用 PrimeVue 组件构建完整的管理界面 -->
	</div>
	<script>
		const { createApp } = Vue;
		createApp({
			data() {
				return {
					settings: {},
					paths: [],
					apiKeys: []
				}
			},
			mounted() {
				this.loadData();
			},
			methods: {
				async loadData() {
					// 加载设置、路径、API密钥等数据
				}
			}
		}).mount('#app');
	</script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
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

