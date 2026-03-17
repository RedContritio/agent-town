import axios from 'axios';
import type { 
  Agent, 
  VisibleArea, 
  Todo, 
  Skill, 
  WorldInfo, 
  WorldTime,
  Event 
} from '@/types';

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api/v1';

const client = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Agent API
export const agentApi = {
  getAgent: (agentId: string) => 
    client.get<Agent>(`/agents/${agentId}`),
  
  getAgentStatus: (agentId: string) => 
    client.get<{ agentId: string; worldTime: string; stamina: number; hunger: number; inBattle: boolean; pendingEvents: Event[] }>(`/agents/${agentId}/status`),
  
  getVisibleArea: (agentId: string) => 
    client.get<VisibleArea>(`/agents/${agentId}/visible-area`),
  
  getSkills: (agentId: string) => 
    client.get<{ agentId: string; skills: Skill[] }>(`/agents/${agentId}/skills`),
  
  getTodos: (agentId: string) => 
    client.get<{ agentId: string; todos: Todo[] }>(`/agents/${agentId}/todos`),
  
  createTodo: (agentId: string, token: string, content: string, priority: number = 5) => 
    client.post<Todo>(`/agents/${agentId}/todos`, { content, priority }, {
      headers: { 'X-Todo-Token': token }
    }),
  
  updateTodo: (agentId: string, token: string, todoId: string, content: string, priority: number) => 
    client.put<Todo>(`/agents/${agentId}/todos/${todoId}`, { content, priority }, {
      headers: { 'X-Todo-Token': token }
    }),
  
  deleteTodo: (agentId: string, token: string, todoId: string) => 
    client.delete(`/agents/${agentId}/todos/${todoId}`, {
      headers: { 'X-Todo-Token': token }
    }),
};

// World API
export const worldApi = {
  getWorldInfo: () => 
    client.get<WorldInfo>('/world/info'),
  
  getWorldTime: () => 
    client.get<WorldTime>('/world/time'),
  
  getMap: (centerX: number, centerY: number, radius: number) => 
    client.get('/world/map', { params: { x: centerX, y: centerY, radius } }),
  
  getBlock: (x: number, y: number) => 
    client.get('/world/block', { params: { x, y } }),
};

// Poll events (long polling)
export const pollEvents = async (agentId: string, timeout: number = 30): Promise<Event[]> => {
  const response = await client.get('/events/poll', { 
    params: { agentId, timeout },
    timeout: (timeout + 5) * 1000, // Add 5s buffer
  });
  return response.data.events;
};

export default client;
