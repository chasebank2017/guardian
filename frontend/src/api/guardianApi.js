import axios from 'axios';

const API_BASE = '/v1'; // 代理到后端

export async function getMessages(agentId) {
  // 实际应为: `${API_BASE}/messages/${agentId}`
  // 这里用mock
  // const resp = await axios.get(`${API_BASE}/messages/${agentId}`);
  // return resp.data;
  return [
    { content: 'Hello from agent ' + agentId, timestamp: Date.now() },
    { content: 'Another message', timestamp: Date.now() - 10000 },
  ];
}

export async function getAgents() {
  // 实际应为: `${API_BASE}/agents`
  // 这里用mock
  return [
    { id: 1, name: 'Agent 1' },
    { id: 2, name: 'Agent 2' },
  ];
}
