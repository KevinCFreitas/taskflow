(() => {
  'use strict';

  const ICONS = {
    menu: '<path d="M4 6h16M4 12h16M4 18h16"/>',
    'layout-dashboard': '<rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/>',
    folder: '<path d="M3 7.5A2.5 2.5 0 0 1 5.5 5H10l2 2h6.5A2.5 2.5 0 0 1 21 9.5v7A2.5 2.5 0 0 1 18.5 19h-13A2.5 2.5 0 0 1 3 16.5v-9Z"/>',
    'check-square': '<rect x="3" y="3" width="18" height="18" rx="3"/><path d="m7 12 3 3 7-7"/>',
    calendar: '<rect x="3" y="4.5" width="18" height="17" rx="2"/><path d="M16 2.5v4M8 2.5v4M3 9.5h18"/>',
    'sticky-note': '<path d="M5 3h14a2 2 0 0 1 2 2v10l-6 6H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2Z"/><path d="M15 21v-6h6M7 8h10M7 12h7"/>',
    settings: '<path d="M12 15.2a3.2 3.2 0 1 0 0-6.4 3.2 3.2 0 0 0 0 6.4Z"/><path d="m19.4 15 .1.1a2 2 0 0 1-2.8 2.8l-.1-.1a2 2 0 0 0-3.4 1.4v.2a2 2 0 0 1-4 0v-.2a2 2 0 0 0-3.4-1.4l-.1.1a2 2 0 0 1-2.8-2.8l.1-.1A2 2 0 0 0 3.7 11H3.5a2 2 0 0 1 0-4h.2a2 2 0 0 0 1.4-3.4L5 3.5A2 2 0 0 1 7.8.7l.1.1a2 2 0 0 0 3.4-1.4v-.2a2 2 0 0 1 4 0v.2a2 2 0 0 0 3.4 1.4l.1-.1A2 2 0 0 1 21.6 3l-.1.1A2 2 0 0 0 20.1 6h.2a2 2 0 0 1 0 4h-.2a2 2 0 0 0-1.4 3.4l.1.1" transform="translate(-.7 2) scale(.83)"/>',
    'more-horizontal': '<circle cx="5" cy="12" r="1" fill="currentColor" stroke="none"/><circle cx="12" cy="12" r="1" fill="currentColor" stroke="none"/><circle cx="19" cy="12" r="1" fill="currentColor" stroke="none"/>',
    search: '<circle cx="10.8" cy="10.8" r="6.8"/><path d="m16 16 5 5"/>',
    'sun-moon': '<path d="M15.5 3.5a8 8 0 1 0 5 11.8A7 7 0 0 1 15.5 3.5Z"/><path d="M5 3v3M3.5 4.5h3M19 17v3M17.5 18.5h3"/>',
    bell: '<path d="M18 9a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9M10 21h4"/>',
    'chevron-down': '<path d="m6 9 6 6 6-6"/>',
    'chevron-right': '<path d="m9 6 6 6-6 6"/>',
    'arrow-left': '<path d="m15 18-6-6 6-6M9 12h11"/>',
    'arrow-right': '<path d="M4 12h16m-6-6 6 6-6 6"/>',
    plus: '<path d="M12 5v14M5 12h14"/>',
    check: '<path d="m5 12 4 4L19 6"/>',
    'check-circle': '<circle cx="12" cy="12" r="9"/><path d="m8 12 2.7 2.7L16.5 9"/>',
    clock: '<circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/>',
    alert: '<path d="M10.3 4.1 2.1 18a2 2 0 0 0 1.7 3h16.4a2 2 0 0 0 1.7-3L13.7 4.1a2 2 0 0 0-3.4 0Z"/><path d="M12 9v4M12 17h.01"/>',
    target: '<circle cx="12" cy="12" r="9"/><circle cx="12" cy="12" r="5"/><circle cx="12" cy="12" r="1" fill="currentColor" stroke="none"/>',
    activity: '<path d="M3 12h4l2-7 4 14 2-7h6"/>',
    list: '<path d="M8 6h13M8 12h13M8 18h13M3 6h.01M3 12h.01M3 18h.01"/>',
    kanban: '<rect x="3" y="4" width="5" height="16" rx="1"/><rect x="10" y="4" width="5" height="10" rx="1"/><rect x="17" y="4" width="4" height="14" rx="1"/>',
    filter: '<path d="M4 5h16M7 12h10M10 19h4"/>',
    sliders: '<path d="M4 6h16M4 12h16M4 18h16"/><circle cx="8" cy="6" r="2"/><circle cx="15" cy="12" r="2"/><circle cx="11" cy="18" r="2"/>',
    edit: '<path d="M12 20h9"/><path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L8 18l-4 1 1-4Z"/>',
    trash: '<path d="M4 7h16M10 11v6M14 11v6M6 7l1 14h10l1-14M9 7V4h6v3"/>',
    copy: '<rect x="8" y="8" width="12" height="12" rx="2"/><path d="M16 8V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2h2"/>',
    x: '<path d="m6 6 12 12M18 6 6 18"/>',
    sparkle: '<path d="m12 3 1.3 5.7L19 10l-5.7 1.3L12 17l-1.3-5.7L5 10l5.7-1.3Z"/><path d="m19 16 .5 2.5L22 19l-2.5.5L19 22l-.5-2.5L16 19l2.5-.5Z"/>',
    inbox: '<path d="M4 5h16l2 10v4a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2v-4Z"/><path d="M2 15h5l2 3h6l2-3h5"/>',
    tag: '<path d="M20 13 13 20l-9-9V4h7Z"/><circle cx="8" cy="8" r="1" fill="currentColor" stroke="none"/>',
    user: '<circle cx="12" cy="8" r="3.5"/><path d="M4.5 21a7.5 7.5 0 0 1 15 0"/>',
    moon: '<path d="M20.5 15.4A8.5 8.5 0 0 1 8.6 3.5 8.5 8.5 0 1 0 20.5 15.4Z"/>',
    sun: '<circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4"/>',
    command: '<path d="M18 8a3 3 0 1 0-3-3v14a3 3 0 1 0 3-3H6a3 3 0 1 0 3 3V5a3 3 0 1 0-3 3Z"/>',
    logOut: '<path d="M10 17 15 12 10 7M15 12H3M21 4v16"/>',
    help: '<circle cx="12" cy="12" r="9"/><path d="M9.5 9a2.6 2.6 0 1 1 4.3 2c-1.2.9-1.8 1.3-1.8 3M12 17h.01"/>',
    keyboard: '<rect x="3" y="6" width="18" height="12" rx="2"/><path d="M7 10h.01M10 10h.01M13 10h.01M16 10h.01M7 14h10"/>',
    paperclip: '<path d="m20.5 11.5-8.7 8.7a5 5 0 0 1-7.1-7.1l8.7-8.7a3.5 3.5 0 0 1 5 5l-8.7 8.7a2 2 0 0 1-2.8-2.8l8.1-8.1"/>',
    message: '<path d="M20 11.5a7.5 7.5 0 0 1-8 7.5 8.7 8.7 0 0 1-3-.5L4 20l1.5-4A7.4 7.4 0 0 1 4 11.5 7.5 7.5 0 0 1 12 4a7.5 7.5 0 0 1 8 7.5Z"/>',
    rotate: '<path d="M20 11a8 8 0 0 0-14.7-3L3 11m0-5v5h5M4 13a8 8 0 0 0 14.7 3L21 13m0 5v-5h-5"/>',
    external: '<path d="M14 4h6v6M20 4l-9 9"/><path d="M18 13v5a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h5"/>',
    layers: '<path d="m12 3 9 5-9 5-9-5 9-5Z"/><path d="m3 12 9 5 9-5M3 16l9 5 9-5"/>',
    'more-vertical': '<circle cx="12" cy="5" r="1" fill="currentColor" stroke="none"/><circle cx="12" cy="12" r="1" fill="currentColor" stroke="none"/><circle cx="12" cy="19" r="1" fill="currentColor" stroke="none"/>',
    'refresh-cw': '<path d="M20 11a8 8 0 0 0-14.9-3L3 11m0-5v5h5M4 13a8 8 0 0 0 14.9 3L21 13m0 5v-5h-5"/>',
    'sort-asc': '<path d="M4 6h16M6 12h12M9 18h6"/>',
    'file-text': '<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8Z"/><path d="M14 2v6h6M8 13h8M8 17h6"/>'
  };

  const state = {
    data: null,
    view: 'dashboard',
    projectId: null,
    projectView: 'board',
    search: '',
    filters: { priority: 'all', assignee: 'all', status: 'all' },
    theme: readStorage('taskflow-theme') || 'light',
    modal: null,
    userMenu: false,
    draggingTask: null
  };

  const appContent = document.querySelector('#app-content');
  const modalRoot = document.querySelector('#modal-root');
  const toastRegion = document.querySelector('#toast-region');

  const api = {
    async request(url, options = {}) {
      const config = { ...options, headers: { ...(options.body ? { 'Content-Type': 'application/json' } : {}), ...(options.headers || {}) } };
      const response = await fetch(url, config);
      const payload = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(payload.error || 'Não foi possível concluir a operação.');
      return payload;
    },
    get(url) { return this.request(url); },
    post(url, body) { return this.request(url, { method: 'POST', body: JSON.stringify(body) }); },
    put(url, body) { return this.request(url, { method: 'PUT', body: JSON.stringify(body) }); },
    patch(url, body) { return this.request(url, { method: 'PATCH', body: JSON.stringify(body) }); },
    del(url) { return this.request(url, { method: 'DELETE' }); }
  };

  function readStorage(key) {
    try { return localStorage.getItem(key); } catch (error) { return null; }
  }

  function writeStorage(key, value) {
    try { localStorage.setItem(key, value); } catch (error) { return null; }
  }

  function icon(name, size = 16) {
    return `<svg width="${size}" height="${size}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${ICONS[name] || ICONS.sparkle}</svg>`;
  }

  function hydrateIcons(root = document) {
    root.querySelectorAll('[data-icon]').forEach((element) => {
      element.innerHTML = icon(element.dataset.icon, Number(element.dataset.iconSize) || 16);
    });
  }

  function escapeHtml(value) {
    return String(value ?? '').replace(/[&<>"']/g, (character) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[character]));
  }

  function safeColor(value, fallback = '#3f6152') {
    return /^#[0-9a-f]{3,8}$/i.test(String(value || '')) ? String(value) : fallback;
  }

  function uid() {
    return `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
  }

  function currentUser() {
    return state.data?.currentUser || state.data?.members?.[0] || { id: 'user-kevin', name: 'Kevin', initials: 'K', color: '#3f6152' };
  }

  function getProject(projectId = state.projectId) {
    return state.data?.projects?.find((project) => project.id === projectId) || state.data?.projects?.[0];
  }

  function getColumn(columnId) {
    return state.data?.columns?.find((column) => column.id === columnId);
  }

  function getMember(memberId) {
    return state.data?.members?.find((member) => member.id === memberId) || currentUser();
  }

  function getTask(taskId) {
    return state.data?.tasks?.find((task) => task.id === taskId);
  }

  function projectTasks(projectId = state.projectId) {
    return (state.data?.tasks || []).filter((task) => task.projectId === projectId);
  }

  function projectColumns(projectId = state.projectId) {
    return (state.data?.columns || []).filter((column) => column.projectId === projectId).sort((a, b) => a.order - b.order);
  }

  function isDone(task) {
    return (getColumn(task.columnId)?.title || '').toLowerCase().includes('conclu');
  }

  function statusLabel(task) {
    return getColumn(task.columnId)?.title || 'Sem status';
  }

  function statusClass(taskOrColumn) {
    const columnId = typeof taskOrColumn === 'string' ? taskOrColumn : taskOrColumn?.columnId;
    const title = (getColumn(columnId)?.title || '').toLowerCase();
    if (title.includes('andamento')) return 'status-progress';
    if (title.includes('revis')) return 'status-review';
    if (title.includes('conclu')) return 'status-done';
    return 'status-todo';
  }

  const priorityLabels = { urgent: 'Urgente', high: 'Alta', medium: 'Média', low: 'Baixa' };

  function priorityLabel(priority) {
    return priorityLabels[priority] || 'Média';
  }

  function labelFor(value) {
    return state.data?.labels?.find((label) => label.id === value || label.name === value) || { name: value, color: '#3f6152' };
  }

  function labelStyle(value) {
    const label = labelFor(value);
    return `--label-color:${safeColor(label.color, '#3f6152')};--label-soft:${safeColor(label.color, '#3f6152')}18`;
  }

  function dateFromValue(value) {
    if (!value) return null;
    const parts = String(value).slice(0, 10).split('-').map(Number);
    if (parts.length !== 3 || parts.some((part) => !Number.isFinite(part))) return null;
    return new Date(parts[0], parts[1] - 1, parts[2], 12);
  }

  function isoDate(date) {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
  }

  function today() {
    const value = new Date();
    value.setHours(12, 0, 0, 0);
    return value;
  }

  function formatDate(value, options = { day: '2-digit', month: 'short' }) {
    const date = dateFromValue(value) || (value ? new Date(value) : null);
    if (!date || Number.isNaN(date.getTime())) return 'Sem prazo';
    return date.toLocaleDateString('pt-BR', options);
  }

  function dueLabel(value) {
    const date = dateFromValue(value);
    if (!date) return 'Sem prazo';
    const todayValue = today();
    const tomorrow = new Date(todayValue);
    tomorrow.setDate(tomorrow.getDate() + 1);
    const yesterday = new Date(todayValue);
    yesterday.setDate(yesterday.getDate() - 1);
    if (isoDate(date) === isoDate(todayValue)) return 'Hoje';
    if (isoDate(date) === isoDate(tomorrow)) return 'Amanhã';
    if (isoDate(date) === isoDate(yesterday)) return 'Ontem';
    return formatDate(value);
  }

  function isOverdue(value) {
    const date = dateFromValue(value);
    return !!date && date < today();
  }

  function isToday(value) {
    const date = dateFromValue(value);
    return !!date && isoDate(date) === isoDate(today());
  }

  function formatMonthDay(date) {
    return date.toLocaleDateString('pt-BR', { month: 'short' }).replace('.', '').slice(0, 3);
  }

  function matchesSearch(task, query) {
    if (!query) return true;
    const normalized = query.toLowerCase();
    return [task.title, task.description, ...(task.labels || [])].some((value) => String(value || '').toLowerCase().includes(normalized));
  }

  function filteredTasks(projectId = state.projectId) {
    const tasks = projectId ? projectTasks(projectId) : (state.data?.tasks || []);
    return tasks.filter((task) => {
      if (!matchesSearch(task, state.search)) return false;
      if (state.filters.priority !== 'all' && task.priority !== state.filters.priority) return false;
      if (state.filters.assignee !== 'all' && task.assigneeId !== state.filters.assignee) return false;
      if (state.filters.status !== 'all' && task.columnId !== state.filters.status) return false;
      return true;
    });
  }

  function taskStats() {
    const tasks = state.data?.tasks || [];
    return {
      total: tasks.length,
      progress: tasks.filter((task) => statusClass(task) === 'status-progress').length,
      done: tasks.filter((task) => isDone(task)).length,
      overdue: tasks.filter((task) => isOverdue(task.dueDate) && !isDone(task)).length
    };
  }

  function projectProgress(projectId) {
    const tasks = projectTasks(projectId);
    if (!tasks.length) return 0;
    return Math.round((tasks.filter((task) => isDone(task)).length / tasks.length) * 100);
  }

  function projectNextDate(projectId) {
    return projectTasks(projectId).filter((task) => !isDone(task) && task.dueDate).sort((a, b) => String(a.dueDate).localeCompare(String(b.dueDate)))[0]?.dueDate;
  }

  function render() {
    if (!state.data) return;
    document.documentElement.dataset.theme = state.theme === 'system' ? (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light') : state.theme;
    renderShell();
    renderPage();
    renderModal();
  }

  function renderShell() {
    const user = currentUser();
    const workspace = state.data.workspace;
    const project = getProject();
    const meta = pageMeta();
    document.querySelector('#workspace-name').textContent = workspace?.name || 'Meu workspace';
    document.querySelector('#sidebar-user-name').textContent = user.name || 'Kevin';
    document.querySelector('#sidebar-user-email').textContent = user.email || 'kevin@taskflow.local';
    document.querySelector('#topbar-user-name').textContent = user.name || 'Kevin';
    document.querySelector('#sidebar-avatar').textContent = user.initials || 'K';
    document.querySelector('#topbar-avatar').textContent = user.initials || 'K';
    document.querySelector('#project-count').textContent = state.data.projects.length;
    document.querySelector('#notes-count').textContent = state.data.notes.length;
    document.querySelector('#page-title').textContent = meta.title;
    document.querySelector('#page-eyebrow').textContent = meta.eyebrow;
    document.querySelectorAll('.nav-item[data-view]').forEach((item) => item.classList.toggle('active', item.dataset.view === state.view));
    const projectNav = document.querySelector('#project-nav');
    projectNav.innerHTML = state.data.projects.map((item) => `<button class="project-nav-item ${item.id === state.projectId && ['board', 'table'].includes(state.view) ? 'active' : ''}" data-project-id="${escapeHtml(item.id)}" type="button"><i class="project-dot" style="background:${safeColor(item.color)}"></i><span>${escapeHtml(item.name)}</span></button>`).join('');
    document.querySelector('#user-popover')?.remove();
    if (state.userMenu) renderUserMenu();
    if (project && ['board', 'table'].includes(state.view)) {
      document.querySelector('#page-eyebrow').textContent = 'Projeto / visualização';
    }
  }

  function pageMeta() {
    const project = getProject();
    if (state.view === 'dashboard') return { title: 'Visão geral', eyebrow: 'Seu espaço de trabalho' };
    if (state.view === 'projects') return { title: 'Projetos', eyebrow: 'Organize seus espaços' };
    if (state.view === 'my-tasks') return { title: 'Minhas tarefas', eyebrow: 'Tudo que depende de você' };
    if (state.view === 'calendar') return { title: 'Calendário', eyebrow: 'Planejamento por prazo' };
    if (state.view === 'notes') return { title: 'Notas', eyebrow: 'Ideias e lembretes' };
    if (state.view === 'settings') return { title: 'Configurações', eyebrow: 'Preferências do workspace' };
    return { title: project?.name || 'Projeto', eyebrow: state.projectView === 'table' ? 'Projeto / tabela' : 'Projeto / quadro' };
  }

  function renderPage() {
    const renderers = {
      dashboard: renderDashboard,
      projects: renderProjects,
      'my-tasks': renderMyTasks,
      board: renderProjectView,
      table: renderProjectView,
      calendar: renderCalendar,
      notes: renderNotes,
      settings: renderSettings
    };
    appContent.innerHTML = (renderers[state.view] || renderDashboard)();
    hydrateIcons(appContent);
  }

  function statusPill(task) {
    return `<span class="status-pill ${statusClass(task)}">${escapeHtml(statusLabel(task))}</span>`;
  }

  function priorityPill(priority) {
    return `<span class="priority-label"><i class="priority-dot ${escapeHtml(priority)}"></i>${escapeHtml(priorityLabel(priority))}</span>`;
  }

  function dueMarkup(value, task = null) {
    if (!value) return '<span class="due-date">Sem prazo</span>';
    const overdue = isOverdue(value) && (!task || !isDone(task));
    const classes = `${overdue ? ' overdue' : ''}${isToday(value) ? ' today' : ''}`;
    return `<span class="due-date${classes}">${icon('clock', 11)}${escapeHtml(dueLabel(value))}</span>`;
  }

  function avatarMarkup(member, size = '') {
    const user = getMember(member?.id || member);
    return `<span class="mini-avatar ${size}" style="--avatar-color:${safeColor(user.color, '#3f6152')}" title="${escapeHtml(user.name || 'Membro')}">${escapeHtml(user.initials || String(user.name || 'K').slice(0, 1).toUpperCase())}</span>`;
  }

  function taskRow(task) {
    const completed = isDone(task);
    return `<div class="task-row ${completed ? 'completed' : ''}" data-action="open-task" data-task-id="${escapeHtml(task.id)}"><button class="task-check ${completed ? 'checked' : ''}" data-action="toggle-task" data-task-id="${escapeHtml(task.id)}" type="button" aria-label="${completed ? 'Reabrir' : 'Concluir'} tarefa">${icon('check', 12)}</button><div class="task-row-main"><span class="task-row-title">${escapeHtml(task.title)}</span><span class="task-row-meta">${statusPill(task)}${priorityPill(task.priority)}</span></div>${dueMarkup(task.dueDate, task)}${avatarMarkup(task.assigneeId)}</div>`;
  }

  function projectCard(project) {
    const tasks = projectTasks(project.id);
    const done = tasks.filter((task) => isDone(task)).length;
    const nextDate = projectNextDate(project.id);
    const color = safeColor(project.color);
    return `<article class="project-card" style="--project-color:${color}" data-action="open-project" data-project-id="${escapeHtml(project.id)}"><div class="project-card-top"><span class="project-icon">${escapeHtml(project.icon || '✦')}</span><button class="project-card-menu" data-action="project-menu" data-project-id="${escapeHtml(project.id)}" type="button" aria-label="Opções do projeto">${icon('more-horizontal', 15)}</button></div><h3>${escapeHtml(project.name)}</h3><p class="project-description">${escapeHtml(project.description || 'Sem descrição adicionada.')}</p><div class="progress-track"><div class="progress-fill" style="width:${projectProgress(project.id)}%"></div></div><div class="project-card-footer"><span>${done} de ${tasks.length} concluídas</span><span>${nextDate ? `Próximo: ${escapeHtml(dueLabel(nextDate))}` : 'Sem próximos prazos'}</span></div></article>`;
  }

  function activityMarkup() {
    const tasks = [...(state.data?.tasks || [])].sort((a, b) => String(b.updatedAt || '').localeCompare(String(a.updatedAt || ''))).slice(0, 4);
    if (!tasks.length) return '<div class="empty-inline">As atividades do projeto aparecerão aqui.</div>';
    return tasks.map((task) => {
      const completed = isDone(task);
      return `<div class="activity-item"><span class="activity-icon">${icon(completed ? 'check-circle' : 'activity', 15)}</span><div class="activity-copy"><strong>${escapeHtml(currentUser().name || 'Você')}</strong> ${completed ? 'concluiu' : 'atualizou'} <strong>${escapeHtml(task.title)}</strong><time>${escapeHtml(formatDate(task.updatedAt || task.createdAt, { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' }))}</time></div></div>`;
    }).join('');
  }

  function renderDashboard() {
    const stats = taskStats();
    const myTasks = (state.data.tasks || []).filter((task) => task.assigneeId === currentUser().id).filter((task) => !isDone(task)).sort((a, b) => String(a.dueDate || '9999').localeCompare(String(b.dueDate || '9999'))).slice(0, 5);
    const deadlines = (state.data.tasks || []).filter((task) => !isDone(task) && task.dueDate).sort((a, b) => String(a.dueDate).localeCompare(String(b.dueDate))).slice(0, 4);
    const percent = stats.total ? Math.round((stats.done / stats.total) * 100) : 0;
    const summary = `<section class="summary">
      <div class="summary-lead">
        <span class="summary-label">Progresso do workspace</span>
        <strong class="summary-value">${percent}<span class="summary-unit">%</span></strong>
        <div class="summary-bar" role="img" aria-label="${percent}% das tarefas concluídas"><span style="width:${percent}%"></span></div>
        <span class="summary-note">${stats.done} de ${stats.total} ${stats.total === 1 ? 'tarefa' : 'tarefas'} concluída${stats.total === 1 ? '' : 's'}</span>
      </div>
      <dl class="summary-figures">
        <div class="summary-figure"><dt>Em andamento</dt><dd>${stats.progress}</dd></div>
        <div class="summary-figure${stats.overdue ? ' is-alert' : ''}"><dt>Atrasadas</dt><dd>${stats.overdue}</dd></div>
        <div class="summary-figure"><dt>Projetos</dt><dd>${(state.data.projects || []).length}</dd></div>
      </dl>
    </section>`;
    return `<div class="page"><div class="page-intro"><div><span class="eyebrow">${greeting()}</span><h2>Olá, ${escapeHtml(currentUser().name || 'Kevin')}</h2><p>Acompanhe o que importa e deixe o resto respirar.</p></div><button class="button primary" data-action="new-task" type="button">${icon('plus', 15)} Nova tarefa <kbd>C</kbd></button></div>${summary}<div class="dashboard-grid"><section class="panel"><div class="panel-header"><div><h2 class="panel-title">Minhas tarefas</h2><p class="panel-subtitle">Próximas ações do seu dia</p></div><button class="text-button" data-view="my-tasks" type="button">Ver todas ${icon('arrow-right', 13)}</button></div><div class="task-list">${myTasks.length ? myTasks.map(taskRow).join('') : '<div class="empty-inline">Você está em dia. Crie uma tarefa para continuar.</div>'}</div></section><section class="panel"><div class="panel-header"><div><h2 class="panel-title">Próximos prazos</h2><p class="panel-subtitle">Não deixe nada para depois</p></div><button class="text-button" data-view="calendar" type="button">Calendário ${icon('arrow-right', 13)}</button></div><div class="deadline-list">${deadlines.length ? deadlines.map((task) => { const date = dateFromValue(task.dueDate); return `<button class="deadline-item" data-action="open-task" data-task-id="${escapeHtml(task.id)}" type="button"><span class="deadline-date"><strong>${date.getDate()}</strong><small>${formatMonthDay(date)}</small></span><span class="deadline-copy"><strong>${escapeHtml(task.title)}</strong><span>${escapeHtml(statusLabel(task))} · ${escapeHtml(dueLabel(task.dueDate))}</span></span>${priorityPill(task.priority)}</button>`; }).join('') : '<div class="empty-inline">Nenhum prazo próximo.</div>'}</div></section><section class="panel panel-wide"><div class="panel-header"><div><h2 class="panel-title">Seus projetos</h2><p class="panel-subtitle">Acompanhe o progresso de cada espaço</p></div><button class="text-button" data-view="projects" type="button">Ver projetos ${icon('arrow-right', 13)}</button></div><div class="project-grid">${state.data.projects.slice(0, 3).map(projectCard).join('')}</div></section><section class="panel panel-wide"><div class="panel-header"><div><h2 class="panel-title">Atividade recente</h2><p class="panel-subtitle">As últimas movimentações do workspace</p></div><span class="status-pill">${icon('rotate', 12)} Atualizado agora</span></div><div class="activity-list">${activityMarkup()}</div></section></div></div>`;
  }

  function renderProjects() {
    return `<div class="page"><div class="page-intro"><div><span class="eyebrow">Espaços de trabalho</span><h2>Seus projetos</h2><p>Separe cada contexto sem deixar a organização virar mais uma tarefa.</p></div><button class="button primary" data-action="new-project" type="button">${icon('plus', 15)} Novo projeto</button></div><div class="project-grid">${state.data.projects.map(projectCard).join('')}<button class="project-card project-card-new" data-action="new-project" type="button"><span class="project-icon">${icon('plus', 18)}</span><h3>Criar novo projeto</h3></button></div></div>`;
  }

  function renderMyTasks() {
    const tasks = (state.data.tasks || []).filter((task) => task.assigneeId === currentUser().id).filter((task) => matchesSearch(task, state.search));
    return `<div class="page"><div class="page-intro"><div><span class="eyebrow">Foco pessoal</span><h2>Minhas tarefas</h2><p>As tarefas atribuídas a você em todos os projetos.</p></div><button class="button primary" data-action="new-task" type="button">${icon('plus', 15)} Nova tarefa</button></div>${tasks.length ? `<div class="panel">${tasks.map(taskRow).join('')}</div>` : `<div class="empty-state"><span class="empty-state-icon">${icon('check-circle', 22)}</span><h3>Sua lista está limpa</h3><p>Você não possui tarefas atribuídas no momento.</p><button class="button primary" data-action="new-task" type="button">${icon('plus', 14)} Criar primeira tarefa</button></div>`}</div>`;
  }

  function renderProjectView() {
    const project = getProject();
    if (!project) return renderProjects();
    const tasks = filteredTasks(project.id);
    const columns = projectColumns(project.id);
    const tabs = `<div class="view-tabs"><button class="view-tab ${state.projectView === 'board' ? 'active' : ''}" data-action="set-project-view" data-project-view="board" type="button">${icon('kanban', 14)} Quadro</button><button class="view-tab ${state.projectView === 'table' ? 'active' : ''}" data-action="set-project-view" data-project-view="table" type="button">${icon('list', 14)} Tabela</button></div>`;
    const filters = `<div class="filter-bar"><select class="filter-select" data-filter="priority" aria-label="Filtrar por prioridade"><option value="all" ${state.filters.priority === 'all' ? 'selected' : ''}>Prioridade</option><option value="urgent" ${state.filters.priority === 'urgent' ? 'selected' : ''}>Urgente</option><option value="high" ${state.filters.priority === 'high' ? 'selected' : ''}>Alta</option><option value="medium" ${state.filters.priority === 'medium' ? 'selected' : ''}>Média</option><option value="low" ${state.filters.priority === 'low' ? 'selected' : ''}>Baixa</option></select><select class="filter-select" data-filter="assignee" aria-label="Filtrar por responsável"><option value="all" ${state.filters.assignee === 'all' ? 'selected' : ''}>Responsável</option>${state.data.members.map((member) => `<option value="${escapeHtml(member.id)}" ${state.filters.assignee === member.id ? 'selected' : ''}>${escapeHtml(member.name)}</option>`).join('')}</select><select class="filter-select" data-filter="status" aria-label="Filtrar por status"><option value="all" ${state.filters.status === 'all' ? 'selected' : ''}>Status</option>${columns.map((column) => `<option value="${escapeHtml(column.id)}" ${state.filters.status === column.id ? 'selected' : ''}>${escapeHtml(column.title)}</option>`).join('')}</select>${state.filters.priority !== 'all' || state.filters.assignee !== 'all' || state.filters.status !== 'all' ? `<button class="filter-chip" data-action="clear-filters" type="button">${icon('x', 12)} Limpar</button>` : ''}</div>`;
    return `<div class="page"><div class="project-hero"><button class="back-button" data-view="projects" type="button" aria-label="Voltar para projetos">${icon('arrow-left', 16)}</button><span class="project-hero-icon" style="--project-color:${safeColor(project.color)}">${escapeHtml(project.icon || '✦')}</span><div class="project-hero-copy"><h2>${escapeHtml(project.name)}</h2><p>${escapeHtml(project.description || 'Sem descrição adicionada.')}</p></div><div class="project-hero-actions"><button class="button secondary small" data-action="project-settings" data-project-id="${escapeHtml(project.id)}" type="button">${icon('settings', 13)} Editar</button><button class="button primary small" data-action="new-task" type="button">${icon('plus', 13)} Tarefa</button></div></div><div class="toolbar">${tabs}${filters}</div>${state.projectView === 'board' ? renderBoard(project, tasks, columns) : renderTable(project, tasks, columns)}</div>`;
  }

  function taskCard(task) {
    const completed = isDone(task);
    const labels = (task.labels || []).slice(0, 2);
    return `<article class="task-card ${completed ? 'completed' : ''}" draggable="true" data-drag-task data-task-id="${escapeHtml(task.id)}"><div class="task-card-top"><button class="task-check ${completed ? 'checked' : ''}" data-action="toggle-task" data-task-id="${escapeHtml(task.id)}" type="button" aria-label="${completed ? 'Reabrir' : 'Concluir'}">${icon('check', 11)}</button><span class="task-card-title" data-action="open-task" data-task-id="${escapeHtml(task.id)}">${escapeHtml(task.title)}</span><button class="task-card-menu" data-action="open-task" data-task-id="${escapeHtml(task.id)}" type="button" aria-label="Abrir tarefa">${icon('more-horizontal', 14)}</button></div>${task.description ? `<p class="task-card-description">${escapeHtml(task.description)}</p>` : ''}<div class="task-card-footer"><div class="task-card-meta">${dueMarkup(task.dueDate, task)}${labels.length ? `<span class="label-list">${labels.map((label) => `<span class="label" style="${labelStyle(label)}">${escapeHtml(labelFor(label).name)}</span>`).join('')}</span>` : ''}</div>${avatarMarkup(task.assigneeId)}</div></article>`;
  }

  function renderBoard(project, tasks, columns) {
    return `<div class="board">${columns.map((column) => { const columnTasks = tasks.filter((task) => task.columnId === column.id).sort((a, b) => a.order - b.order); return `<section class="board-column"><div class="board-column-header"><i class="column-status-dot" style="background:${safeColor(column.color, '#64748b')}"></i><span class="column-title">${escapeHtml(column.title)}</span><span class="column-count">${columnTasks.length}</span><button class="column-menu" data-action="new-task" data-column-id="${escapeHtml(column.id)}" type="button" aria-label="Adicionar tarefa">${icon('plus', 15)}</button></div><div class="column-dropzone" data-drop-column="${escapeHtml(column.id)}" data-project-id="${escapeHtml(project.id)}">${columnTasks.length ? columnTasks.map(taskCard).join('') : `<div class="empty-column">Solte tarefas aqui</div>`}</div></section>`; }).join('')}</div>`;
  }

  function renderTable(project, tasks, columns) {
    if (!tasks.length) return '<div class="empty-state"><span class="empty-state-icon">' + icon('inbox', 22) + '</span><h3>Nenhuma tarefa encontrada</h3><p>Ajuste os filtros ou crie uma nova tarefa para este projeto.</p><button class="button primary" data-action="new-task" type="button">' + icon('plus', 14) + ' Criar tarefa</button></div>';
    return `<div class="table-wrap"><table class="data-table"><thead><tr><th>Tarefa</th><th>Responsável</th><th>Prioridade</th><th>Prazo</th><th>Status</th><th></th></tr></thead><tbody>${tasks.map((task) => `<tr data-action="open-task" data-task-id="${escapeHtml(task.id)}"><td><div class="table-task-title"><button class="task-check ${isDone(task) ? 'checked' : ''}" data-action="toggle-task" data-task-id="${escapeHtml(task.id)}" type="button">${icon('check', 11)}</button><span>${escapeHtml(task.title)}</span></div></td><td><span class="table-assignee">${avatarMarkup(task.assigneeId)} ${escapeHtml(getMember(task.assigneeId).name || 'Sem responsável')}</span></td><td><select class="priority-select" data-table-field="priority" data-task-id="${escapeHtml(task.id)}" aria-label="Prioridade da tarefa"><option value="urgent" ${task.priority === 'urgent' ? 'selected' : ''}>Urgente</option><option value="high" ${task.priority === 'high' ? 'selected' : ''}>Alta</option><option value="medium" ${task.priority === 'medium' ? 'selected' : ''}>Média</option><option value="low" ${task.priority === 'low' ? 'selected' : ''}>Baixa</option></select></td><td class="inline-date">${dueMarkup(task.dueDate, task)}</td><td><select class="status-select" data-table-field="columnId" data-task-id="${escapeHtml(task.id)}" aria-label="Status da tarefa">${columns.map((column) => `<option value="${escapeHtml(column.id)}" ${column.id === task.columnId ? 'selected' : ''}>${escapeHtml(column.title)}</option>`).join('')}</select></td><td><button class="icon-button" data-action="open-task" data-task-id="${escapeHtml(task.id)}" type="button" aria-label="Abrir tarefa">${icon('chevron-right', 15)}</button></td></tr>`).join('')}</tbody></table></div>`;
  }

  function renderCalendar() {
    const start = today();
    const days = Array.from({ length: 7 }, (_, index) => { const date = new Date(start); date.setDate(start.getDate() + index); return date; });
    return `<div class="page"><div class="page-intro"><div><span class="eyebrow">Planejamento</span><h2>Calendário</h2><p>Uma visão rápida dos próximos sete dias.</p></div><button class="button primary" data-action="new-task" type="button">${icon('plus', 15)} Nova tarefa</button></div><div class="panel calendar-panel"><div class="calendar-week">${days.map((date) => { const dayTasks = (state.data.tasks || []).filter((task) => task.dueDate === isoDate(date) && !isDone(task)); return `<div class="calendar-day ${isoDate(date) === isoDate(start) ? 'today' : ''}"><div class="calendar-day-head"><span>${date.toLocaleDateString('pt-BR', { weekday: 'short' }).replace('.', '')}</span><strong>${date.getDate()}</strong></div>${dayTasks.length ? dayTasks.map((task) => `<button class="calendar-task" data-action="open-task" data-task-id="${escapeHtml(task.id)}" type="button">${escapeHtml(task.title)}</button>`).join('') : '<span class="empty-inline">—</span>'}</div>`; }).join('')}</div></div></div>`;
  }

  function renderNotes() {
    return `<div class="page"><div class="page-intro"><div><span class="eyebrow">Capture sem perder</span><h2>Notas</h2><p>Ideias, decisões e lembretes que merecem um lugar.</p></div><button class="button primary" data-action="new-note" type="button">${icon('plus', 15)} Nova nota</button></div>${state.data.notes.length ? `<div class="notes-grid">${state.data.notes.map((note) => { const color = safeColor(note.color === 'default' ? '#3f6152' : ({ blue: '#3c6377', green: '#2b6b4f', amber: '#905f18', purple: '#5b5170', rose: '#8a4a5c' }[note.color] || note.color)); return `<article class="note-card ${note.pinned ? 'pinned' : ''}" style="--note-color:${color}" data-action="open-note" data-note-id="${escapeHtml(note.id)}"><div class="note-card-head">${note.pinned ? `<span class="note-pin">${icon('sparkle', 14)}</span>` : ''}<h3>${escapeHtml(note.title || 'Sem título')}</h3><button class="icon-button" data-action="open-note" data-note-id="${escapeHtml(note.id)}" type="button" aria-label="Abrir nota">${icon('more-horizontal', 14)}</button></div><p>${escapeHtml(note.content || 'Nota sem conteúdo.')}</p><div class="note-card-footer"><span>Atualizada ${escapeHtml(formatDate(note.updatedAt || note.createdAt))}</span><span>${note.pinned ? 'Fixada' : 'Nota'}</span></div></article>`; }).join('')}</div>` : `<div class="empty-state"><span class="empty-state-icon">${icon('sticky-note', 22)}</span><h3>Nenhuma nota ainda</h3><p>Comece capturando uma ideia, decisão ou lembrete.</p><button class="button primary" data-action="new-note" type="button">${icon('plus', 14)} Criar primeira nota</button></div>`}</div>`;
  }

  function renderSettings() {
    return `<div class="page"><div class="page-intro"><div><span class="eyebrow">Personalização</span><h2>Configurações</h2><p>Deixe o Taskflow com a sua forma de trabalhar.</p></div></div><div class="settings-grid"><section class="settings-card"><h3>Aparência</h3><p>Escolha como o workspace deve aparecer para você.</p><div class="theme-options"><button class="theme-option ${state.theme === 'light' ? 'active' : ''}" data-action="set-theme" data-theme="light" type="button">${icon('sun', 18)} Claro</button><button class="theme-option ${state.theme === 'dark' ? 'active' : ''}" data-action="set-theme" data-theme="dark" type="button">${icon('moon', 18)} Escuro</button><button class="theme-option ${state.theme === 'system' ? 'active' : ''}" data-action="set-theme" data-theme="system" type="button">${icon('sun-moon', 18)} Sistema</button></div></section><section class="settings-card"><h3>Workspace</h3><p>Você está trabalhando neste espaço.</p><div class="workspace-summary"><span class="workspace-avatar">${escapeHtml((state.data.workspace?.name || 'P').slice(0, 1).toUpperCase())}</span><div><strong>${escapeHtml(state.data.workspace?.name || 'Meu workspace')}</strong><span>${escapeHtml(state.data.workspace?.description || 'Organização pessoal e profissional')}</span></div></div></section><section class="settings-card"><h3>Atalhos</h3><p>Atalhos para manter o ritmo.</p><div class="shortcut-list"><div class="shortcut-row"><span>Nova tarefa</span><span class="shortcut-key">C</span></div><div class="shortcut-row"><span>Pesquisa global</span><span class="shortcut-key">⌘ K</span></div><div class="shortcut-row"><span>Fechar janelas</span><span class="shortcut-key">Esc</span></div></div></section><section class="settings-card"><h3>Conta</h3><p>Seu perfil local neste workspace.</p><div class="workspace-summary"><span class="avatar avatar-topbar">${escapeHtml(currentUser().initials || 'K')}</span><div><strong>${escapeHtml(currentUser().name || 'Kevin')}</strong><span>${escapeHtml(currentUser().email || 'kevin@taskflow.local')}</span></div></div></section></div></div>`;
  }

  function renderUserMenu() {
    document.querySelector('#user-popover')?.remove();
    if (!state.userMenu) return;
    const popover = document.createElement('div');
    popover.className = 'user-popover';
    popover.id = 'user-popover';
    popover.innerHTML = `<button data-action="user-settings" type="button">${icon('settings', 15)} Preferências</button><button data-action="user-shortcuts" type="button">${icon('keyboard', 15)} Atalhos de teclado</button><button data-action="logout" type="button">${icon('logOut', 15)} Sair da conta</button>`;
    document.body.appendChild(popover);
  }

  function openModal(type, payload = {}) {
    state.modal = { type, ...payload };
    renderModal();
    setTimeout(() => modalRoot.querySelector('input:not([type="hidden"]), textarea, select')?.focus(), 40);
  }

  function closeModal() {
    state.modal = null;
    modalRoot.innerHTML = '';
  }

  function optionList(values, selected, valueKey = 'id', labelKey = 'name') {
    return values.map((value) => `<option value="${escapeHtml(value[valueKey])}" ${value[valueKey] === selected ? 'selected' : ''}>${escapeHtml(value[labelKey])}</option>`).join('');
  }

  function renderModal() {
    if (!state.modal) {
      modalRoot.innerHTML = '';
      return;
    }
    if (state.modal.type === 'task') modalRoot.innerHTML = taskModalTemplate(state.modal.task);
    if (state.modal.type === 'project') modalRoot.innerHTML = projectModalTemplate(state.modal.project);
    if (state.modal.type === 'note') modalRoot.innerHTML = noteModalTemplate(state.modal.note);
    if (state.modal.type === 'search') modalRoot.innerHTML = searchModalTemplate(state.modal.query || '');
    hydrateIcons(modalRoot);
  }

  function modalShell(title, subtitle, body, footer, form = '') {
    return `<div class="modal-backdrop" data-close-modal><section class="modal" data-modal-card>${form ? `<form data-form="${form}">` : ''}<div class="modal-header"><div><h2>${escapeHtml(title)}</h2>${subtitle ? `<p>${escapeHtml(subtitle)}</p>` : ''}</div><button class="icon-button" data-action="close-modal" type="button" aria-label="Fechar">${icon('x', 17)}</button></div><div class="modal-body">${body}</div>${footer ? `<div class="modal-footer">${footer}</div>` : ''}${form ? '</form>' : ''}</section></div>`;
  }

  function taskModalTemplate(task = null) {
    const current = task || { title: '', description: '', projectId: state.projectId || state.data.projects[0]?.id, columnId: projectColumns(task?.projectId || state.projectId)[0]?.id, priority: 'medium', assigneeId: currentUser().id, dueDate: '', labels: [] };
    const projectId = current.projectId || state.projectId || state.data.projects[0]?.id;
    const columns = projectColumns(projectId);
    const body = `<input type="hidden" name="id" value="${escapeHtml(current.id || '')}"><div class="form-grid"><label class="form-field full"><span class="form-label">Título <em>*</em></span><input class="form-control" name="title" required maxlength="120" placeholder="O que precisa ser feito?" value="${escapeHtml(current.title || '')}"></label><label class="form-field full"><span class="form-label">Descrição</span><textarea class="form-control" name="description" rows="3" placeholder="Adicione contexto para não esquecer o essencial...">${escapeHtml(current.description || '')}</textarea></label><label class="form-field"><span class="form-label">Projeto</span><select class="form-control" name="projectId">${optionList(state.data.projects, projectId)}</select></label><label class="form-field"><span class="form-label">Status</span><select class="form-control" name="columnId">${optionList(columns, current.columnId, 'id', 'title')}</select></label><label class="form-field"><span class="form-label">Responsável</span><select class="form-control" name="assigneeId">${optionList(state.data.members, current.assigneeId || currentUser().id)}</select></label><label class="form-field"><span class="form-label">Prioridade</span><select class="form-control" name="priority"><option value="urgent" ${current.priority === 'urgent' ? 'selected' : ''}>Urgente</option><option value="high" ${current.priority === 'high' ? 'selected' : ''}>Alta</option><option value="medium" ${current.priority === 'medium' ? 'selected' : ''}>Média</option><option value="low" ${current.priority === 'low' ? 'selected' : ''}>Baixa</option></select></label><label class="form-field"><span class="form-label">Prazo</span><input class="form-control" type="date" name="dueDate" value="${escapeHtml(String(current.dueDate || '').slice(0, 10))}"></label><label class="form-field"><span class="form-label">Etiquetas</span><input class="form-control" name="labels" placeholder="foco, urgente" value="${escapeHtml((current.labels || []).join(', '))}"></label></div>`;
    const footer = `${current.id ? `<button class="button danger danger-action" data-action="delete-task" data-task-id="${escapeHtml(current.id)}" type="button">${icon('trash', 13)} Excluir</button><button class="button ghost" data-action="duplicate-task" data-task-id="${escapeHtml(current.id)}" type="button">${icon('copy', 13)} Duplicar</button>` : ''}<button class="button secondary" data-action="close-modal" type="button">Cancelar</button><button class="button primary" type="submit">${current.id ? 'Salvar alterações' : 'Criar tarefa'}</button>`;
    return modalShell(current.id ? 'Editar tarefa' : 'Nova tarefa', current.id ? 'Atualize os detalhes e salve as mudanças.' : 'Adicione uma tarefa em poucos segundos.', body, footer, 'task');
  }

  function projectModalTemplate(project = null) {
    const current = project || { name: '', description: '', icon: '✦', color: '#3f6152' };
    const colors = ['#3f6152', '#3c6377', '#905f18', '#6b5b8f', '#8a4a5c', '#5a6472'];
    const icons = ['✦', '↗', '◈', '⌁', '◎', '◇', '＋'];
    const body = `<input type="hidden" name="id" value="${escapeHtml(current.id || '')}"><div class="form-grid"><label class="form-field full"><span class="form-label">Nome do projeto <em>*</em></span><input class="form-control" name="name" required maxlength="80" placeholder="Ex.: Lançamento do produto" value="${escapeHtml(current.name || '')}"></label><label class="form-field full"><span class="form-label">Descrição</span><textarea class="form-control" name="description" rows="3" placeholder="Qual é o objetivo deste projeto?">${escapeHtml(current.description || '')}</textarea></label><div class="form-field"><span class="form-label">Ícone</span><div class="icon-options">${icons.map((value) => `<button class="icon-option ${current.icon === value ? 'active' : ''}" data-select-field="icon" data-value="${value}" type="button">${value}</button>`).join('')}</div><input type="hidden" name="icon" value="${escapeHtml(current.icon || '✦')}"></div><div class="form-field"><span class="form-label">Cor</span><div class="color-options">${colors.map((value) => `<button class="color-option ${current.color === value ? 'active' : ''}" style="--color-value:${value}" data-select-field="color" data-value="${value}" type="button" aria-label="Cor ${value}"></button>`).join('')}</div><input type="hidden" name="color" value="${escapeHtml(current.color || '#3f6152')}"></div></div>`;
    const footer = `${current.id ? `<button class="button danger danger-action" data-action="delete-project" data-project-id="${escapeHtml(current.id)}" type="button">${icon('trash', 13)} Excluir</button>` : ''}<button class="button secondary" data-action="close-modal" type="button">Cancelar</button><button class="button primary" type="submit">${current.id ? 'Salvar projeto' : 'Criar projeto'}</button>`;
    return modalShell(current.id ? 'Editar projeto' : 'Novo projeto', current.id ? 'Atualize a identidade deste espaço.' : 'Crie um espaço para manter o contexto separado.', body, footer, 'project');
  }

  function noteModalTemplate(note = null) {
    const current = note || { title: '', content: '', color: 'default', pinned: false };
    const colors = { default: '#3f6152', blue: '#3c6377', green: '#2b6b4f', amber: '#905f18', purple: '#5b5170', rose: '#8a4a5c' };
    const colorKeys = Object.keys(colors);
    const body = `<input type="hidden" name="id" value="${escapeHtml(current.id || '')}"><input type="hidden" name="color" value="${escapeHtml(current.color || 'default')}"><div class="form-grid"><label class="form-field full"><span class="form-label">Título</span><input class="form-control" name="title" maxlength="100" placeholder="Dê um nome para a nota" value="${escapeHtml(current.title || '')}"></label><label class="form-field full"><span class="form-label">Conteúdo</span><textarea class="form-control" name="content" rows="7" placeholder="Escreva sua ideia...">${escapeHtml(current.content || '')}</textarea></label><div class="form-field full"><span class="form-label">Cor</span><div class="color-options">${colorKeys.map((key) => `<button class="color-option ${current.color === key ? 'active' : ''}" style="--color-value:${colors[key]}" data-select-field="color" data-value="${key}" type="button" aria-label="Cor ${key}"></button>`).join('')}</div></div><label class="form-field full"><span class="form-label">Fixar no topo</span><span class="workspace-summary"><input type="checkbox" name="pinned" ${current.pinned ? 'checked' : ''}> Manter esta nota sempre visível</span></label></div>`;
    const footer = `${current.id ? `<button class="button danger danger-action" data-action="delete-note" data-note-id="${escapeHtml(current.id)}" type="button">${icon('trash', 13)} Excluir</button>` : ''}<button class="button secondary" data-action="close-modal" type="button">Cancelar</button><button class="button primary" type="submit">${current.id ? 'Salvar nota' : 'Criar nota'}</button>`;
    return modalShell(current.id ? 'Editar nota' : 'Nova nota', 'Guarde uma ideia sem interromper o fluxo.', body, footer, 'note');
  }

  function searchModalTemplate(query) {
    const results = searchResults(query);
    return `<div class="modal-backdrop" data-close-modal><section class="modal search-modal" data-modal-card><div class="modal-header"><div><h2>Busca global</h2><p>Encontre tarefas e projetos rapidamente.</p></div><button class="icon-button" data-action="close-modal" type="button" aria-label="Fechar">${icon('x', 17)}</button></div><div class="modal-body"><div class="search-input-wrap">${icon('search', 17)}<input data-search-modal-input type="search" placeholder="Digite para buscar..." value="${escapeHtml(query)}" autocomplete="off"></div><div class="search-results" data-search-results>${results}</div></div></section></div>`;
  }

  function searchResults(query) {
    const normalized = query.trim().toLowerCase();
    if (!normalized) return '<div class="empty-inline">Digite para encontrar tarefas e projetos.</div>';
    const projects = state.data.projects.filter((project) => `${project.name} ${project.description}`.toLowerCase().includes(normalized)).slice(0, 4).map((project) => `<button class="search-result" data-action="open-project" data-project-id="${escapeHtml(project.id)}" type="button"><span class="search-result-icon">${icon('folder', 15)}</span><span class="search-result-copy"><strong>${escapeHtml(project.name)}</strong><span>Projeto · ${escapeHtml(project.description || 'Sem descrição')}</span></span>${icon('arrow-right', 14)}</button>`);
    const tasks = state.data.tasks.filter((task) => matchesSearch(task, query)).slice(0, 8).map((task) => `<button class="search-result" data-action="open-task" data-task-id="${escapeHtml(task.id)}" type="button"><span class="search-result-icon">${icon('check-square', 15)}</span><span class="search-result-copy"><strong>${escapeHtml(task.title)}</strong><span>${escapeHtml(statusLabel(task))} · ${escapeHtml(getProject(task.projectId)?.name || 'Projeto')}</span></span>${icon('arrow-right', 14)}</button>`);
    const all = [...projects, ...tasks];
    return all.length ? all.join('') : '<div class="empty-inline">Nenhum resultado encontrado.</div>';
  }

  function openTask(task = null, projectId = state.projectId, columnId = null) {
    const targetProject = projectId || task?.projectId || state.data.projects[0]?.id;
    const target = task ? { ...task } : { projectId: targetProject, columnId: columnId || projectColumns(targetProject)[0]?.id, priority: 'medium', assigneeId: currentUser().id, dueDate: '', labels: [] };
    openModal('task', { task: target });
  }

  function openProject(projectId) {
    const project = getProject(projectId);
    if (!project) return;
    state.projectId = project.id;
    state.view = 'board';
    state.projectView = 'board';
    state.filters = { priority: 'all', assignee: 'all', status: 'all' };
    closeModal();
    render();
  }

  function showToast(message, type = 'success') {
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.innerHTML = `<span class="toast-icon">${icon(type === 'error' ? 'alert' : 'check-circle', 15)}</span><span>${escapeHtml(message)}</span>`;
    toastRegion.appendChild(toast);
    setTimeout(() => toast.remove(), 3200);
  }

  async function refresh(message) {
    try {
      state.data = await api.get('/api/bootstrap');
      if (!state.projectId || !state.data.projects.some((project) => project.id === state.projectId)) state.projectId = state.data.projects[0]?.id || null;
      render();
      if (message) showToast(message);
    } catch (error) {
      appContent.innerHTML = `<div class="empty-state"><span class="empty-state-icon">${icon('alert', 22)}</span><h3>Não foi possível carregar</h3><p>${escapeHtml(error.message)}</p><button class="button primary" data-action="retry" type="button">${icon('refresh-cw', 14)} Tentar novamente</button></div>`;
      hydrateIcons(appContent);
    }
  }

  async function submitTask(form) {
    const formData = new FormData(form);
    const id = formData.get('id');
    const payload = { title: formData.get('title'), description: formData.get('description'), projectId: formData.get('projectId'), columnId: formData.get('columnId'), assigneeId: formData.get('assigneeId'), priority: formData.get('priority'), dueDate: formData.get('dueDate') || null, labels: String(formData.get('labels') || '').split(',').map((label) => label.trim()).filter(Boolean) };
    try {
      if (id) await api.put(`/api/tasks/${id}`, payload);
      else await api.post('/api/tasks', payload);
      closeModal();
      await refresh(id ? 'Tarefa atualizada' : 'Tarefa criada');
    } catch (error) {
      showToast(error.message, 'error');
    }
  }

  async function submitProject(form) {
    const formData = new FormData(form);
    const id = formData.get('id');
    const payload = { name: formData.get('name'), description: formData.get('description'), icon: formData.get('icon'), color: formData.get('color') };
    try {
      const project = id ? await api.put(`/api/projects/${id}`, payload) : await api.post('/api/projects', payload);
      closeModal();
      state.projectId = project.id;
      await refresh(id ? 'Projeto atualizado' : 'Projeto criado');
    } catch (error) {
      showToast(error.message, 'error');
    }
  }

  async function submitNote(form) {
    const formData = new FormData(form);
    const id = formData.get('id');
    const payload = { title: formData.get('title'), content: formData.get('content'), color: formData.get('color'), pinned: formData.get('pinned') === 'on' };
    try {
      if (id) await api.put(`/api/notes/${id}`, payload);
      else await api.post('/api/notes', payload);
      closeModal();
      await refresh(id ? 'Nota atualizada' : 'Nota criada');
    } catch (error) {
      showToast(error.message, 'error');
    }
  }

  async function toggleTask(taskId) {
    const task = getTask(taskId);
    if (!task) return;
    const targetColumn = isDone(task) ? projectColumns(task.projectId)[0] : projectColumns(task.projectId).find((column) => column.title.toLowerCase().includes('conclu')) || projectColumns(task.projectId).at(-1);
    try {
      await api.put(`/api/tasks/${taskId}`, { columnId: targetColumn.id });
      await refresh(isDone(task) ? 'Tarefa reaberta' : 'Tarefa concluída');
    } catch (error) {
      showToast(error.message, 'error');
    }
  }

  async function deleteTask(taskId) {
    const task = getTask(taskId);
    if (!task || !window.confirm(`Excluir a tarefa “${task.title}”?`)) return;
    try {
      await api.del(`/api/tasks/${taskId}`);
      closeModal();
      await refresh('Tarefa excluída');
    } catch (error) {
      showToast(error.message, 'error');
    }
  }

  async function duplicateTask(taskId) {
    const task = getTask(taskId);
    if (!task) return;
    try {
      await api.post('/api/tasks', { ...task, title: `Cópia de ${task.title}` });
      closeModal();
      await refresh('Tarefa duplicada');
    } catch (error) {
      showToast(error.message, 'error');
    }
  }

  async function deleteProject(projectId) {
    const project = getProject(projectId);
    if (!project || !window.confirm(`Excluir o projeto “${project.name}” e suas tarefas?`)) return;
    try {
      await api.del(`/api/projects/${projectId}`);
      closeModal();
      state.projectId = state.data.projects.find((item) => item.id !== projectId)?.id || null;
      await refresh('Projeto excluído');
    } catch (error) {
      showToast(error.message, 'error');
    }
  }

  async function deleteNote(noteId) {
    const note = state.data.notes.find((item) => item.id === noteId);
    if (!note || !window.confirm('Excluir esta nota?')) return;
    try {
      await api.del(`/api/notes/${noteId}`);
      closeModal();
      await refresh('Nota excluída');
    } catch (error) {
      showToast(error.message, 'error');
    }
  }

  function greeting() {
    const hour = new Date().getHours();
    if (hour < 12) return 'Bom dia';
    if (hour < 18) return 'Boa tarde';
    return 'Boa noite';
  }

  function setView(view) {
    if (!['dashboard', 'projects', 'my-tasks', 'calendar', 'notes', 'settings'].includes(view)) return;
    state.view = view;
    state.userMenu = false;
    closeSidebar();
    render();
  }

  function openSidebar() {
    document.querySelector('#sidebar').classList.add('open');
    document.querySelector('#sidebar-overlay').classList.add('open');
  }

  function closeSidebar() {
    document.querySelector('#sidebar').classList.remove('open');
    document.querySelector('#sidebar-overlay').classList.remove('open');
  }

  function applyTheme(theme) {
    const systemDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    state.theme = theme;
    writeStorage('taskflow-theme', theme);
    document.documentElement.dataset.theme = theme === 'system' ? (systemDark ? 'dark' : 'light') : theme;
  }

  function handleClick(event) {
    const closeTarget = event.target.closest('[data-close-modal]');
    if (closeTarget && !event.target.closest('[data-modal-card]')) {
      closeModal();
      return;
    }
    if (event.target.closest('select')) return;
    const viewTarget = event.target.closest('[data-view]');
    if (viewTarget) {
      setView(viewTarget.dataset.view);
      return;
    }
    const projectTarget = event.target.closest('[data-project-id]');
    const actionTarget = event.target.closest('[data-action]');
    if (actionTarget) {
      const action = actionTarget.dataset.action;
      if (action === 'close-modal') { closeModal(); return; }
      if (action === 'new-task') { openTask(null, actionTarget.dataset.projectId || state.projectId, actionTarget.dataset.columnId); return; }
      if (action === 'new-project') { openModal('project', { project: null }); return; }
      if (action === 'new-note') { openModal('note', { note: null }); return; }
      if (action === 'open-task') { openTask(getTask(actionTarget.dataset.taskId)); return; }
      if (action === 'open-note') { openModal('note', { note: state.data.notes.find((note) => note.id === actionTarget.dataset.noteId) }); return; }
      if (action === 'open-project') { openProject(actionTarget.dataset.projectId || projectTarget?.dataset.projectId); return; }
      if (action === 'project-settings') { openModal('project', { project: getProject(actionTarget.dataset.projectId) }); return; }
      if (action === 'project-menu') { openModal('project', { project: getProject(actionTarget.dataset.projectId) }); return; }
      if (action === 'set-project-view') { state.projectView = actionTarget.dataset.projectView; render(); return; }
      if (action === 'clear-filters') { state.filters = { priority: 'all', assignee: 'all', status: 'all' }; render(); return; }
      if (action === 'toggle-task') { toggleTask(actionTarget.dataset.taskId); return; }
      if (action === 'delete-task') { deleteTask(actionTarget.dataset.taskId); return; }
      if (action === 'duplicate-task') { duplicateTask(actionTarget.dataset.taskId); return; }
      if (action === 'delete-project') { deleteProject(actionTarget.dataset.projectId); return; }
      if (action === 'delete-note') { deleteNote(actionTarget.dataset.noteId); return; }
      if (action === 'set-theme') { applyTheme(actionTarget.dataset.theme); render(); return; }
      if (action === 'user-menu' || action === 'topbar-user') { state.userMenu = !state.userMenu; renderUserMenu(); return; }
      if (action === 'user-settings') { state.userMenu = false; setView('settings'); return; }
      if (action === 'user-shortcuts') { state.userMenu = false; openModal('search', { query: '' }); return; }
      if (action === 'logout') { state.userMenu = false; showToast('Sessão local mantida para demonstração'); return; }
      if (action === 'retry') { refresh(); return; }
    }
    if (projectTarget && !actionTarget) {
      openProject(projectTarget.dataset.projectId);
    }
  }

  function handleChange(event) {
    const filter = event.target.closest('[data-filter]');
    if (filter) {
      state.filters[filter.dataset.filter] = filter.value;
      render();
      return;
    }
    const field = event.target.closest('[data-table-field]');
    if (field) {
      const task = getTask(field.dataset.taskId);
      if (!task) return;
      api.put(`/api/tasks/${task.id}`, { [field.dataset.tableField]: field.value }).then(() => refresh('Tarefa atualizada')).catch((error) => showToast(error.message, 'error'));
      return;
    }
    const projectField = event.target.closest('[data-form="task"] [name="projectId"]');
    if (projectField) {
      const form = projectField.closest('[data-form="task"]');
      const statusField = form.elements.namedItem('columnId');
      const selectedTitle = statusField.selectedOptions[0]?.textContent || '';
      const columns = projectColumns(projectField.value);
      const selectedColumn = columns.find((column) => column.title === selectedTitle) || columns[0];
      statusField.innerHTML = optionList(columns, selectedColumn?.id, 'id', 'title');
    }
  }

  function handleInput(event) {
    if (event.target.matches('[data-search-modal-input]')) {
      state.modal.query = event.target.value;
      const results = modalRoot.querySelector('[data-search-results]');
      if (results) results.innerHTML = searchResults(event.target.value);
      return;
    }
    if (event.target.matches('[data-project-search]')) {
      state.search = event.target.value.trim();
      render();
      const input = document.querySelector('[data-project-search]');
      if (input) { input.focus(); input.setSelectionRange(input.value.length, input.value.length); }
    }
  }

  function handleKeydown(event) {
    const editing = ['INPUT', 'TEXTAREA', 'SELECT'].includes(document.activeElement?.tagName) || document.activeElement?.isContentEditable;
    if (event.key === 'Escape') {
      if (state.modal) closeModal();
      else if (state.userMenu) { state.userMenu = false; renderUserMenu(); }
      else closeSidebar();
      return;
    }
    if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
      event.preventDefault();
      openModal('search', { query: state.search });
      return;
    }
    if (!editing && !state.modal && event.key.toLowerCase() === 'c') {
      event.preventDefault();
      openTask(null, state.projectId);
    }
  }

  function handleDragStart(event) {
    const card = event.target.closest('[data-drag-task]');
    if (!card) return;
    state.draggingTask = card.dataset.taskId;
    card.classList.add('dragging');
    event.dataTransfer.effectAllowed = 'move';
    event.dataTransfer.setData('text/plain', card.dataset.taskId);
  }

  function handleDragOver(event) {
    const zone = event.target.closest('[data-drop-column]');
    if (!zone) return;
    event.preventDefault();
    zone.classList.add('drag-over');
  }

  function handleDragLeave(event) {
    const zone = event.target.closest('[data-drop-column]');
    if (zone && !zone.contains(event.relatedTarget)) zone.classList.remove('drag-over');
  }

  async function handleDrop(event) {
    const zone = event.target.closest('[data-drop-column]');
    if (!zone) return;
    event.preventDefault();
    zone.classList.remove('drag-over');
    const taskId = state.draggingTask || event.dataTransfer.getData('text/plain');
    const columnId = zone.dataset.dropColumn;
    if (!taskId) return;
    const cards = [...zone.querySelectorAll('[data-drag-task]')];
    const targetCard = event.target.closest('[data-drag-task]');
    const index = targetCard ? cards.indexOf(targetCard) : cards.length;
    try {
      await api.post('/api/tasks/reorder', { taskId, columnId, index });
      await refresh('Tarefa movida');
    } catch (error) {
      showToast(error.message, 'error');
      await refresh();
    }
  }

  function handleDragEnd() {
    state.draggingTask = null;
    document.querySelectorAll('.dragging, .drag-over').forEach((element) => element.classList.remove('dragging', 'drag-over'));
  }

  document.addEventListener('click', handleClick);
  document.addEventListener('change', handleChange);
  document.addEventListener('input', handleInput);
  document.addEventListener('keydown', handleKeydown);
  appContent.addEventListener('dragstart', handleDragStart);
  appContent.addEventListener('dragover', handleDragOver);
  appContent.addEventListener('dragleave', handleDragLeave);
  appContent.addEventListener('drop', handleDrop);
  appContent.addEventListener('dragend', handleDragEnd);
  document.querySelector('#menu-button').addEventListener('click', openSidebar);
  document.querySelector('#sidebar-overlay').addEventListener('click', closeSidebar);
  document.querySelector('#search-trigger').addEventListener('click', () => openModal('search', { query: state.search }));
  document.querySelector('#theme-button').addEventListener('click', () => applyTheme(state.theme === 'dark' ? 'light' : 'dark'));
  document.querySelector('#topbar-user').addEventListener('click', () => { state.userMenu = !state.userMenu; renderUserMenu(); });
  modalRoot.addEventListener('submit', (event) => {
    event.preventDefault();
    const form = event.target.closest('[data-form]');
    if (!form) return;
    if (form.dataset.form === 'task') submitTask(form);
    if (form.dataset.form === 'project') submitProject(form);
    if (form.dataset.form === 'note') submitNote(form);
  });
  modalRoot.addEventListener('click', (event) => {
    const select = event.target.closest('[data-select-field]');
    if (!select) return;
    const form = select.closest('[data-form]');
    if (!form) return;
    const input = form.querySelector(`[name="${select.dataset.selectField}"]`);
    if (input) input.value = select.dataset.value;
    form.querySelectorAll(`[data-select-field="${select.dataset.selectField}"]`).forEach((element) => element.classList.toggle('active', element === select));
  });

  hydrateIcons(document);
  applyTheme(state.theme);
  refresh();
})();
