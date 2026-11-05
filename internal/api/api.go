// Package api expõe a API HTTP do Taskflow e serve o front estático.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"taskflow/internal/model"
	"taskflow/internal/store"
)

const maxBodySize = 1 << 20 // 1mb, igual ao limite antigo do express

// Server conecta as rotas HTTP ao armazenamento do Taskflow.
type Server struct {
	store      *store.Store
	staticRoot string
	logger     *slog.Logger
}

// New monta o servidor com o armazenamento e a pasta de arquivos estáticos.
func New(dataStore *store.Store, staticRoot string, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	return &Server{store: dataStore, staticRoot: staticRoot, logger: logger}
}

// Handler devolve o roteador completo: API em /api e SPA no restante.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", s.handleHealth)

	mux.HandleFunc("POST /api/auth/register", s.handleRegister)
	mux.HandleFunc("POST /api/auth/login", s.handleLogin)
	mux.HandleFunc("POST /api/auth/logout", s.handleLogout)
	mux.HandleFunc("GET /api/auth/me", s.handleMe)
	mux.HandleFunc("POST /api/register", s.handleRegister)
	mux.HandleFunc("POST /api/login", s.handleLogin)
	mux.HandleFunc("POST /api/logout", s.handleLogout)
	mux.HandleFunc("GET /api/me", s.handleMe)

	mux.HandleFunc("GET /api/bootstrap", s.handleBootstrap)

	mux.HandleFunc("GET /api/workspaces", s.handleListWorkspaces)
	mux.HandleFunc("POST /api/workspaces", s.handleCreateWorkspace)

	mux.HandleFunc("GET /api/projects", s.handleListProjects)
	mux.HandleFunc("POST /api/projects", s.handleCreateProject)
	mux.HandleFunc("GET /api/projects/{id}", s.handleGetProject)
	mux.HandleFunc("PUT /api/projects/{id}", s.handleUpdateProject)
	mux.HandleFunc("PATCH /api/projects/{id}", s.handleUpdateProject)
	mux.HandleFunc("DELETE /api/projects/{id}", s.handleDeleteProject)
	mux.HandleFunc("POST /api/projects/{id}/share", s.handleShareProject)

	mux.HandleFunc("GET /api/members", s.handleListMembers)
	mux.HandleFunc("GET /api/labels", s.handleListLabels)

	mux.HandleFunc("GET /api/columns", s.handleListColumns)

	mux.HandleFunc("GET /api/tasks", s.handleListTasks)
	mux.HandleFunc("POST /api/tasks", s.handleCreateTask)
	mux.HandleFunc("POST /api/tasks/reorder", s.handleReorderTasks)
	mux.HandleFunc("PUT /api/tasks/{id}", s.handleUpdateTask)
	mux.HandleFunc("PATCH /api/tasks/{id}", s.handleUpdateTask)
	mux.HandleFunc("DELETE /api/tasks/{id}", s.handleDeleteTask)

	mux.HandleFunc("GET /api/notes", s.handleListNotes)
	mux.HandleFunc("POST /api/notes", s.handleCreateNote)
	mux.HandleFunc("PUT /api/notes/{id}", s.handleUpdateNote)
	mux.HandleFunc("DELETE /api/notes/{id}", s.handleDeleteNote)
	mux.HandleFunc("POST /api/notes/{id}/share", s.handleShareNote)

	mux.Handle("GET /", s.staticHandler())

	return s.securityHeaders(s.recoverPanic(s.authMiddleware(mux)))
}

func (s *Server) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if problem := recover(); problem != nil {
				s.logger.Error("panic ao tratar requisição", "path", r.URL.Path, "erro", problem)
				writeError(w, http.StatusInternalServerError, "Erro interno do Taskflow.")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

type contextKey string

const userIDContextKey contextKey = "userID"

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}
		if isPublicAuthPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		if !sameOriginRequest(r) && isStateChanging(r.Method) {
			writeError(w, http.StatusForbidden, "Origem da requisição não permitida.")
			return
		}
		token := sessionToken(r)
		userID, ok := s.store.SessionUser(token)
		if !ok {
			writeError(w, http.StatusUnauthorized, "Não autenticado.")
			return
		}
		ctx := context.WithValue(r.Context(), userIDContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func isPublicAuthPath(path string) bool {
	return path == "/api/health" || path == "/api/auth/register" || path == "/api/auth/login" || path == "/api/auth/logout" || path == "/api/auth/me" || path == "/api/register" || path == "/api/login" || path == "/api/logout" || path == "/api/me"
}

func isStateChanging(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete
}

func sameOriginRequest(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	host := r.Host
	return strings.Contains(origin, host)
}

func sessionToken(r *http.Request) string {
	if cookie, err := r.Cookie("taskflow_session"); err == nil {
		return cookie.Value
	}
	header := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return strings.TrimSpace(header[7:])
	}
	return ""
}

func requestUserID(r *http.Request) string {
	if value := r.Context().Value(userIDContextKey); value != nil {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return ""
}

func (s *Server) setSessionCookie(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     "taskflow_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 3600,
	})
}

func (s *Server) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "taskflow_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// ---------- helpers de resposta ----------

type errorBody struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("falha ao serializar resposta", "erro", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorBody{Error: message})
}

// decodeBody lê e valida o JSON do corpo, respeitando o limite de tamanho.
func decodeBody(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodySize))
	if err := decoder.Decode(target); err != nil {
		var maxBytes *http.MaxBytesError
		if errors.As(err, &maxBytes) {
			writeError(w, http.StatusRequestEntityTooLarge, "Conteúdo enviado é muito grande.")
			return false
		}
		writeError(w, http.StatusBadRequest, "Corpo da requisição inválido.")
		return false
	}
	return true
}

// nullableString distingue "campo ausente" de "null explícito". Sem essa
// distinção não seria possível limpar um prazo ou uma data de conclusão.
type nullableString struct {
	Value   *string
	Present bool
}

func (n *nullableString) UnmarshalJSON(data []byte) error {
	n.Present = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		n.Value = nil
		return nil
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	n.Value = &value
	return nil
}

// ---------- payloads de entrada ----------

type workspaceRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type projectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Icon        *string `json:"icon"`
	Color       *string `json:"color"`
}

type taskRequest struct {
	Title       *string        `json:"title"`
	Description *string        `json:"description"`
	ProjectID   *string        `json:"projectId"`
	ColumnID    *string        `json:"columnId"`
	Order       *int           `json:"order"`
	AssigneeID  *string        `json:"assigneeId"`
	Priority    *string        `json:"priority"`
	DueDate     nullableString `json:"dueDate"`
	Labels      *[]string      `json:"labels"`
	CreatedAt   *string        `json:"createdAt"`
	CompletedAt nullableString `json:"completedAt"`
}

type reorderRequest struct {
	TaskID   *string `json:"taskId"`
	ColumnID *string `json:"columnId"`
	Index    *int    `json:"index"`
}

type noteRequest struct {
	Title   *string `json:"title"`
	Content *string `json:"content"`
	Color   *string `json:"color"`
	Pinned  *bool   `json:"pinned"`
}

// ---------- health e bootstrap ----------

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "service": "taskflow"})
}

type authRequest struct {
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

type authBucket struct {
	mu     sync.Mutex
	count  int
	reset  time.Time
}

var authLimits sync.Map

func allowAuthRequest(ip string) bool {
	now := time.Now()
	value, _ := authLimits.LoadOrStore(ip, &authBucket{reset: now.Add(time.Minute)})
	bucket := value.(*authBucket)
	bucket.mu.Lock()
	defer bucket.mu.Unlock()
	if now.After(bucket.reset) {
		bucket.count = 0
		bucket.reset = now.Add(time.Minute)
	}
	if bucket.count >= 100 {
		return false
	}
	bucket.count++
	return true
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	host := r.RemoteAddr
	if index := strings.LastIndex(host, ":"); index >= 0 {
		host = host[:index]
	}
	return host
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if !allowAuthRequest(clientIP(r)) {
		writeError(w, http.StatusTooManyRequests, "Muitas tentativas. Tente novamente mais tarde.")
		return
	}
	var body authRequest
	if !decodeBody(w, r, &body) {
		return
	}
	name := model.Text(body.Name, "")
	email := strings.ToLower(strings.TrimSpace(model.Text(body.Email, "")))
	password := model.Text(body.Password, "")
	if name == "" || email == "" || password == "" {
		writeError(w, http.StatusBadRequest, "Nome, e-mail e senha são obrigatórios.")
		return
	}
	if !strings.Contains(email, "@") {
		writeError(w, http.StatusBadRequest, "E-mail inválido.")
		return
	}
	if len(password) < 8 {
		writeError(w, http.StatusBadRequest, "A senha deve ter ao menos 8 caracteres.")
		return
	}
	hash, err := model.HashPassword(password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Não foi possível criar a conta.")
		return
	}
	var created model.User
	err = s.store.Update(func(data *model.Data) {
		for _, user := range data.Users {
			if strings.EqualFold(user.Email, email) {
				return
			}
		}
		created = model.User{
			ID:           model.UID(),
			Name:         name,
			Email:        email,
			Initials:     initialsFor(name),
			Color:        "#3f6152",
			Role:         "MEMBER",
			PasswordHash: hash,
		}
		data.Users = append(data.Users, created)
		seedUserData(data, created)
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	if created.ID == "" {
		writeError(w, http.StatusConflict, "E-mail já cadastrado.")
		return
	}
	token, err := s.store.CreateSession(created.ID)
	if err != nil {
		s.fail(w, err)
		return
	}
	secure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
	s.setSessionCookie(w, token, secure)
	writeJSON(w, http.StatusCreated, map[string]any{"user": publicUser(created), "token": token})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if !allowAuthRequest(clientIP(r)) {
		writeError(w, http.StatusTooManyRequests, "Muitas tentativas. Tente novamente mais tarde.")
		return
	}
	var body authRequest
	if !decodeBody(w, r, &body) {
		return
	}
	email := strings.ToLower(strings.TrimSpace(model.Text(body.Email, "")))
	password := model.Text(body.Password, "")
	if email == "" || password == "" {
		writeError(w, http.StatusBadRequest, "E-mail e senha são obrigatórios.")
		return
	}
	data := s.store.Snapshot()
	var found *model.User
	for index := range data.Users {
		if strings.EqualFold(data.Users[index].Email, email) {
			found = &data.Users[index]
			break
		}
	}
	if found == nil || !model.VerifyPassword(found.PasswordHash, password) {
		writeError(w, http.StatusUnauthorized, "Credenciais inválidas.")
		return
	}
	token, err := s.store.CreateSession(found.ID)
	if err != nil {
		s.fail(w, err)
		return
	}
	secure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
	s.setSessionCookie(w, token, secure)
	writeJSON(w, http.StatusOK, map[string]any{"user": publicUser(*found), "token": token})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	s.store.DeleteSession(sessionToken(r))
	s.clearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	data := s.store.Snapshot()
	user, ok := findUserByID(&data, requestUserID(r))
	if !ok {
		writeError(w, http.StatusUnauthorized, "Não autenticado.")
		return
	}
	writeJSON(w, http.StatusOK, publicUser(user))
}

func (s *Server) handleBootstrap(w http.ResponseWriter, r *http.Request) {
	userID := requestUserID(r)
	data := s.store.Snapshot()
	user := findUser(&data, userID)
	workspaces := workspacesOf(data, userID)
	projects := accessibleProjects(data, userID)
	columns := columnsOfUser(data, userID)
	tasks := tasksOfUser(data, userID)
	notes := notesOfUser(data, userID)
	labels := labelsOfUser(data, userID)
	workspace := model.Workspace{}
	if len(workspaces) > 0 {
		workspace = workspaces[0]
	}
	safeUser := publicUser(user)
	writeJSON(w, http.StatusOK, model.Bootstrap{
		CurrentUser: &safeUser,
		Workspace:   &workspace,
		Workspaces:  workspaces,
		Projects:    projects,
		Members:     []model.User{safeUser},
		Columns:     columns,
		Tasks:       tasks,
		Notes:       notes,
		Labels:      labels,
	})
}

// ---------- workspaces ----------

func (s *Server) handleListWorkspaces(w http.ResponseWriter, r *http.Request) {
	userID := requestUserID(r)
	data := s.store.Snapshot()
	writeJSON(w, http.StatusOK, workspacesOf(data, userID))
}

func (s *Server) handleCreateWorkspace(w http.ResponseWriter, r *http.Request) {
	var body workspaceRequest
	if !decodeBody(w, r, &body) {
		return
	}
	userID := requestUserID(r)
	var created model.Workspace
	err := s.store.Update(func(data *model.Data) {
		created = model.Workspace{
			ID:          model.UID(),
			Name:        model.Text(body.Name, "Novo workspace"),
			Description: model.Text(body.Description, ""),
			OwnerID:     userID,
			CreatedAt:   model.Now(),
		}
		data.Workspaces = append(data.Workspaces, created)
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// ---------- projetos ----------

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	userID := requestUserID(r)
	data := s.store.Snapshot()
	writeJSON(w, http.StatusOK, accessibleProjects(data, userID))
}

func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var body projectRequest
	if !decodeBody(w, r, &body) {
		return
	}
	userID := requestUserID(r)
	var created model.Project
	err := s.store.Update(func(data *model.Data) {
		workspaceID := ""
		workspaces := workspacesOf(*data, userID)
		if len(workspaces) > 0 {
			workspaceID = workspaces[0].ID
		} else {
			workspace := model.Workspace{
				ID:        model.UID(),
				Name:      "Meu workspace",
				OwnerID:   userID,
				CreatedAt: model.Now(),
			}
			data.Workspaces = append(data.Workspaces, workspace)
			workspaceID = workspace.ID
		}
		created = model.Project{
			ID:          model.UID(),
			WorkspaceID: workspaceID,
			Name:        model.Text(body.Name, "Novo projeto"),
			Description: model.Text(body.Description, ""),
			Icon:        model.Text(body.Icon, "✦"),
			Color:       model.Text(body.Color, "#3f6152"),
			OwnerID:     userID,
			CreatedAt:   model.Now(),
		}
		data.Projects = append(data.Projects, created)
		data.Columns = append(data.Columns, store.CreateColumns(created.ID, nil)...)
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) {
	userID := requestUserID(r)
	data := s.store.Snapshot()
	project, ok := findProject(&data, r.PathValue("id"))
	if !ok || project.OwnerID != userID {
		writeError(w, http.StatusNotFound, "Projeto não encontrado")
		return
	}
	writeJSON(w, http.StatusOK, model.ProjectDetail{
		Project: project,
		Columns: orderedColumns(columnsOf(&data, project.ID)),
		Tasks:   tasksOf(&data, project.ID),
	})
}

func (s *Server) handleUpdateProject(w http.ResponseWriter, r *http.Request) {
	var body projectRequest
	if !decodeBody(w, r, &body) {
		return
	}
	userID := requestUserID(r)
	var updated model.Project
	found := false
	err := s.store.Update(func(data *model.Data) {
		project, ok := findProject(data, r.PathValue("id"))
		if !ok || project.OwnerID != userID {
			return
		}
		project.Name = model.Text(body.Name, project.Name)
		project.Description = model.Text(body.Description, project.Description)
		project.Icon = model.Text(body.Icon, project.Icon)
		project.Color = model.Text(body.Color, project.Color)
		data.Projects[slices.IndexFunc(data.Projects, func(p model.Project) bool { return p.ID == project.ID })] = project
		updated = project
		found = true
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "Projeto não encontrado")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleShareProject(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email *string `json:"email"`
	}
	if !decodeBody(w, r, &body) {
		return
	}
	email := strings.ToLower(strings.TrimSpace(model.Text(body.Email, "")))
	if email == "" {
		writeError(w, http.StatusBadRequest, "E-mail obrigatório.")
		return
	}
	userID := requestUserID(r)
	var updated model.Project
	found := false
	err := s.store.Update(func(data *model.Data) {
		index := slices.IndexFunc(data.Projects, func(p model.Project) bool { return p.ID == r.PathValue("id") })
		if index < 0 || data.Projects[index].OwnerID != userID {
			return
		}
		targetID := ""
		for _, user := range data.Users {
			if strings.EqualFold(user.Email, email) {
				targetID = user.ID
				break
			}
		}
		if targetID == "" || targetID == userID {
			return
		}
		project := data.Projects[index]
		if !containsString(project.SharedWith, targetID) {
			project.SharedWith = append(project.SharedWith, targetID)
		}
		data.Projects[index] = project
		updated = project
		found = true
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	if !found {
		writeError(w, http.StatusBadRequest, "Usuário não encontrado ou projeto inválido.")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	userID := requestUserID(r)
	var (
		deleted bool
		lastOne bool
	)
	err := s.store.Update(func(data *model.Data) {
		index := slices.IndexFunc(data.Projects, func(p model.Project) bool { return p.ID == id && p.OwnerID == userID })
		if index < 0 {
			return
		}
		userProjects := projectsOf(*data, userID)
		if len(userProjects) == 1 {
			lastOne = true
			return
		}
		data.Projects = slices.Delete(data.Projects, index, index+1)
		data.Columns = slices.DeleteFunc(data.Columns, func(c model.Column) bool { return c.ProjectID == id })
		data.Tasks = slices.DeleteFunc(data.Tasks, func(t model.Task) bool { return t.ProjectID == id })
		deleted = true
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	switch {
	case deleted:
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	case lastOne:
		writeError(w, http.StatusBadRequest, "Não é possível excluir o último projeto")
	default:
		writeError(w, http.StatusNotFound, "Projeto não encontrado")
	}
}

// ---------- membros, etiquetas e colunas ----------

func (s *Server) handleListMembers(w http.ResponseWriter, r *http.Request) {
	userID := requestUserID(r)
	data := s.store.Snapshot()
	user := findUser(&data, userID)
	writeJSON(w, http.StatusOK, []model.User{user})
}

func (s *Server) handleListLabels(w http.ResponseWriter, r *http.Request) {
	userID := requestUserID(r)
	data := s.store.Snapshot()
	writeJSON(w, http.StatusOK, labelsOfUser(data, userID))
}

func (s *Server) handleListColumns(w http.ResponseWriter, r *http.Request) {
	userID := requestUserID(r)
	data := s.store.Snapshot()
	columns := columnsOfUser(data, userID)
	if projectID := r.URL.Query().Get("projectId"); projectID != "" {
		if !projectOwnedBy(data, projectID, userID) {
			writeError(w, http.StatusNotFound, "Projeto não encontrado")
			return
		}
		columns = columnsOf(&data, projectID)
	}
	writeJSON(w, http.StatusOK, orderedColumns(columns))
}

// ---------- tarefas ----------

func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	data := s.store.Snapshot()
	tasks := data.Tasks
	if projectID := r.URL.Query().Get("projectId"); projectID != "" {
		tasks = tasksOf(&data, projectID)
	}
	if assigneeID := r.URL.Query().Get("assigneeId"); assigneeID != "" {
		tasks = slices.DeleteFunc(tasks, func(t model.Task) bool { return t.AssigneeID != assigneeID })
	}
	sort.SliceStable(tasks, func(i, j int) bool { return tasks[i].Order < tasks[j].Order })
	writeJSON(w, http.StatusOK, tasks)
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var body taskRequest
	if !decodeBody(w, r, &body) {
		return
	}
	userID := requestUserID(r)
	var created model.Task
	err := s.store.Update(func(data *model.Data) {
		if body.ProjectID != nil && !projectOwnedBy(*data, *body.ProjectID, userID) {
			return
		}
		if body.ProjectID == nil {
			projects := projectsOf(*data, userID)
			if len(projects) == 0 {
				return
			}
			id := projects[0].ID
			body.ProjectID = &id
		}
		created = buildTask(data, body, nil, "", userID)
		if created.ProjectID == "" {
			return
		}
		data.Tasks = append(data.Tasks, created)
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	if created.ID == "" {
		writeError(w, http.StatusBadRequest, "Projeto não encontrado")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	var body taskRequest
	if !decodeBody(w, r, &body) {
		return
	}
	userID := requestUserID(r)
	var updated model.Task
	found := false
	err := s.store.Update(func(data *model.Data) {
		index := slices.IndexFunc(data.Tasks, func(t model.Task) bool { return t.ID == r.PathValue("id") })
		if index < 0 {
			return
		}
		current := data.Tasks[index]
		if !projectOwnedBy(*data, current.ProjectID, userID) {
			return
		}
		rebuilt := buildTask(data, body, &current, current.ID, userID)
		rebuilt.CreatedAt = current.CreatedAt
		data.Tasks[index] = rebuilt
		updated = rebuilt
		found = true
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "Tarefa não encontrada")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	userID := requestUserID(r)
	removed := false
	err := s.store.Update(func(data *model.Data) {
		index := slices.IndexFunc(data.Tasks, func(t model.Task) bool { return t.ID == id })
		if index < 0 {
			return
		}
		if !projectOwnedBy(*data, data.Tasks[index].ProjectID, userID) {
			return
		}
		data.Tasks = slices.Delete(data.Tasks, index, index+1)
		removed = true
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	if !removed {
		writeError(w, http.StatusNotFound, "Tarefa não encontrada")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// handleReorderTasks move uma tarefa para outra posição do quadro, reindexando
// a coluna de destino de forma contígua.
func (s *Server) handleReorderTasks(w http.ResponseWriter, r *http.Request) {
	var body reorderRequest
	if !decodeBody(w, r, &body) {
		return
	}
	if body.TaskID == nil || body.ColumnID == nil {
		writeError(w, http.StatusBadRequest, "Parâmetros inválidos")
		return
	}
	userID := requestUserID(r)
	var moved bool
	err := s.store.Update(func(data *model.Data) {
		task, ok := findTask(data, *body.TaskID)
		if !ok {
			return
		}
		if !projectOwnedBy(*data, task.ProjectID, userID) {
			return
		}
		column, ok := findColumn(data, *body.ColumnID)
		if !ok || task.ProjectID != column.ProjectID || !projectOwnedBy(*data, column.ProjectID, userID) {
			return
		}

		now := model.Now()
		siblings := make([]model.Task, 0, len(data.Tasks))
		for _, item := range data.Tasks {
			if item.ColumnID == column.ID && item.ID != task.ID {
				siblings = append(siblings, item)
			}
		}
		sort.SliceStable(siblings, func(i, j int) bool { return siblings[i].Order < siblings[j].Order })

		position := len(siblings)
		if body.Index != nil {
			position = clamp(*body.Index, 0, len(siblings))
		}
		siblings = slices.Insert(siblings, position, *task)

		for order := range siblings {
			siblings[order].ColumnID = column.ID
			siblings[order].Order = order
			siblings[order].UpdatedAt = now
		}
		if model.ColumnDone(*column) && siblings[position].CompletedAt == nil {
			siblings[position].CompletedAt = model.StrPtr(now)
		}

		positions := make(map[string]model.Task, len(siblings))
		for _, item := range siblings {
			positions[item.ID] = item
		}
		for index := range data.Tasks {
			if replacement, ok := positions[data.Tasks[index].ID]; ok {
				data.Tasks[index] = replacement
			}
		}
		moved = true
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	if !moved {
		writeError(w, http.StatusBadRequest, "Parâmetros inválidos")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// ---------- notas ----------

func (s *Server) handleListNotes(w http.ResponseWriter, r *http.Request) {
	userID := requestUserID(r)
	data := s.store.Snapshot()
	writeJSON(w, http.StatusOK, accessibleNotes(data, userID))
}

func (s *Server) handleCreateNote(w http.ResponseWriter, r *http.Request) {
	var body noteRequest
	if !decodeBody(w, r, &body) {
		return
	}
	userID := requestUserID(r)
	now := model.Now()
	note := model.Note{
		ID:         model.UID(),
		UserID:     userID,
		Title:      model.Text(body.Title, ""),
		Content:    model.Text(body.Content, ""),
		Color:      model.Text(body.Color, "default"),
		Pinned:     body.Pinned != nil && *body.Pinned,
		SharedWith: []string{},
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.store.Update(func(data *model.Data) {
		data.Notes = append(data.Notes, note)
	}); err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, note)
}

func (s *Server) handleShareNote(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email *string `json:"email"`
	}
	if !decodeBody(w, r, &body) {
		return
	}
	email := strings.ToLower(strings.TrimSpace(model.Text(body.Email, "")))
	if email == "" {
		writeError(w, http.StatusBadRequest, "E-mail obrigatório.")
		return
	}
	userID := requestUserID(r)
	var updated model.Note
	found := false
	err := s.store.Update(func(data *model.Data) {
		index := slices.IndexFunc(data.Notes, func(n model.Note) bool { return n.ID == r.PathValue("id") })
		if index < 0 || data.Notes[index].UserID != userID {
			return
		}
		targetID := ""
		for _, user := range data.Users {
			if strings.EqualFold(user.Email, email) {
				targetID = user.ID
				break
			}
		}
		if targetID == "" || targetID == userID {
			return
		}
		note := data.Notes[index]
		if !containsString(note.SharedWith, targetID) {
			note.SharedWith = append(note.SharedWith, targetID)
		}
		note.UpdatedAt = model.Now()
		data.Notes[index] = note
		updated = note
		found = true
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	if !found {
		writeError(w, http.StatusBadRequest, "Usuário não encontrado ou nota inválida.")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleUpdateNote(w http.ResponseWriter, r *http.Request) {
	var body noteRequest
	if !decodeBody(w, r, &body) {
		return
	}
	userID := requestUserID(r)
	var updated model.Note
	found := false
	err := s.store.Update(func(data *model.Data) {
		index := slices.IndexFunc(data.Notes, func(n model.Note) bool { return n.ID == r.PathValue("id") })
		if index < 0 || data.Notes[index].UserID != userID {
			return
		}
		note := data.Notes[index]
		note.Title = model.Text(body.Title, note.Title)
		note.Content = model.Text(body.Content, note.Content)
		note.Color = model.Text(body.Color, note.Color)
		if body.Pinned != nil {
			note.Pinned = *body.Pinned
		}
		note.UpdatedAt = model.Now()
		data.Notes[index] = note
		updated = note
		found = true
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "Nota não encontrada")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleDeleteNote(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	userID := requestUserID(r)
	removed := false
	err := s.store.Update(func(data *model.Data) {
		index := slices.IndexFunc(data.Notes, func(n model.Note) bool { return n.ID == id })
		if index < 0 || data.Notes[index].UserID != userID {
			return
		}
		data.Notes = slices.Delete(data.Notes, index, index+1)
		removed = true
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	if !removed {
		writeError(w, http.StatusNotFound, "Nota não encontrada")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// ---------- arquivos estáticos ----------

// staticHandler serve public/ e cai no index.html para rotas do SPA.
func (s *Server) staticHandler() http.Handler {
	files := http.FileServer(http.Dir(s.staticRoot))
	index := filepath.Join(s.staticRoot, "index.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := path.Clean("/" + strings.TrimPrefix(r.URL.Path, "/"))
		candidate := filepath.Join(s.staticRoot, filepath.FromSlash(strings.TrimPrefix(clean, "/")))
		relative, err := filepath.Rel(s.staticRoot, candidate)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			w.Header().Set("Cache-Control", "no-cache")
			s.serveIndex(w, index)
			return
		}
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			if clean != "/index.html" {
				w.Header().Set("Cache-Control", "no-cache")
			}
			files.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-cache")
		s.serveIndex(w, index)
	})
}

func (s *Server) serveIndex(w http.ResponseWriter, index string) {
	file, err := os.Open(index)
	if err != nil {
		writeError(w, http.StatusNotFound, "Página não encontrada")
		return
	}
	defer file.Close()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.Copy(w, file); err != nil {
		s.logger.Error("falha ao servir index.html", "erro", err)
	}
}

func initialsFor(name string) string {
	fields := strings.Fields(name)
	if len(fields) == 0 {
		return "?"
	}
	if len(fields) == 1 {
		return strings.ToUpper(string([]rune(fields[0])[:1]))
	}
	return strings.ToUpper(string([]rune(fields[0])[:1])) + strings.ToUpper(string([]rune(fields[len(fields)-1])[:1]))
}

func seedUserData(data *model.Data, user model.User) {
	workspaceID := model.UID()
	workspace := model.Workspace{
		ID:          workspaceID,
		Name:        "Meu workspace",
		Description: "Organização pessoal e profissional",
		OwnerID:     user.ID,
		CreatedAt:   model.Now(),
	}
	data.Workspaces = append(data.Workspaces, workspace)

	projects := []struct {
		name, description, icon, color string
	}{
		{"Operação principal", "Tarefas importantes do dia a dia", "✦", "#3f6152"},
		{"Estudos e crescimento", "Aprendizados e planos pessoais", "↗", "#3c6377"},
		{"Conteúdo", "Ideias, publicações e projetos criativos", "◈", "#905f18"},
	}
	for _, item := range projects {
		projectID := model.UID()
		project := model.Project{
			ID:          projectID,
			WorkspaceID: workspaceID,
			Name:        item.name,
			Description: item.description,
			Icon:        item.icon,
			Color:       item.color,
			OwnerID:     user.ID,
			CreatedAt:   model.Now(),
		}
		data.Projects = append(data.Projects, project)
		data.Columns = append(data.Columns, store.CreateColumns(projectID, nil)...)
	}

	for _, label := range []struct {
		name, color string
	}{
		{"Foco", "#3f6152"},
		{"Urgente", "#a8392f"},
		{"Equipe", "#3c6377"},
	} {
		data.Labels = append(data.Labels, model.Label{
			ID:     model.UID(),
			UserID: user.ID,
			Name:   label.name,
			Color:  label.color,
		})
	}
}

func workspacesOf(data model.Data, userID string) []model.Workspace {
	result := []model.Workspace{}
	for _, workspace := range data.Workspaces {
		if workspace.OwnerID == userID {
			result = append(result, workspace)
		}
	}
	return result
}

func projectsOf(data model.Data, userID string) []model.Project {
	result := []model.Project{}
	for _, project := range data.Projects {
		if project.OwnerID == userID {
			result = append(result, project)
		}
	}
	return result
}

func accessibleProjects(data model.Data, userID string) []model.Project {
	result := []model.Project{}
	for _, project := range data.Projects {
		if project.OwnerID == userID || containsString(project.SharedWith, userID) {
			result = append(result, project)
		}
	}
	return result
}

func projectIDsOf(projects []model.Project) map[string]bool {
	result := make(map[string]bool, len(projects))
	for _, project := range projects {
		result[project.ID] = true
	}
	return result
}

func columnsOfUser(data model.Data, userID string) []model.Column {
	projects := projectIDsOf(accessibleProjects(data, userID))
	result := []model.Column{}
	for _, column := range data.Columns {
		if projects[column.ProjectID] {
			result = append(result, column)
		}
	}
	return result
}

func tasksOfUser(data model.Data, userID string) []model.Task {
	projects := projectIDsOf(accessibleProjects(data, userID))
	result := []model.Task{}
	for _, task := range data.Tasks {
		if projects[task.ProjectID] {
			result = append(result, task)
		}
	}
	return result
}

func notesOfUser(data model.Data, userID string) []model.Note {
	result := []model.Note{}
	for _, note := range data.Notes {
		if note.UserID == userID {
			result = append(result, note)
		}
	}
	return result
}

func accessibleNotes(data model.Data, userID string) []model.Note {
	result := []model.Note{}
	seen := map[string]bool{}
	for _, note := range data.Notes {
		if note.UserID == userID || containsString(note.SharedWith, userID) {
			if seen[note.ID] {
				continue
			}
			seen[note.ID] = true
			result = append(result, note)
		}
	}
	return result
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func labelsOfUser(data model.Data, userID string) []model.Label {
	result := []model.Label{}
	for _, label := range data.Labels {
		if label.UserID == userID {
			result = append(result, label)
		}
	}
	return result
}

func projectOwnedBy(data model.Data, projectID, userID string) bool {
	project, ok := findProject(&data, projectID)
	return ok && project.OwnerID == userID
}

func publicUser(user model.User) model.User {
	user.PasswordHash = ""
	return user
}

func (s *Server) fail(w http.ResponseWriter, err error) {
	s.logger.Error("falha ao persistir alteração", "erro", err)
	writeError(w, http.StatusInternalServerError, "Não foi possível salvar as alterações.")
}

// ---------- construção de tarefas ----------

// buildTask resolve uma tarefa a partir do corpo recebido. Quando current não
// é nil, os campos ausentes preservam os valores já existentes — e um null
// explícito limpa o valor, como fazia o servidor anterior.
func buildTask(data *model.Data, body taskRequest, current *model.Task, id string, userID string) model.Task {
	now := model.Now()
	task := model.Task{
		Title:      "Tarefa sem título",
		Priority:   "medium",
		ProjectID:  data.Projects[0].ID,
		AssigneeID: data.CurrentUserID,
		Labels:     []string{},
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if current != nil {
		task = *current
	}
	task.ID = model.UID()
	switch {
	case current != nil:
		task.ID = current.ID
	case id != "":
		task.ID = id
	}

	if body.ProjectID != nil {
		task.ProjectID = *body.ProjectID
	}
	project, ok := findProject(data, task.ProjectID)
	if !ok || (userID != "" && project.OwnerID != userID) {
		projects := []model.Project{}
		for _, candidate := range data.Projects {
			if userID == "" || candidate.OwnerID == userID {
				projects = append(projects, candidate)
			}
		}
		if len(projects) > 0 {
			project = projects[0]
			task.ProjectID = project.ID
		} else if len(data.Projects) > 0 {
			project = data.Projects[0]
			task.ProjectID = project.ID
		}
	}
	projectColumns := orderedColumns(columnsOf(data, project.ID))

	if body.Title != nil {
		if title := strings.TrimSpace(*body.Title); title != "" {
			task.Title = title
		}
	}
	if body.Description != nil {
		task.Description = strings.TrimSpace(*body.Description)
	}
	if body.Priority != nil {
		task.Priority = model.NormalizePriority(*body.Priority)
	}
	if body.DueDate.Present {
		task.DueDate = model.CleanDate(body.DueDate.Value)
	}
	if body.AssigneeID != nil {
		task.AssigneeID = resolveAssignee(data, *body.AssigneeID)
	}
	if body.Labels != nil {
		labels := make([]string, 0, len(*body.Labels))
		for _, label := range *body.Labels {
			labels = append(labels, label)
		}
		task.Labels = slices.Clip(labels)
	}
	if body.CreatedAt != nil && *body.CreatedAt != "" {
		task.CreatedAt = *body.CreatedAt
	}
	if body.Order != nil {
		task.Order = *body.Order
	} else if current == nil {
		task.Order = len(projectColumns)
	}

	columnID := ""
	if body.ColumnID != nil {
		if requested, ok := findColumnIn(projectColumns, *body.ColumnID); ok {
			columnID = requested.ID
		}
	}
	if columnID == "" {
		if kept, ok := findColumnIn(projectColumns, task.ColumnID); ok {
			columnID = kept.ID
		}
	}
	if columnID == "" {
		columnID = projectColumns[0].ID
	}
	task.ColumnID = columnID
	task.UpdatedAt = now
	if model.IsDoneTitle(columnTitleOf(projectColumns, columnID)) {
		completed := now
		switch {
		case current != nil && current.CompletedAt != nil:
			completed = *current.CompletedAt
		case body.CompletedAt.Present && body.CompletedAt.Value != nil && *body.CompletedAt.Value != "":
			completed = *body.CompletedAt.Value
		}
		task.CompletedAt = model.StrPtr(completed)
	} else {
		task.CompletedAt = nil
	}
	return task
}

// ---------- busca em memória ----------

func findUser(data *model.Data, id string) model.User {
	if user, ok := findUserByID(data, id); ok {
		return user
	}
	return data.Users[0]
}

func findUserByID(data *model.Data, id string) (model.User, bool) {
	index := slices.IndexFunc(data.Users, func(u model.User) bool { return u.ID == id })
	if index < 0 {
		return model.User{}, false
	}
	return data.Users[index], true
}

func resolveAssignee(data *model.Data, requested string) string {
	if user, ok := findUserByID(data, requested); ok {
		return user.ID
	}
	return data.Users[0].ID
}

func findProject(data *model.Data, id string) (model.Project, bool) {
	index := slices.IndexFunc(data.Projects, func(p model.Project) bool { return p.ID == id })
	if index < 0 {
		return model.Project{}, false
	}
	return data.Projects[index], true
}

func findTask(data *model.Data, id string) (*model.Task, bool) {
	index := slices.IndexFunc(data.Tasks, func(t model.Task) bool { return t.ID == id })
	if index < 0 {
		return nil, false
	}
	return &data.Tasks[index], true
}

func findColumn(data *model.Data, id string) (*model.Column, bool) {
	index := slices.IndexFunc(data.Columns, func(c model.Column) bool { return c.ID == id })
	if index < 0 {
		return nil, false
	}
	return &data.Columns[index], true
}

func findColumnIn(columns []model.Column, id string) (*model.Column, bool) {
	index := slices.IndexFunc(columns, func(c model.Column) bool { return c.ID == id })
	if index < 0 {
		return nil, false
	}
	return &columns[index], true
}

func columnTitleOf(columns []model.Column, id string) string {
	if column, ok := findColumnIn(columns, id); ok {
		return column.Title
	}
	return ""
}

func columnsOf(data *model.Data, projectID string) []model.Column {
	return slices.DeleteFunc(slices.Clone(data.Columns), func(c model.Column) bool { return c.ProjectID != projectID })
}

func tasksOf(data *model.Data, projectID string) []model.Task {
	return slices.DeleteFunc(slices.Clone(data.Tasks), func(t model.Task) bool { return t.ProjectID != projectID })
}

func orderedColumns(columns []model.Column) []model.Column {
	sort.SliceStable(columns, func(i, j int) bool { return columns[i].Order < columns[j].Order })
	return columns
}

func clamp(value, low, high int) int {
	return min(max(value, low), high)
}
