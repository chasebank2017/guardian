import axios, { AxiosError, AxiosInstance } from 'axios';

export interface Agent {
  id: number;
  name: string;
}

export interface Message {
  content: string;
  timestamp: number;
}

const apiBaseUrl = (import.meta as any)?.env?.VITE_API_BASE || '';

const api: AxiosInstance = axios.create({
  baseURL: apiBaseUrl,
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers = config.headers || {};
    (config.headers as any)['Authorization'] = `Bearer ${token}`;
  }
  return config;
});

function toError(err: unknown): Error {
  const axErr = err as AxiosError<any>;
  const responseData = axErr?.response?.data;
  const message =
    typeof responseData === 'string'
      ? responseData
      : responseData?.message || axErr.message || '请求失败';
  return new Error(message);
}

export async function login(username: string, password: string): Promise<string> {
  try {
    const resp = await api.post('/login', { username, password });
    return (resp.data as { token: string }).token;
  } catch (e) {
    throw toError(e);
  }
}

export async function fetchAgents(): Promise<Agent[]> {
  try {
    const resp = await api.get('/v1/agents');
    return resp.data as Agent[];
  } catch (e) {
    throw toError(e);
  }
}

export async function fetchMessagesForAgent(agentId: number): Promise<Message[]> {
  try {
    const resp = await api.get(`/v1/agents/${agentId}/messages`);
    return resp.data as Message[];
  } catch (e) {
    throw toError(e);
  }
}

export default api;


