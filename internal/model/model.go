// Package model define as entidades de domínio do Taskflow e o contrato
// JSON compartilhado entre a API e o armazenamento.
package model

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// User representa um membro do workspace.
type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Initials string `json:"initials"`
	Color    string `json:"color"`
	Role     string `json:"role"`
}

// Workspace agrupa projetos sob um mesmo contexto de trabalho.
type Workspace struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     string `json:"ownerId"`
	CreatedAt   string `json:"createdAt"`
}

// Project é a unidade de organização que possui colunas e tarefas.
type Project struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspaceId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
	OwnerID     string `json:"ownerId"`
	CreatedAt   string `json:"createdAt"`
}

// Column é uma etapa do quadro (kanban) de um projeto.
type Column struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Title     string `json:"title"`
	Color     string `json:"color"`
	Order     int    `json:"order"`
}

// Task é a unidade de trabalho atribuída a uma coluna.
type Task struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Priority    string   `json:"priority"`
	DueDate     *string  `json:"dueDate"`
	ColumnID    string   `json:"columnId"`
	Order       int      `json:"order"`
	ProjectID   string   `json:"projectId"`
	AssigneeID  string   `json:"assigneeId"`
	Labels      []string `json:"labels"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
	CompletedAt *string  `json:"completedAt"`
}

// Note é um cartão de texto livre (ideias, lembretes, decisões).
type Note struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Color     string `json:"color"`
	Pinned    bool   `json:"pinned"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// Label é uma etiqueta colorida aplicável a tarefas.
type Label struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// Data é o documento completo persistido em storage/data.json.
type Data struct {
	SchemaVersion int         `json:"schemaVersion"`
	CurrentUserID string      `json:"currentUserId"`
	Users         []User      `json:"users"`
	Workspaces    []Workspace `json:"workspaces"`
	Projects      []Project   `json:"projects"`
	Columns       []Column    `json:"columns"`
	Tasks         []Task      `json:"tasks"`
	Notes         []Note      `json:"notes"`
	Labels        []Label     `json:"labels"`
}

// Bootstrap é o payload consumido pelo front em /api/bootstrap.
type Bootstrap struct {
	CurrentUser *User       `json:"currentUser"`
	Workspace   *Workspace  `json:"workspace"`
	Workspaces  []Workspace `json:"workspaces"`
	Projects    []Project   `json:"projects"`
	Members     []User      `json:"members"`
	Columns     []Column    `json:"columns"`
	Tasks       []Task      `json:"tasks"`
	Notes       []Note      `json:"notes"`
	Labels      []Label     `json:"labels"`
}

// ProjectDetail é a resposta de GET /api/projects/{id}.
type ProjectDetail struct {
	Project
	Columns []Column `json:"columns"`
	Tasks   []Task   `json:"tasks"`
}

// IsDoneTitle detecta colunas de conclusão pelo título, mantendo a mesma
// regra usada historicamente pelo Taskflow ("Concluído", "Concluido", ...).
func IsDoneTitle(title string) bool {
	return strings.Contains(strings.ToLower(title), "conclu")
}

// ColumnDone informa se uma coluna representa um estado concluído.
func ColumnDone(c Column) bool { return IsDoneTitle(c.Title) }

// Text normaliza um texto recebido do cliente, aplicando o fallback
// informado quando o valor não veio preenchido.
func Text(value *string, fallback string) string {
	if value == nil {
		return fallback
	}
	return strings.TrimSpace(*value)
}

// NormalizePriority traduz os aliases aceitos de prioridade para os quatro
// valores canônicos: urgent, high, medium e low.
func NormalizePriority(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "urgente", "urgent":
		return "urgent"
	case "alta", "high":
		return "high"
	case "media", "médio", "medium":
		return "medium"
	case "baixa", "low":
		return "low"
	default:
		return "medium"
	}
}

// CleanDate devolve nil para datas vazias, mantendo a semântica de JSON null.
func CleanDate(value *string) *string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	cleaned := strings.TrimSpace(*value)
	return &cleaned
}

// StrPtr devolve um ponteiro para o texto informado.
func StrPtr(value string) *string { return &value }

// Now retorna o instante atual no formato ISO-8601 usado em toda a API.
func Now() string { return time.Now().UTC().Format("2006-01-02T15:04:05.000Z07:00") }

// UID gera um identificador único no formato UUID v4.
func UID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	raw[6] = (raw[6] & 0x0f) | 0x40
	raw[8] = (raw[8] & 0x3f) | 0x80
	encoded := hex.EncodeToString(raw[:])
	return strings.Join([]string{
		encoded[0:8], encoded[8:12], encoded[12:16], encoded[16:20], encoded[20:32],
	}, "-")
}
