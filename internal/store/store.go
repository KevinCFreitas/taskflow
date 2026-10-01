// Package store cuida da leitura, normalização e gravação do documento
// do Taskflow em um arquivo JSON, com escrita atômica e acesso seguro
// para requisições concorrentes.
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"taskflow/internal/model"
)

const currentUserID = "user-kevin"

// legacyColors mapeia a paleta anterior para a paleta atual. Sem isso, um
// storage ja gravado continuaria exibindo as cores antigas apos a troca.
var legacyColors = map[string]string{
	"#635bff": "#3f6152",
	"#6366f1": "#5a6472",
	"#0ea5e9": "#3c6377",
	"#ec4899": "#905f18",
	"#ef4444": "#a8392f",
	"#10b981": "#2b6b4f",
	"#f59e0b": "#905f18",
	"#64748b": "#5a6472",
	"#3b82f6": "#3c6377",
	"#8b5cf6": "#5b5170",
	"#f43f5e": "#8a4a5c",
}

// migrateColor troca cores conhecidas da paleta antiga pela atual, ignorando
// maiusculas e nomes de cor do CSS.
func migrateColor(value string) string {
	trimmed := strings.TrimSpace(value)
	if replacement, ok := legacyColors[strings.ToLower(trimmed)]; ok {
		return replacement
	}
	return trimmed
}

var defaultColumns = []model.Column{
	{ID: "todo", Title: "A fazer", Color: "#5a6472", Order: 0},
	{ID: "progress", Title: "Em andamento", Color: "#905f18", Order: 1},
	{ID: "review", Title: "Em revisão", Color: "#3c6377", Order: 2},
	{ID: "done", Title: "Concluído", Color: "#2b6b4f", Order: 3},
}

func defaultData() model.Data {
	return model.Data{
		SchemaVersion: 2,
		CurrentUserID: currentUserID,
		Users: []model.User{{
			ID:       currentUserID,
			Name:     "Kevin",
			Email:    "kevin@taskflow.local",
			Initials: "K",
			Color:    "#3f6152",
			Role:     "ADMIN",
		}},
		Workspaces: []model.Workspace{{
			ID:          "workspace-personal",
			Name:        "Meu workspace",
			Description: "Organização pessoal e profissional",
			OwnerID:     currentUserID,
			CreatedAt:   "2026-09-20T09:00:00.000Z",
		}},
		Projects: []model.Project{
			{
				ID:          "project-main",
				WorkspaceID: "workspace-personal",
				Name:        "Operação principal",
				Description: "Tarefas importantes do dia a dia",
				Icon:        "✦",
				Color:       "#3f6152",
				OwnerID:     currentUserID,
				CreatedAt:   "2026-09-20T09:00:00.000Z",
			},
			{
				ID:          "project-growth",
				WorkspaceID: "workspace-personal",
				Name:        "Estudos e crescimento",
				Description: "Aprendizados e planos pessoais",
				Icon:        "↗",
				Color:       "#3c6377",
				OwnerID:     currentUserID,
				CreatedAt:   "2026-09-21T09:00:00.000Z",
			},
			{
				ID:          "project-content",
				WorkspaceID: "workspace-personal",
				Name:        "Conteúdo",
				Description: "Ideias, publicações e projetos criativos",
				Icon:        "◈",
				Color:       "#905f18",
				OwnerID:     currentUserID,
				CreatedAt:   "2026-09-22T09:00:00.000Z",
			},
		},
		Columns: []model.Column{},
		Tasks:   []model.Task{},
		Notes:   []model.Note{},
		Labels: []model.Label{
			{ID: "label-focus", Name: "Foco", Color: "#3f6152"},
			{ID: "label-urgent", Name: "Urgente", Color: "#a8392f"},
			{ID: "label-team", Name: "Equipe", Color: "#3c6377"},
		},
	}
}

// Store mantém o documento em memória e o persiste no disco a cada alteração.
type Store struct {
	mu   sync.RWMutex
	file string
	data model.Data
}

// New carrega o documento do arquivo informado. Se o arquivo não existir,
// o conteúdo padrão é criado. O caminho é resolvido a partir de envVar.
func New(file string) (*Store, error) {
	s := &Store{file: file}
	data, err := readFile(file)
	if err != nil {
		return nil, err
	}
	s.data = normalize(data)
	if err := s.write(); err != nil {
		return nil, err
	}
	return s, nil
}

// ResolveFile devolve o caminho do arquivo de dados, considerando a variável
// de ambiente TASKFLOW_DATA_FILE e, por padrão, storage/data.json.
func ResolveFile(envVar, baseDir string) string {
	if custom := os.Getenv(envVar); custom != "" {
		return custom
	}
	return filepath.Join(baseDir, "storage", "data.json")
}

func readFile(file string) (model.Data, error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultData(), nil
		}
		return model.Data{}, err
	}
	var data model.Data
	if err := json.Unmarshal(raw, &data); err != nil {
		fmt.Fprintf(os.Stderr, "taskflow: dados inválidos em %s, recriando: %v\n", file, err)
		return defaultData(), nil
	}
	return data, nil
}

// Snapshot devolve uma cópia isolada do documento atual.
func (s *Store) Snapshot() model.Data {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data
}

// Update aplica uma mutação sobre o documento e persiste o resultado.
func (s *Store) Update(mutate func(data *model.Data)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	mutate(&s.data)
	s.data = normalize(s.data)
	return s.write()
}

func (s *Store) write() error {
	if err := os.MkdirAll(filepath.Dir(s.file), 0o755); err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	temporary := s.file + ".tmp"
	if err := os.WriteFile(temporary, append(encoded, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(temporary, s.file)
}

// normalize garante que o documento respeite os invariantes do domínio:
// listas nunca nulas, colunas existentes para todo projeto, tarefas sempre
// apontando para uma coluna válida e metadados de criação preenchidos.
func normalize(raw model.Data) model.Data {
	defaults := defaultData()
	data := raw
	data.SchemaVersion = 2
	if data.CurrentUserID == "" {
		data.CurrentUserID = currentUserID
	}
	if len(data.Users) == 0 {
		data.Users = defaults.Users
	}
	if len(data.Workspaces) == 0 {
		data.Workspaces = defaults.Workspaces
	}
	if len(data.Projects) == 0 {
		data.Projects = defaults.Projects
	}
	if len(data.Labels) == 0 {
		data.Labels = defaults.Labels
	}
	if data.Notes == nil {
		data.Notes = []model.Note{}
	}
	if data.Columns == nil {
		data.Columns = []model.Column{}
	}
	if data.Tasks == nil {
		data.Tasks = []model.Task{}
	}

	primaryWorkspace := data.Workspaces[0]
	projects := make([]model.Project, 0, len(data.Projects))
	for _, project := range data.Projects {
		if project.ID == "" {
			project.ID = model.UID()
		}
		if project.WorkspaceID == "" {
			project.WorkspaceID = primaryWorkspace.ID
		}
		if project.Name == "" {
			project.Name = "Novo projeto"
		}
		if project.Icon == "" {
			project.Icon = "✦"
		}
		if project.Color == "" {
			project.Color = "#3f6152"
		}
		if project.OwnerID == "" {
			project.OwnerID = data.CurrentUserID
		}
		if project.CreatedAt == "" {
			project.CreatedAt = model.Now()
		}
		projects = append(projects, project)
	}
	data.Projects = projects

	migratePalette(&data)

	mainProject := data.Projects[0]
	ensureColumns(&data, mainProject, true)
	for _, project := range data.Projects[1:] {
		ensureColumns(&data, project, false)
	}

	validColumnIDs := make(map[string]bool, len(data.Columns))
	for _, column := range data.Columns {
		validColumnIDs[column.ID] = true
	}
	projectIDs := make(map[string]bool, len(data.Projects))
	for _, project := range data.Projects {
		projectIDs[project.ID] = true
	}

	tasks := make([]model.Task, 0, len(data.Tasks))
	for _, task := range data.Tasks {
		if task.ID == "" {
			task.ID = model.UID()
		}
		if !projectIDs[task.ProjectID] {
			task.ProjectID = mainProject.ID
		}
		if !validColumnIDs[task.ColumnID] {
			task.ColumnID = firstColumnOf(data, task.ProjectID)
		}
		if task.Title == "" {
			task.Title = "Tarefa sem título"
		}
		task.Priority = model.NormalizePriority(task.Priority)
		task.DueDate = model.CleanDate(task.DueDate)
		if task.Labels == nil {
			task.Labels = []string{}
		}
		if task.AssigneeID == "" {
			task.AssigneeID = data.CurrentUserID
		}
		if task.CreatedAt == "" {
			task.CreatedAt = model.Now()
		}
		if task.UpdatedAt == "" {
			task.UpdatedAt = task.CreatedAt
		}
		if model.IsDoneTitle(columnTitle(data, task.ColumnID)) {
			if task.CompletedAt == nil {
				task.CompletedAt = model.StrPtr(task.UpdatedAt)
			}
		} else {
			task.CompletedAt = nil
		}
		tasks = append(tasks, task)
	}
	data.Tasks = tasks

	notes := make([]model.Note, 0, len(data.Notes))
	for _, note := range data.Notes {
		if note.ID == "" {
			note.ID = model.UID()
		}
		if note.Color == "" {
			note.Color = "default"
		}
		if note.CreatedAt == "" {
			note.CreatedAt = model.Now()
		}
		if note.UpdatedAt == "" {
			note.UpdatedAt = note.CreatedAt
		}
		notes = append(notes, note)
	}
	data.Notes = notes

	return data
}

// migratePalette atualiza as cores de usuários, projetos, etiquetas e colunas
// que ainda apontam para a paleta anterior. Cores livres do usuário são
// preservadas.
func migratePalette(data *model.Data) {
	for index := range data.Users {
		data.Users[index].Color = migrateColor(data.Users[index].Color)
	}
	for index := range data.Projects {
		data.Projects[index].Color = migrateColor(data.Projects[index].Color)
	}
	for index := range data.Labels {
		data.Labels[index].Color = migrateColor(data.Labels[index].Color)
	}
	for index := range data.Columns {
		data.Columns[index].Color = migrateColor(data.Columns[index].Color)
	}
}

// ensureColumns cria o conjunto padrão de colunas de um projeto. Quando
// reorderMain é verdadeiro, colunas legadas sem ordem explícita são realinhadas
// com a ordem canônica, preservando personalizações que já existam.
func ensureColumns(data *model.Data, project model.Project, reorderMain bool) {
	fixed := map[string]string(nil)
	if reorderMain {
		fixed = map[string]string{"todo": "todo", "progress": "progress", "review": "review", "done": "done"}
	}

	var own []int
	for index, column := range data.Columns {
		if column.ProjectID == "" || column.ProjectID == project.ID {
			own = append(own, index)
		}
	}

	if len(own) == 0 {
		data.Columns = append(data.Columns, CreateColumns(project.ID, fixed)...)
		return
	}

	if !reorderMain {
		return
	}

	hasCustomOrder := false
	for _, index := range own {
		if data.Columns[index].Order > 0 {
			hasCustomOrder = true
			break
		}
	}
	for position, index := range own {
		column := &data.Columns[index]
		column.ProjectID = project.ID
		if hasCustomOrder {
			continue
		}
		matched := -1
		for candidate, defaultColumn := range defaultColumns {
			sameID := defaultColumn.ID == column.ID
			sameTitle := strings.EqualFold(column.Title, defaultColumn.Title)
			if sameID || sameTitle {
				matched = candidate
				break
			}
		}
		if matched >= 0 {
			column.Order = matched
		} else {
			column.Order = position
		}
	}
}

// CreateColumns monta as colunas padrão de um projeto. O chamador pode fixar
// os identificadores por chave (usado pelo projeto principal, que historicamente
// usa ids como "todo" e "done"). Passar nil gera ids no formato "<projeto>-<etapa>".
func CreateColumns(projectID string, fixed map[string]string) []model.Column {
	columns := make([]model.Column, 0, len(defaultColumns))
	for _, column := range defaultColumns {
		id := projectID + "-" + column.ID
		if fixed != nil {
			if custom, ok := fixed[column.ID]; ok && custom != "" {
				id = custom
			}
		}
		columns = append(columns, model.Column{
			ID:        id,
			ProjectID: projectID,
			Title:     column.Title,
			Color:     column.Color,
			Order:     column.Order,
		})
	}
	return columns
}

func firstColumnOf(data model.Data, projectID string) string {
	for _, column := range data.Columns {
		if column.ProjectID == projectID {
			return column.ID
		}
	}
	return "todo"
}

func columnTitle(data model.Data, columnID string) string {
	for _, column := range data.Columns {
		if column.ID == columnID {
			return column.Title
		}
	}
	return ""
}
