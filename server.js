const express = require('express');
const path = require('path');
const fs = require('fs');
const crypto = require('crypto');

const app = express();
const PORT = Number(process.env.PORT) || 3000;
const DATA_DIR = path.join(__dirname, 'storage');
const DATA_FILE = process.env.TASKFLOW_DATA_FILE || path.join(DATA_DIR, 'data.json');
const CURRENT_USER_ID = 'user-kevin';

app.use(express.json({ limit: '1mb' }));
app.use(express.static(path.join(__dirname, 'public')));

const DEFAULT_COLUMNS = [
  { id: 'todo', title: 'A fazer', color: '#64748b', order: 0 },
  { id: 'progress', title: 'Em andamento', color: '#f59e0b', order: 1 },
  { id: 'review', title: 'Em revisão', color: '#0ea5e9', order: 2 },
  { id: 'done', title: 'Concluído', color: '#10b981', order: 3 }
];

const DEFAULT_DATA = {
  schemaVersion: 2,
  currentUserId: CURRENT_USER_ID,
  users: [
    {
      id: CURRENT_USER_ID,
      name: 'Kevin',
      email: 'kevin@taskflow.local',
      initials: 'K',
      color: '#635bff',
      role: 'ADMIN'
    }
  ],
  workspaces: [
    {
      id: 'workspace-personal',
      name: 'Meu workspace',
      description: 'Organização pessoal e profissional',
      ownerId: CURRENT_USER_ID,
      createdAt: '2026-09-20T09:00:00.000Z'
    }
  ],
  projects: [
    {
      id: 'project-main',
      workspaceId: 'workspace-personal',
      name: 'Operação principal',
      description: 'Tarefas importantes do dia a dia',
      icon: '✦',
      color: '#635bff',
      ownerId: CURRENT_USER_ID,
      createdAt: '2026-09-20T09:00:00.000Z'
    },
    {
      id: 'project-growth',
      workspaceId: 'workspace-personal',
      name: 'Estudos e crescimento',
      description: 'Aprendizados e planos pessoais',
      icon: '↗',
      color: '#0ea5e9',
      ownerId: CURRENT_USER_ID,
      createdAt: '2026-09-21T09:00:00.000Z'
    },
    {
      id: 'project-content',
      workspaceId: 'workspace-personal',
      name: 'Conteúdo',
      description: 'Ideias, publicações e projetos criativos',
      icon: '◈',
      color: '#ec4899',
      ownerId: CURRENT_USER_ID,
      createdAt: '2026-09-22T09:00:00.000Z'
    }
  ],
  columns: [],
  tasks: [],
  notes: [],
  labels: [
    { id: 'label-focus', name: 'Foco', color: '#635bff' },
    { id: 'label-urgent', name: 'Urgente', color: '#ef4444' },
    { id: 'label-team', name: 'Equipe', color: '#0ea5e9' }
  ]
};

let db = loadData();
saveData(db);

function clone(value) {
  return JSON.parse(JSON.stringify(value));
}

function loadData() {
  try {
    if (fs.existsSync(DATA_FILE)) {
      const raw = JSON.parse(fs.readFileSync(DATA_FILE, 'utf-8'));
      return normalizeData({ ...clone(DEFAULT_DATA), ...raw });
    }
  } catch (error) {
    console.error('Erro ao carregar dados, usando dados padrão:', error.message);
  }
  const initial = normalizeData(clone(DEFAULT_DATA));
  saveData(initial);
  return initial;
}

function normalizeData(raw) {
  const data = { ...clone(DEFAULT_DATA), ...raw };
  data.schemaVersion = 2;
  data.currentUserId = data.currentUserId || CURRENT_USER_ID;
  data.users = Array.isArray(data.users) && data.users.length ? data.users : clone(DEFAULT_DATA.users);
  data.workspaces = Array.isArray(data.workspaces) && data.workspaces.length ? data.workspaces : clone(DEFAULT_DATA.workspaces);
  data.projects = Array.isArray(data.projects) && data.projects.length ? data.projects : clone(DEFAULT_DATA.projects);
  data.notes = Array.isArray(data.notes) ? data.notes : [];
  data.labels = Array.isArray(data.labels) && data.labels.length ? data.labels : clone(DEFAULT_DATA.labels);
  data.columns = Array.isArray(data.columns) ? data.columns : [];

  const workspace = data.workspaces[0];
  data.projects = data.projects.map((project) => ({
    id: String(project.id || uid()),
    workspaceId: project.workspaceId || workspace.id,
    name: String(project.name || 'Novo projeto').trim(),
    description: String(project.description || '').trim(),
    icon: String(project.icon || '✦'),
    color: String(project.color || '#635bff'),
    ownerId: project.ownerId || data.currentUserId,
    createdAt: pr oject.createdAt || new Date().toISOString()
  }));

  const mainProject = data.projects[0];
  const mainColumns = data.columns.filter((column) => !column.projectId || column.projectId === mainProject.id);
  if (!mainColumns.length) {
    data.columns.push(...createColumns(mainProject.id, { todo: 'todo', progress: 'progress', review: 'review', done: 'done' }));
  } else {
    const hasCustomOrder = mainColumns.some((column) => Number.isFinite(+column.order) && +column.order > 0);
    mainColumns.forEach((column, index) => {
      const defaultIndex = DEFAULT_COLUMNS.findIndex((defaultColumn) => defaultColumn.id === column.id || defaultColumn.title.toLowerCase() === String(column.title || '').toLowerCase());
      column.projectId = mainProject.id;
      column.order = hasCustomOrder && Number.isFinite(+column.order) ? +column.order : defaultIndex >= 0 ? defaultIndex : index;
    });
  }

  data.projects.forEach((project) => {
    const hasColumns = data.columns.some((column) => column.projectId === project.id);
    if (!hasColumns) data.columns.push(...createColumns(project.id));
  });

  data.columns = data.columns.map((column) => ({
    id: String(column.id || uid()),
    projectId: column.projectId || mainProject.id,
    title: String(column.title || 'Nova etapa').trim(),
    color: String(column.color || '#64748b'),
    order: Number.isFinite(+column.order) ? +column.order : 0
  }));

  const validColumnIds = new Set(data.columns.map((column) => column.id));
  data.tasks = (Array.isArray(data.tasks) ? data.tasks : []).map((task) => {
    const projectId = data.projects.some((project) => project.id === task.projectId) ? task.projectId : mainProject.id;
    const columnId = validColumnIds.has(task.columnId) ? task.columnId : (data.columns.find((column) => column.projectId === projectId)?.id || 'todo');
    const done = data.columns.find((column) => column.id === columnId)?.title.toLowerCase().includes('conclu');
    return {
      id: String(task.id || uid()),
      title: String(task.title || 'Tarefa sem título').trim(),
      description: String(task.description || '').trim(),
      priority: priorityValue(task.priority),
      dueDate: task.dueDate || null,
      columnId,
      order: Number.isFinite(+task.order) ? +task.order : 0,
      projectId,
      assigneeId: task.assigneeId || data.currentUserId,
      labels: Array.isArray(task.labels) ? task.labels : [],
      createdAt: task.createdAt || new Date().toISOString(),
      updatedAt: task.updatedAt || task.createdAt || new Date().toISOString(),
      completedAt: done ? task.completedAt || task.updatedAt || task.createdAt || new Date().toISOString() : task.completedAt || null
    };
  });

  data.notes = data.notes.map((note) => ({
    id: String(note.id || uid()),
    title: String(note.title || '').trim(),
    content: String(note.content || '').trim(),
    color: note.color || 'default',
    pinned: !!note.pinned,
    createdAt: note.createdAt || new Date().toISOString(),
    updatedAt: note.updatedAt || note.createdAt || new Date().toISOString()
  }));

  return data;
}

function createColumns(projectId, ids = {}) {
  return DEFAULT_COLUMNS.map((column, index) => ({
    id: ids[column.id] || `${projectId}-${column.id}`,
    projectId,
    title: column.title,
    color: column.color,
    order: index
  }));
}

function saveData(value = db) {
  const dataDirectory = path.dirname(DATA_FILE);
  if (!fs.existsSync(dataDirectory)) fs.mkdirSync(dataDirectory, { recursive: true });
  const temporaryFile = `${DATA_FILE}.tmp`;
  fs.writeFileSync(temporaryFile, JSON.stringify(value, null, 2), 'utf-8');
  fs.renameSync(temporaryFile, DATA_FILE);
}

function uid() {
  return crypto.randomUUID();
}

function text(value, fallback = '') {
  return String(value ?? fallback).trim();
}

function priorityValue(value) {
  const normalized = String(value || '').toLowerCase();
  const aliases = { urgente: 'urgent', urgent: 'urgent', alta: 'high', high: 'high', media: 'medium', medium: 'medium', baixa: 'low', low: 'low' };
  return aliases[normalized] || 'medium';
}

function getProject(projectId) {
  return db.projects.find((project) => project.id === projectId);
}

function getColumn(columnId) {
  return db.columns.find((column) => column.id === columnId);
}

function getUser(userId) {
  return db.users.find((user) => user.id === userId) || db.users[0];
}

function normalizeTask(body, id) {
  const project = getProject(body.projectId) || db.projects[0];
  const projectColumns = db.columns.filter((column) => column.projectId === project.id).sort((a, b) => a.order - b.order);
  const requestedColumn = getColumn(body.columnId);
  const column = requestedColumn && requestedColumn.projectId === project.id ? requestedColumn : projectColumns[0];
  const now = new Date().toISOString();
  const done = column.title.toLowerCase().includes('conclu');
  return {
    id: id || uid(),
    title: text(body.title, 'Tarefa sem título'),
    description: text(body.description),
    priority: priorityValue(body.priority),
    dueDate: body.dueDate || null,
    columnId: column.id,
    order: Number.isFinite(+body.order) ? +body.order : projectColumns.filter((item) => item.id !== id).length,
    projectId: project.id,
    assigneeId: getUser(body.assigneeId)?.id || db.currentUserId,
    labels: Array.isArray(body.labels) ? body.labels.filter((label) => typeof label === 'string').slice(0, 8) : [],
    createdAt: body.createdAt || now,
    updatedAt: now,
    completedAt: done ? body.completedAt || now : null
  };
}

function publicUser(user) {
  if (!user) return null;
  return { id: user.id, name: user.name, email: user.email, initials: user.initials, color: user.color, role: user.role };
}

function bootstrapPayload() {
  const workspace = db.workspaces.find((item) => item.id === db.workspaces[0].id) || db.workspaces[0];
  return {
    currentUser: publicUser(getUser(db.currentUserId)),
    workspace,
    workspaces: db.workspaces,
    projects: db.projects,
    members: db.users.map(publicUser),
    columns: db.columns,
    tasks: db.tasks,
    notes: db.notes,
    labels: db.labels
  };
}

app.get('/api/health', (req, res) => res.json({ ok: true, service: 'taskflow' }));

app.get('/api/bootstrap', (req, res) => res.json(bootstrapPayload()));

app.get('/api/workspaces', (req, res) => res.json(db.workspaces));

app.post('/api/workspaces', (req, res) => {
  const workspace = {
    id: uid(),
    name: text(req.body.name, 'Novo workspace'),
    description: text(req.body.description),
    ownerId: db.currentUserId,
    createdAt: new Date().toISOString()
  };
  db.workspaces.push(workspace);
  saveData();
  res.status(201).json(workspace);
});

app.get('/api/projects', (req, res) => res.json(db.projects));

app.post('/api/projects', (req, res) => {
  const project = {
    id: uid(),
    workspaceId: db.workspaces[0]?.id || 'workspace-personal',
    name: text(req.body.name, 'Novo projeto'),
    description: text(req.body.description),
    icon: text(req.body.icon, '✦'),
    color: text(req.body.color, '#635bff'),
    ownerId: db.currentUserId,
    createdAt: new Date().toISOString()
  };
  db.projects.push(project);
  db.columns.push(...createColumns(project.id));
  saveData();
  res.status(201).json(project);
});

app.get('/api/projects/:id', (req, res) => {
  const project = getProject(req.params.id);
  if (!project) return res.status(404).json({ error: 'Projeto não encontrado' });
  res.json({ ...project, columns: db.columns.filter((column) => column.projectId === project.id), tasks: db.tasks.filter((task) => task.projectId === project.id) });
});

app.patch('/api/projects/:id', (req, res) => {
  const project = getProject(req.params.id);
  if (!project) return res.status(404).json({ error: 'Projeto não encontrado' });
  Object.assign(project, {
    name: text(req.body.name, project.name),
    description: text(req.body.description, project.description),
    icon: text(req.body.icon, project.icon),
    color: text(req.body.color, project.color)
  });
  saveData();
  res.json(project);
});

app.put('/api/projects/:id', (req, res) => {
  const project = getProject(req.params.id);
  if (!project) return res.status(404).json({ error: 'Projeto não encontrado' });
  Object.assign(project, {
    name: text(req.body.name, project.name),
    description: text(req.body.description, project.description),
    icon: text(req.body.icon, project.icon),
    color: text(req.body.color, project.color)
  });
  saveData();
  res.json(project);
});

app.delete('/api/projects/:id', (req, res) => {
  const index = db.projects.findIndex((project) => project.id === req.params.id);
  if (index === -1) return res.status(404).json({ error: 'Projeto não encontrado' });
  if (db.projects.length === 1) return res.status(400).json({ error: 'Não é possível excluir o último projeto' });
  db.projects.splice(index, 1);
  db.columns = db.columns.filter((column) => column.projectId !== req.params.id);
  db.tasks = db.tasks.filter((task) => task.projectId !== req.params.id);
  saveData();
  res.json({ ok: true });
});

app.get('/api/members', (req, res) => res.json(db.users.map(publicUser)));

app.get('/api/labels', (req, res) => res.json(db.labels));

app.get('/api/columns', (req, res) => {
  const columns = req.query.projectId ? db.columns.filter((column) => column.projectId === req.query.projectId) : db.columns;
  res.json(columns.sort((a, b) => a.order - b.order));
});

app.get('/api/tasks', (req, res) => {
  let tasks = [...db.tasks];
  if (req.query.projectId) tasks = tasks.filter((task) => task.projectId === req.query.projectId);
  if (req.query.assigneeId) tasks = tasks.filter((task) => task.assigneeId === req.query.assigneeId);
  res.json(tasks.sort((a, b) => a.order - b.order));
});

app.post('/api/tasks', (req, res) => {
  const task = normalizeTask(req.body);
  db.tasks.push(task);
  saveData();
  res.status(201).json(task);
});

app.put('/api/tasks/:id', (req, res) => {
  const index = db.tasks.findIndex((task) => task.id === req.params.id);
  if (index === -1) return res.status(404).json({ error: 'Tarefa não encontrada' });
  const current = db.tasks[index];
  const merged = normalizeTask({ ...current, ...req.body }, current.id);
  merged.createdAt = current.createdAt;
  db.tasks[index] = merged;
  saveData();
  res.json(merged);
});

app.patch('/api/tasks/:id', (req, res) => {
  const index = db.tasks.findIndex((task) => task.id === req.params.id);
  if (index === -1) return res.status(404).json({ error: 'Tarefa não encontrada' });
  const current = db.tasks[index];
  db.tasks[index] = normalizeTask({ ...current, ...req.body }, current.id);
  db.tasks[index].createdAt = current.createdAt;
  saveData();
  res.json(db.tasks[index]);
});

app.delete('/api/tasks/:id', (req, res) => {
  const before = db.tasks.length;
  db.tasks = db.tasks.filter((task) => task.id !== req.params.id);
  if (db.tasks.length === before) return res.status(404).json({ error: 'Tarefa não encontrada' });
  saveData();
  res.json({ ok: true });
});

app.post('/api/tasks/reorder', (req, res) => {
  const { taskId, columnId, index } = req.body;
  const task = db.tasks.find((item) => item.id === taskId);
  const column = getColumn(columnId);
  if (!task || !column || task.projectId !== column.projectId) return res.status(400).json({ error: 'Parâmetros inválidos' });
  const columnTasks = db.tasks.filter((item) => item.columnId === columnId && item.id !== taskId).sort((a, b) => a.order - b.order);
  const position = Math.max(0, Math.min(Number.isFinite(+index) ? +index : columnTasks.length, columnTasks.length));
  columnTasks.splice(position, 0, task);
  columnTasks.forEach((item, order) => {
    item.columnId = columnId;
    item.order = order;
    item.updatedAt = new Date().toISOString();
  });
  task.columnId = columnId;
  task.completedAt = column.title.toLowerCase().includes('conclu') ? task.completedAt || new Date().toISOString() : null;
  saveData();
  res.json({ ok: true });
});

app.get('/api/notes', (req, res) => res.json(db.notes));

app.post('/api/notes', (req, res) => {
  const now = new Date().toISOString();
  const note = {
    id: uid(),
    title: text(req.body.title),
    content: text(req.body.content),
    color: text(req.body.color, 'default'),
    pinned: !!req.body.pinned,
    createdAt: now,
    updatedAt: now
  };
  db.notes.push(note);
  saveData();
  res.status(201).json(note);
});

app.put('/api/notes/:id', (req, res) => {
  const index = db.notes.findIndex((note) => note.id === req.params.id);
  if (index === -1) return res.status(404).json({ error: 'Nota não encontrada' });
  const current = db.notes[index];
  db.notes[index] = {
    ...current,
    title: text(req.body.title, current.title),
    content: text(req.body.content, current.content),
    color: text(req.body.color, current.color),
    pinned: req.body.pinned === undefined ? current.pinned : !!req.body.pinned,
    updatedAt: new Date().toISOString()
  };
  saveData();
  res.json(db.notes[index]);
});

app.delete('/api/notes/:id', (req, res) => {
  const before = db.notes.length;
  db.notes = db.notes.filter((note) => note.id !== req.params.id);
  if (db.notes.length === before) return res.status(404).json({ error: 'Nota não encontrada' });
  saveData();
  res.json({ ok: true });
});

app.get('*', (req, res) => {
  res.sendFile(path.join(__dirname, 'public', 'index.html'));
});

function startServer(port = PORT, remainingAttempts = 10) {
  const server = app.listen(port);
  server.once('listening', () => {
    const address = server.address();
    const activePort = typeof address === 'object' && address ? address.port : port;
    console.log(`TASKFLOW rodando em http://localhost:${activePort}`);
  });
  server.once('error', (error) => {
    if (error.code !== 'EADDRINUSE' || process.env.PORT || remainingAttempts <= 1) {
      console.error(`Não foi possível iniciar o Taskflow na porta ${port}: ${error.message}`);
      process.exitCode = 1;
      return;
    }
    const nextPort = port + 1;
    console.warn(`Porta ${port} ocupada. Tentando a porta ${nextPort}...`);
    startServer(nextPort, remainingAttempts - 1);
  });
  return server;
}

if (require.main === module) {
  startServer();
}

module.exports = app;
