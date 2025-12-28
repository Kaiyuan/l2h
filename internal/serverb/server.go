package serverb

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
	port   int
	db     *Database
	webrtc *webrtc.Manager
}

func NewServer(port int, dbPath string) *Server {
	db, err := NewDatabase(dbPath)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	return &Server{
		port:   port,
		db:     db,
		webrtc: webrtc.NewManager(),
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/api/", s.handleAPI)

	log.Printf("服务器B启动在端口 %d", s.port)
	return http.ListenAndServe(":"+strconv.Itoa(s.port), mux)
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "admin"
	}

	// 检查是否是管理页面
	if path == "admin" {
		s.serveAdminPage(w, r)
		return
	}

	// 处理 WebRTC 连接请求
	s.handleWebRTCRequest(w, r, path)
}

func (s *Server) handleAPI(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/")

	switch {
	case path == "bindings" && r.Method == "GET":
		s.handleGetBindings(w, r)
	case path == "bindings" && r.Method == "POST":
		s.handleAddBinding(w, r)
	case strings.HasPrefix(path, "bindings/") && r.Method == "DELETE":
		s.handleDeleteBinding(w, r)
	case path == "webrtc/answer" && r.Method == "POST":
		s.handleWebRTCAnswer(w, r)
	default:
		writeError(w, http.StatusNotFound, "Not found")
	}
}

func (s *Server) handleGetBindings(w http.ResponseWriter, r *http.Request) {
	bindings, err := s.db.GetBindings()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, bindings)
}

func (s *Server) handleAddBinding(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path     string `json:"path"`
		Port     int    `json:"port"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.db.AddBinding(req.Path, req.Port, req.Password); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleDeleteBinding(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/bindings/")
	var id int
	if _, err := fmt.Sscanf(path, "%d", &id); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := s.db.DeleteBinding(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleWebRTCAnswer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Answer string `json:"answer"`
		Path   string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 处理 WebRTC answer
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) serveAdminPage(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
	<title>L2H 服务端管理</title>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<body>
	<h1>L2H 服务端管理</h1>
	<div id="app">
		<h2>路径绑定管理</h2>
		<div id="bindings"></div>
		<button onclick="loadBindings()">刷新</button>
		<button onclick="showAddForm()">添加绑定</button>
	</div>
	<script>
		async function loadBindings() {
			const response = await fetch('/api/bindings');
			const bindings = await response.json();
			const div = document.getElementById('bindings');
			div.innerHTML = '<table border="1"><tr><th>ID</th><th>路径</th><th>端口</th><th>操作</th></tr>' +
				bindings.map(b => '<tr><td>' + b.id + '</td><td>' + b.path + '</td><td>' + b.port + '</td><td><button onclick="deleteBinding(' + b.id + ')">删除</button></td></tr>').join('') +
				'</table>';
		}
		async function deleteBinding(id) {
			if (confirm('确定要删除吗？')) {
				await fetch('/api/bindings/' + id, {method: 'DELETE'});
				loadBindings();
			}
		}
		function showAddForm() {
			const path = prompt('请输入路径:');
			const port = prompt('请输入端口:');
			const password = prompt('请输入密码（可选，直接回车跳过）:');
			if (path && port) {
				fetch('/api/bindings', {
					method: 'POST',
					headers: {'Content-Type': 'application/json'},
					body: JSON.stringify({path: path, port: parseInt(port), password: password || ''})
				}).then(() => loadBindings());
			}
		}
		loadBindings();
	</script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func (s *Server) handleWebRTCRequest(w http.ResponseWriter, r *http.Request, path string) {
	// 查找对应的绑定
	binding, err := s.db.GetBindingByPath(path)
	if err != nil || binding == nil {
		http.NotFound(w, r)
		return
	}

	// 这里应该实现 WebRTC 连接逻辑，将请求转发到本地端口
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<title>WebRTC 代理</title>
	<meta charset="utf-8">
</head>
<body>
	<h1>WebRTC 代理到端口 %d</h1>
	<div id="status">连接中...</div>
	<script>
		// WebRTC 连接逻辑，连接到本地端口 %d
	</script>
</body>
</html>`, binding.Port, binding.Port)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

