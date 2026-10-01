// Package api expõe a API HTTP do Taskflow e serve o front estático.
package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

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
	mux.HandleFunc("GET /api/bootstrap", s.handleBootstrap)

	mux.HandleFunc("GET /api/workspaces", s.handleListWorkspaces)
	mux.HandleFunc("POST /api/workspaces", s.handleCreateWorkspace)

	mux.HandleFunc("GET /api/projects", s.handleListProjects)
	mux.HandleFunc("POST /api/projects", s.handleCreateProject)
	mux.HandleFunc("GET /api/projects/{id}", s.handleGetProject)
	mux.HandleFunc("PUT /api/projects/{id}", s.handleUpdateProject)
	mux.HandleFunc("PATCH /api/projects/{id}", s.handleUpdateProject)
	mux.HandleFunc("DELETE /api/projects/{id}", s.handleDeleteProject)

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

	mux.Handle("GET /", s.staticHandler())

	return s.recoverPanic(mux)
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

func (s *Server) handleBootstrap(w http.ResponseWriter, _ *http.Request) {
	data := s.store.Snapshot()
	workspace := data.Workspaces[0]
	currentUser := findUser(&data, data.CurrentUserID)
	writeJSON(w, http.StatusOK, model.Bootstrap{
		CurrentUser: &currentUser,
		Workspace:   &workspace,
		Workspaces:  data.Workspaces,
		Projects:    data.Projects,
		Members:     data.Users,
		Columns:     data.Columns,
		Tasks:       data.Tasks,
		Notes:       data.Notes,
		Labels:      data.Labels,
	})
}

// ---------- workspaces ----------

func (s *Server) handleListWorkspaces(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.store.Snapshot().Workspaces)
}

func (s *Server) handleCreateWorkspace(w http.ResponseWriter, r *http.Request) {
	var body workspaceRequest
	if !decodeBody(w, r, &body) {
		return
	}
	var created model.Workspace
	err := s.store.Update(func(data *model.Data) {
		created = model.Workspace{
			ID:          model.UID(),
			Name:        model.Text(body.Name, "Novo workspace"),
			Description: model.Text(body.Description, ""),
			OwnerID:     data.CurrentUserID,
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

func (s *Server) handleListProjects(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.store.Snapshot().Projects)
}

func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var body projectRequest
	if !decodeBody(w, r, &body) {
		return
	}
	var created model.Project
	err := s.store.Update(func(data *model.Data) {
		workspaceID := "workspace-personal"
		if len(data.Workspaces) > 0 {
			workspaceID = data.Workspaces[0].ID
		}
		created = model.Project{
			ID:          model.UID(),
			WorkspaceID: workspaceID,
			Name:        model.Text(body.Name, "Novo projeto"),
			Description: model.Text(body.Description, ""),
			Icon:        model.Text(body.Icon, "✦"),
			Color:       model.Text(body.Color, "#3f6152"),
			OwnerID:     data.CurrentUserID,
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
	data := s.store.Snapshot()
	project, ok := findProject(&data, r.PathValue("id"))
	if !ok {
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
	var updated model.Project
	found := false
	err := s.store.Update(func(data *model.Data) {
		project, ok := findProject(data, r.PathValue("id"))
		if !ok {
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

func (s *Server) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var (
		deleted bool
		lastOne bool
	)
	err := s.store.Update(func(data *model.Data) {
		index := slices.IndexFunc(data.Projects, func(p model.Project) bool { return p.ID == id })
		if index < 0 {
			return
		}
		if len(data.Projects) == 1 {
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

func (s *Server) handleListMembers(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.store.Snapshot().Users)
}

func (s *Server) handleListLabels(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.store.Snapshot().Labels)
}

func (s *Server) handleListColumns(w http.ResponseWriter, r *http.Request) {
	data := s.store.Snapshot()
	columns := data.Columns
	if projectID := r.URL.Query().Get("projectId"); projectID != "" {
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
	var created model.Task
	err := s.store.Update(func(data *model.Data) {
		created = buildTask(data, body, nil, "")
		data.Tasks = append(data.Tasks, created)
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	var body taskRequest
	if !decodeBody(w, r, &body) {
		return
	}
	var updated model.Task
	found := false
	err := s.store.Update(func(data *model.Data) {
		index := slices.IndexFunc(data.Tasks, func(t model.Task) bool { return t.ID == r.PathValue("id") })
		if index < 0 {
			return
		}
		current := data.Tasks[index]
		rebuilt := buildTask(data, body, &current, current.ID)
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
	removed := false
	err := s.store.Update(func(data *model.Data) {
		index := slices.IndexFunc(data.Tasks, func(t model.Task) bool { return t.ID == id })
		if index < 0 {
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
	var moved bool
	err := s.store.Update(func(data *model.Data) {
		task, ok := findTask(data, *body.TaskID)
		if !ok {
			return
		}
		column, ok := findColumn(data, *body.ColumnID)
		if !ok || task.ProjectID != column.ProjectID {
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

func (s *Server) handleListNotes(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.store.Snapshot().Notes)
}

func (s *Server) handleCreateNote(w http.ResponseWriter, r *http.Request) {
	var body noteRequest
	if !decodeBody(w, r, &body) {
		return
	}
	now := model.Now()
	note := model.Note{
		ID:        model.UID(),
		Title:     model.Text(body.Title, ""),
		Content:   model.Text(body.Content, ""),
		Color:     model.Text(body.Color, "default"),
		Pinned:    body.Pinned != nil && *body.Pinned,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.store.Update(func(data *model.Data) {
		data.Notes = append(data.Notes, note)
	}); err != nil {
		s.fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, note)
}

func (s *Server) handleUpdateNote(w http.ResponseWriter, r *http.Request) {
	var body noteRequest
	if !decodeBody(w, r, &body) {
		return
	}
	var updated model.Note
	found := false
	err := s.store.Update(func(data *model.Data) {
		index := slices.IndexFunc(data.Notes, func(n model.Note) bool { return n.ID == r.PathValue("id") })
		if index < 0 {
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
	removed := false
	err := s.store.Update(func(data *model.Data) {
		index := slices.IndexFunc(data.Notes, func(n model.Note) bool { return n.ID == id })
		if index < 0 {
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
		clean := filepath.Clean("/" + strings.TrimPrefix(r.URL.Path, "/"))
		candidate := filepath.Join(s.staticRoot, clean)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
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

func (s *Server) fail(w http.ResponseWriter, err error) {
	s.logger.Error("falha ao persistir alteração", "erro", err)
	writeError(w, http.StatusInternalServerError, "Não foi possível salvar as alterações.")
}

// ---------- construção de tarefas ----------

// buildTask resolve uma tarefa a partir do corpo recebido. Quando current não
// é nil, os campos ausentes preservam os valores já existentes — e um null
// explícito limpa o valor, como fazia o servidor anterior.
func buildTask(data *model.Data, body taskRequest, current *model.Task, id string) model.Task {
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
	if !ok {
		project = data.Projects[0]
		task.ProjectID = project.ID
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
