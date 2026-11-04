const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');

const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'taskflow-test-'));
process.env.TASKFLOW_DATA_FILE = path.join(tempDir, 'nested', 'data.json');
const app = require('../server');

async function startServer() {
  const server = await new Promise((resolve) => {
    const instance = app.listen(0, () => resolve(instance));
  });
  const address = server.address();
  return { server, baseUrl: `http://127.0.0.1:${address.port}` };
}

async function request(baseUrl, route, method = 'GET', body) {
  const response = await fetch(`${baseUrl}${route}`, {
    method,
    headers: body ? { 'Content-Type': 'application/json' } : {},
    body: body ? JSON.stringify(body) : undefined
  });
  const payload = await response.json();
  return { status: response.status, payload };
}

test('API do Taskflow executa o ciclo principal de projeto e tarefa', async () => {
  const { server, baseUrl } = await startServer();
  try {
    const health = await request(baseUrl, '/api/health');
    assert.equal(health.status, 200);
    assert.equal(health.payload.service, 'taskflow');

    const bootstrap = await request(baseUrl, '/api/bootstrap');
    assert.equal(bootstrap.status, 200);
    assert.equal(bootstrap.payload.currentUser.name, 'Kevin');
    assert.ok(bootstrap.payload.projects.length >= 3);

    const projectResponse = await request(baseUrl, '/api/projects', 'POST', {
      name: 'Projeto de teste',
      description: 'Validação automatizada',
      icon: '◇',
      color: '#123456'
    });
    assert.equal(projectResponse.status, 201);
    const project = projectResponse.payload;

    const columnsResponse = await request(baseUrl, `/api/columns?projectId=${project.id}`);
    assert.equal(columnsResponse.status, 200);
    assert.equal(columnsResponse.payload.length, 4);

    const taskResponse = await request(baseUrl, '/api/tasks', 'POST', {
      title: 'Tarefa de teste',
      description: 'Validar persistência',
      projectId: project.id,
      priority: 'urgent',
      dueDate: '2030-01-10'
    });
    assert.equal(taskResponse.status, 201);
    const task = taskResponse.payload;
    assert.equal(task.projectId, project.id);
    assert.equal(task.priority, 'urgent');

    const updateResponse = await request(baseUrl, `/api/tasks/${task.id}`, 'PUT', {
      title: 'Tarefa atualizada',
      priority: 'high'
    });
    assert.equal(updateResponse.status, 200);
    assert.equal(updateResponse.payload.title, 'Tarefa atualizada');

    const progressColumn = columnsResponse.payload.find((column) => column.title.includes('andamento'));
    const reorderResponse = await request(baseUrl, '/api/tasks/reorder', 'POST', {
      taskId: task.id,
      columnId: progressColumn.id,
      index: 0
    });
    assert.equal(reorderResponse.status, 200);

    const deleteTaskResponse = await request(baseUrl, `/api/tasks/${task.id}`, 'DELETE');
    assert.equal(deleteTaskResponse.status, 200);
    const deleteProjectResponse = await request(baseUrl, `/api/projects/${project.id}`, 'DELETE');
    assert.equal(deleteProjectResponse.status, 200);
  } finally {
    await new Promise((resolve) => server.close(resolve));
    fs.rmSync(tempDir, { recursive: true, force: true });
  }
});
