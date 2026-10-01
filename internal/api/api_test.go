package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"taskflow/internal/api"
	"taskflow/internal/model"
	"taskflow/internal/store"
)

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	dataFile := filepath.Join(t.TempDir(), "nested", "data.json")
	dataStore, err := store.New(dataFile)
	if err != nil {
		t.Fatalf("não foi possível criar o armazenamento: %v", err)
	}
	return api.New(dataStore, "../../public", nil).Handler()
}

func do(t *testing.T, handler http.Handler, method, route string, body any) (int, []byte) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("não foi possível codificar o corpo: %v", err)
		}
		reader = bytes.NewReader(encoded)
	}
	request := httptest.NewRequest(method, route, reader)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder.Code, recorder.Body.Bytes()
}

func decodeInto(t *testing.T, payload []byte, target any) {
	t.Helper()
	if err := json.Unmarshal(payload, target); err != nil {
		t.Fatalf("resposta não é JSON válido: %v (%s)", err, payload)
	}
}

func TestFluxoCompletoDeProjetoETarefa(t *testing.T) {
	handler := newTestServer(t)

	t.Run("health responde ok", func(t *testing.T) {
		status, payload := do(t, handler, http.MethodGet, "/api/health", nil)
		if status != http.StatusOK {
			t.Fatalf("status inesperado: %d", status)
		}
		var health struct {
			OK      bool   `json:"ok"`
			Service string `json:"service"`
		}
		decodeInto(t, payload, &health)
		if !health.OK || health.Service != "taskflow" {
			t.Fatalf("health inesperado: %+v", health)
		}
	})

	t.Run("bootstrap entrega o workspace inicial", func(t *testing.T) {
		status, payload := do(t, handler, http.MethodGet, "/api/bootstrap", nil)
		if status != http.StatusOK {
			t.Fatalf("status inesperado: %d", status)
		}
		var boot model.Bootstrap
		decodeInto(t, payload, &boot)
		if boot.CurrentUser == nil || boot.CurrentUser.Name != "Kevin" {
			t.Fatalf("usuário atual inesperado: %+v", boot.CurrentUser)
		}
		if len(boot.Projects) < 3 {
			t.Fatalf("esperava ao menos 3 projetos, recebeu %d", len(boot.Projects))
		}
		if boot.Workspace == nil || boot.Workspace.Name != "Meu workspace" {
			t.Fatalf("workspace inesperado: %+v", boot.Workspace)
		}
	})

	var projectID string

	t.Run("cria projeto com quatro colunas", func(t *testing.T) {
		status, payload := do(t, handler, http.MethodPost, "/api/projects", map[string]any{
			"name":        "Projeto de teste",
			"description": "Validação automatizada",
			"icon":        "◇",
			"color":       "#123456",
		})
		if status != http.StatusCreated {
			t.Fatalf("status inesperado: %d (%s)", status, payload)
		}
		var project model.Project
		decodeInto(t, payload, &project)
		projectID = project.ID
		if project.Name != "Projeto de teste" || project.Icon != "◇" {
			t.Fatalf("projeto inesperado: %+v", project)
		}

		status, payload = do(t, handler, http.MethodGet, "/api/columns?projectId="+projectID, nil)
		if status != http.StatusOK {
			t.Fatalf("status inesperado: %d", status)
		}
		var columns []model.Column
		decodeInto(t, payload, &columns)
		if len(columns) != 4 {
			t.Fatalf("esperava 4 colunas, recebeu %d", len(columns))
		}
		for index, column := range columns {
			if column.Order != index {
				t.Fatalf("coluna %q fora de ordem: %d", column.Title, column.Order)
			}
		}
	})

	var taskID string

	t.Run("cria tarefa com prioridade e prazo", func(t *testing.T) {
		status, payload := do(t, handler, http.MethodPost, "/api/tasks", map[string]any{
			"title":       "Tarefa de teste",
			"description": "Validar persistência",
			"projectId":   projectID,
			"priority":    "urgent",
			"dueDate":     "2030-01-10",
		})
		if status != http.StatusCreated {
			t.Fatalf("status inesperado: %d (%s)", status, payload)
		}
		var task model.Task
		decodeInto(t, payload, &task)
		taskID = task.ID
		if task.ProjectID != projectID {
			t.Fatalf("projeto incorreto: %q", task.ProjectID)
		}
		if task.Priority != "urgent" {
			t.Fatalf("prioridade não normalizada: %q", task.Priority)
		}
		if task.DueDate == nil || *task.DueDate != "2030-01-10" {
			t.Fatalf("prazo inesperado: %+v", task.DueDate)
		}
		if task.CompletedAt != nil {
			t.Fatal("tarefa recém-criada não deveria estar concluída")
		}
	})

	t.Run("atualiza tarefa preservando campos ausentes", func(t *testing.T) {
		status, payload := do(t, handler, http.MethodPut, "/api/tasks/"+taskID, map[string]any{
			"title":    "Tarefa atualizada",
			"priority": "high",
		})
		if status != http.StatusOK {
			t.Fatalf("status inesperado: %d (%s)", status, payload)
		}
		var task model.Task
		decodeInto(t, payload, &task)
		if task.Title != "Tarefa atualizada" {
			t.Fatalf("título inesperado: %q", task.Title)
		}
		if task.Priority != "high" {
			t.Fatalf("prioridade inesperada: %q", task.Priority)
		}
		if task.Description != "Validar persistência" {
			t.Fatalf("descrição deveria ter sido preservada: %q", task.Description)
		}
		if task.DueDate == nil || *task.DueDate != "2030-01-10" {
			t.Fatal("prazo deveria ter sido preservado")
		}
	})

	t.Run("limpa o prazo quando null é enviado explicitamente", func(t *testing.T) {
		status, payload := do(t, handler, http.MethodPatch, "/api/tasks/"+taskID, map[string]any{
			"dueDate": nil,
		})
		if status != http.StatusOK {
			t.Fatalf("status inesperado: %d (%s)", status, payload)
		}
		var task model.Task
		decodeInto(t, payload, &task)
		if task.DueDate != nil {
			t.Fatalf("prazo deveria ter sido limpo, veio %q", *task.DueDate)
		}
	})

	t.Run("altera a coluna via PUT com apenas columnId", func(t *testing.T) {
		var columns []model.Column
		_, payload := do(t, handler, http.MethodGet, "/api/columns?projectId="+projectID, nil)
		decodeInto(t, payload, &columns)
		backlog := columns[0]

		status, body := do(t, handler, http.MethodPost, "/api/tasks", map[string]any{
			"title":     "Alternar conclusão",
			"projectId": projectID,
			"columnId":  backlog.ID,
		})
		if status != http.StatusCreated {
			t.Fatalf("status inesperado: %d (%s)", status, body)
		}
		var task model.Task
		decodeInto(t, body, &task)

		status, body = do(t, handler, http.MethodPut, "/api/tasks/"+task.ID, map[string]any{
			"columnId": columns[3].ID,
		})
		if status != http.StatusOK {
			t.Fatalf("status inesperado: %d (%s)", status, body)
		}
		decodeInto(t, body, &task)
		if task.ColumnID != columns[3].ID {
			t.Fatalf("tarefa não mudou de coluna: %q", task.ColumnID)
		}
		if task.CompletedAt == nil {
			t.Fatal("completedAt deveria ser preenchido ao mover para a coluna concluída")
		}

		status, body = do(t, handler, http.MethodPut, "/api/tasks/"+task.ID, map[string]any{
			"columnId": backlog.ID,
		})
		if status != http.StatusOK {
			t.Fatalf("status inesperado ao reabrir: %d (%s)", status, body)
		}
		decodeInto(t, body, &task)
		if task.ColumnID != backlog.ID {
			t.Fatalf("tarefa não voltou para a coluna inicial: %q", task.ColumnID)
		}
		if task.CompletedAt != nil {
			t.Fatal("completedAt deveria ser limpo ao sair da coluna concluída")
		}
		if status, body = do(t, handler, http.MethodDelete, "/api/tasks/"+task.ID, nil); status != http.StatusOK {
			t.Fatalf("status inesperado na limpeza: %d (%s)", status, body)
		}
	})

	t.Run("ignora columnId de outro projeto e preserva a coluna atual", func(t *testing.T) {
		var own []model.Column
		_, payload := do(t, handler, http.MethodGet, "/api/columns?projectId="+projectID, nil)
		decodeInto(t, payload, &own)

		var other []model.Column
		_, payload = do(t, handler, http.MethodGet, "/api/columns?projectId=project-main", nil)
		decodeInto(t, payload, &other)

		status, body := do(t, handler, http.MethodPost, "/api/tasks", map[string]any{
			"title":     "Coluna inválida",
			"projectId": projectID,
			"columnId":  own[0].ID,
		})
		if status != http.StatusCreated {
			t.Fatalf("status inesperado: %d (%s)", status, body)
		}
		var task model.Task
		decodeInto(t, body, &task)

		status, body = do(t, handler, http.MethodPut, "/api/tasks/"+task.ID, map[string]any{
			"columnId": other[3].ID,
		})
		if status != http.StatusOK {
			t.Fatalf("status inesperado: %d (%s)", status, body)
		}
		decodeInto(t, body, &task)
		if task.ColumnID != own[0].ID {
			t.Fatalf("aceitou coluna de outro projeto: %q", task.ColumnID)
		}
		if status, body = do(t, handler, http.MethodDelete, "/api/tasks/"+task.ID, nil); status != http.StatusOK {
			t.Fatalf("status inesperado na limpeza: %d (%s)", status, body)
		}
	})

	t.Run("preserva a coluna quando o corpo não manda columnId", func(t *testing.T) {
		var columns []model.Column
		_, payload := do(t, handler, http.MethodGet, "/api/columns?projectId="+projectID, nil)
		decodeInto(t, payload, &columns)

		status, body := do(t, handler, http.MethodPost, "/api/tasks", map[string]any{
			"title":     "Sem columnId",
			"projectId": projectID,
			"columnId":  columns[2].ID,
		})
		if status != http.StatusCreated {
			t.Fatalf("status inesperado: %d (%s)", status, body)
		}
		var task model.Task
		decodeInto(t, body, &task)

		status, body = do(t, handler, http.MethodPatch, "/api/tasks/"+task.ID, map[string]any{
			"priority": "low",
		})
		if status != http.StatusOK {
			t.Fatalf("status inesperado: %d (%s)", status, body)
		}
		decodeInto(t, body, &task)
		if task.ColumnID != columns[2].ID {
			t.Fatalf("coluna deveria ter sido preservada, ficou em %q", task.ColumnID)
		}
		if status, body = do(t, handler, http.MethodDelete, "/api/tasks/"+task.ID, nil); status != http.StatusOK {
			t.Fatalf("status inesperado na limpeza: %d (%s)", status, body)
		}
	})

	t.Run("move a tarefa entre colunas", func(t *testing.T) {
		var columns []model.Column
		_, payload := do(t, handler, http.MethodGet, "/api/columns?projectId="+projectID, nil)
		decodeInto(t, payload, &columns)

		progress := columns[1]
		status, body := do(t, handler, http.MethodPost, "/api/tasks/reorder", map[string]any{
			"taskId":   taskID,
			"columnId": progress.ID,
			"index":    0,
		})
		if status != http.StatusOK {
			t.Fatalf("status inesperado: %d (%s)", status, body)
		}

		_, payload = do(t, handler, http.MethodGet, "/api/tasks?projectId="+projectID, nil)
		var tasks []model.Task
		decodeInto(t, payload, &tasks)
		if len(tasks) != 1 || tasks[0].ColumnID != progress.ID || tasks[0].Order != 0 {
			t.Fatalf("tarefa não foi movida: %+v", tasks)
		}
	})

	t.Run("move para a coluna de conclusão e registra completedAt", func(t *testing.T) {
		var columns []model.Column
		_, payload := do(t, handler, http.MethodGet, "/api/columns?projectId="+projectID, nil)
		decodeInto(t, payload, &columns)

		status, body := do(t, handler, http.MethodPost, "/api/tasks/reorder", map[string]any{
			"taskId":   taskID,
			"columnId": columns[3].ID,
			"index":    0,
		})
		if status != http.StatusOK {
			t.Fatalf("status inesperado: %d (%s)", status, body)
		}

		_, payload = do(t, handler, http.MethodGet, "/api/tasks?projectId="+projectID, nil)
		var tasks []model.Task
		decodeInto(t, payload, &tasks)
		if tasks[0].CompletedAt == nil {
			t.Fatal("completedAt deveria ter sido preenchido ao mover para a coluna final")
		}
	})

	t.Run("rejeita mover tarefa para coluna de outro projeto", func(t *testing.T) {
		var columns []model.Column
		_, payload := do(t, handler, http.MethodGet, "/api/columns?projectId=project-main", nil)
		decodeInto(t, payload, &columns)

		status, _ := do(t, handler, http.MethodPost, "/api/tasks/reorder", map[string]any{
			"taskId":   taskID,
			"columnId": columns[0].ID,
			"index":    0,
		})
		if status != http.StatusBadRequest {
			t.Fatalf("esperava 400 ao cruzar projetos, recebeu %d", status)
		}
	})

	t.Run("reordena mantendo a coluna contígua", func(t *testing.T) {
		var columns []model.Column
		_, payload := do(t, handler, http.MethodGet, "/api/columns?projectId="+projectID, nil)
		decodeInto(t, payload, &columns)

		var created []string
		for _, title := range []string{"Alpha", "Beta", "Gamma"} {
			status, body := do(t, handler, http.MethodPost, "/api/tasks", map[string]any{
				"title":     title,
				"projectId": projectID,
				"columnId":  columns[0].ID,
			})
			if status != http.StatusCreated {
				t.Fatalf("status inesperado ao criar %q: %d (%s)", title, status, body)
			}
			var task model.Task
			decodeInto(t, body, &task)
			created = append(created, task.ID)
		}

		status, body := do(t, handler, http.MethodPost, "/api/tasks/reorder", map[string]any{
			"taskId":   created[2],
			"columnId": columns[0].ID,
			"index":    0,
		})
		if status != http.StatusOK {
			t.Fatalf("status inesperado: %d (%s)", status, body)
		}

		_, payload = do(t, handler, http.MethodGet, "/api/tasks?projectId="+projectID+"&assigneeId=user-kevin", nil)
		var tasks []model.Task
		decodeInto(t, payload, &tasks)

		orders := make([]int, 0, 3)
		for _, task := range tasks {
			if task.ColumnID == columns[0].ID {
				orders = append(orders, task.Order)
			}
		}
		slices.Sort(orders)
		for index, order := range orders {
			if order != index {
				t.Fatalf("ordens não ficaram contíguas: %v", orders)
			}
		}
	})

	t.Run("notas seguem o ciclo completo", func(t *testing.T) {
		status, payload := do(t, handler, http.MethodPost, "/api/notes", map[string]any{
			"title":   "Ideia",
			"content": "Anotar depois",
			"pinned":  true,
		})
		if status != http.StatusCreated {
			t.Fatalf("status inesperado: %d (%s)", status, payload)
		}
		var note model.Note
		decodeInto(t, payload, &note)
		if !note.Pinned {
			t.Fatal("nota deveria estar fixada")
		}

		status, payload = do(t, handler, http.MethodPut, "/api/notes/"+note.ID, map[string]any{
			"title": "Ideia revisada",
		})
		if status != http.StatusOK {
			t.Fatalf("status inesperado: %d", status)
		}
		decodeInto(t, payload, &note)
		if note.Title != "Ideia revisada" {
			t.Fatalf("título inesperado: %q", note.Title)
		}
		if note.Content != "Anotar depois" {
			t.Fatalf("conteúdo deveria ter sido preservado: %q", note.Content)
		}
		if !note.Pinned {
			t.Fatal("fixação deveria ter sido preservada")
		}

		if status, _ = do(t, handler, http.MethodDelete, "/api/notes/"+note.ID, nil); status != http.StatusOK {
			t.Fatalf("status inesperado ao excluir: %d", status)
		}
		if status, _ = do(t, handler, http.MethodDelete, "/api/notes/"+note.ID, nil); status != http.StatusNotFound {
			t.Fatalf("esperava 404 ao excluir duas vezes, recebeu %d", status)
		}
	})

	t.Run("exclui tarefa e projeto", func(t *testing.T) {
		if status, payload := do(t, handler, http.MethodDelete, "/api/tasks/"+taskID, nil); status != http.StatusOK {
			t.Fatalf("status inesperado ao excluir tarefa: %d (%s)", status, payload)
		}
		if status, payload := do(t, handler, http.MethodDelete, "/api/projects/"+projectID, nil); status != http.StatusOK {
			t.Fatalf("status inesperado ao excluir projeto: %d (%s)", status, payload)
		}
		if status, _ := do(t, handler, http.MethodGet, "/api/projects/"+projectID, nil); status != http.StatusNotFound {
			t.Fatalf("esperava 404 para projeto removido, recebeu %d", status)
		}
	})

	t.Run("não permite excluir o último projeto", func(t *testing.T) {
		_, payload := do(t, handler, http.MethodGet, "/api/projects", nil)
		var projects []model.Project
		decodeInto(t, payload, &projects)

		for _, project := range projects[1:] {
			if status, _ := do(t, handler, http.MethodDelete, "/api/projects/"+project.ID, nil); status != http.StatusOK {
				t.Fatalf("falha ao remover %q", project.Name)
			}
		}

		_, payload = do(t, handler, http.MethodGet, "/api/projects", nil)
		decodeInto(t, payload, &projects)
		if len(projects) != 1 {
			t.Fatalf("esperava 1 projeto restante, recebeu %d", len(projects))
		}
		if status, _ := do(t, handler, http.MethodDelete, "/api/projects/"+projects[0].ID, nil); status != http.StatusBadRequest {
			t.Fatalf("esperava 400 ao remover o último projeto, recebeu %d", status)
		}
	})
}

func TestRotasDoSPA(t *testing.T) {
	handler := newTestServer(t)

	for _, route := range []string{"/", "/qualquer/rota", "/projetos/abc"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, route, nil))
		if recorder.Code != http.StatusOK {
			t.Fatalf("rota %q deveria servir o index, recebeu %d", route, recorder.Code)
		}
		if contentType := recorder.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
			t.Fatalf("rota %q com Content-Type inesperado: %q", route, contentType)
		}
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/css/style.css", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("css deveria ser servido, recebeu %d", recorder.Code)
	}
}

func TestCorpoInvalidoRetorna400(t *testing.T) {
	handler := newTestServer(t)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBufferString("{ isso não é json"))
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400 para corpo inválido, recebeu %d", recorder.Code)
	}
	var errorPayload struct {
		Error string `json:"error"`
	}
	decodeInto(t, recorder.Body.Bytes(), &errorPayload)
	if errorPayload.Error == "" {
		t.Fatal("resposta de erro deveria ter a propriedade error preenchida")
	}
}

func TestMigracaoDaPaletaAntiga(t *testing.T) {
	dataFile := filepath.Join(t.TempDir(), "data.json")
	legacy := `{
  "schemaVersion": 1,
  "currentUserId": "user-kevin",
  "users": [{"id":"user-kevin","name":"Kevin","initials":"K","color":"#635bff"}],
  "workspaces": [{"id":"w1","name":"Workspace"}],
  "projects": [
    {"id":"p1","workspaceId":"w1","name":"Principal","color":"#635bff"},
    {"id":"p2","workspaceId":"w1","name":"Estudos","color":"#0ea5e9"},
    {"id":"p3","workspaceId":"w1","name":"Conteudo","color":"#ec4899"}
  ],
  "labels": [{"id":"l1","name":"Foco","color":"#635bff"}],
  "columns": [
    {"id":"todo","projectId":"p1","title":"A fazer","color":"#6366f1","order":0},
    {"id":"done","projectId":"p1","title":"Concluído","color":"#10b981","order":1}
  ]
}`
	if err := os.WriteFile(dataFile, []byte(legacy), 0o644); err != nil {
		t.Fatalf("não foi possível gravar o storage legado: %v", err)
	}
	if _, err := store.New(dataFile); err != nil {
		t.Fatalf("não foi possível carregar o storage legado: %v", err)
	}

	dataStore, err := store.New(dataFile)
	if err != nil {
		t.Fatalf("não foi possível abrir o armazenamento: %v", err)
	}
	data := dataStore.Snapshot()

	if strings.EqualFold(data.Users[0].Color, "#635bff") {
		t.Fatalf("cor do usuário não migrada: %q", data.Users[0].Color)
	}
	for _, project := range data.Projects {
		if strings.EqualFold(project.Color, "#635bff") ||
			strings.EqualFold(project.Color, "#0ea5e9") ||
			strings.EqualFold(project.Color, "#ec4899") {
			t.Fatalf("projeto %q ficou com a cor antiga: %q", project.Name, project.Color)
		}
	}
	for _, label := range data.Labels {
		if strings.EqualFold(label.Color, "#635bff") {
			t.Fatalf("etiqueta %q ficou com a cor antiga: %q", label.Name, label.Color)
		}
	}
	for _, column := range data.Columns {
		if strings.EqualFold(column.Color, "#6366f1") || strings.EqualFold(column.Color, "#10b981") {
			t.Fatalf("coluna %q ficou com a cor antiga: %q", column.Title, column.Color)
		}
	}
}

func TestMigracaoPreservaCoresLivres(t *testing.T) {
	dataFile := filepath.Join(t.TempDir(), "data.json")
	custom := `{
  "schemaVersion": 1,
  "currentUserId": "user-kevin",
  "users": [{"id":"user-kevin","name":"Kevin","initials":"K","color":"#123456"}],
  "workspaces": [{"id":"w1","name":"Workspace"}],
  "projects": [
    {"id":"p1","workspaceId":"w1","name":"Um","color":"#123456"},
    {"id":"p2","workspaceId":"w1","name":"Dois","color":"#abcdef"}
  ],
  "labels": [{"id":"l1","name":"Foco","color":"#123456"}]
}`
	if err := os.WriteFile(dataFile, []byte(custom), 0o644); err != nil {
		t.Fatalf("não foi possível gravar o storage: %v", err)
	}
	dataStore, err := store.New(dataFile)
	if err != nil {
		t.Fatalf("não foi possível abrir o armazenamento: %v", err)
	}
	data := dataStore.Snapshot()

	if data.Users[0].Color != "#123456" {
		t.Fatalf("cor personalizada do usuário foi alterada: %q", data.Users[0].Color)
	}
	if data.Projects[0].Color != "#123456" || data.Projects[1].Color != "#abcdef" {
		t.Fatalf("cores personalizadas foram alteradas: %q, %q", data.Projects[0].Color, data.Projects[1].Color)
	}
	if data.Labels[0].Color != "#123456" {
		t.Fatalf("cor personalizada da etiqueta foi alterada: %q", data.Labels[0].Color)
	}
}
