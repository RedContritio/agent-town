import { create } from 'zustand';
import type { Agent, Todo, Skill, Event } from '@/types';

interface AgentState {
  // Current viewing agent
  currentAgent: Agent | null;
  todos: Todo[];
  skills: Skill[];
  pendingEvents: Event[];
  todoToken: string | null;
  
  // Actions
  setCurrentAgent: (agent: Agent | null) => void;
  setTodos: (todos: Todo[]) => void;
  addTodo: (todo: Todo) => void;
  updateTodo: (todo: Todo) => void;
  removeTodo: (todoId: string) => void;
  setSkills: (skills: Skill[]) => void;
  setPendingEvents: (events: Event[]) => void;
  addEvent: (event: Event) => void;
  setTodoToken: (token: string | null) => void;
}

export const useAgentStore = create<AgentState>((set) => ({
  currentAgent: null,
  todos: [],
  skills: [],
  pendingEvents: [],
  todoToken: null,
  
  setCurrentAgent: (agent) => set({ currentAgent: agent }),
  setTodos: (todos) => set({ todos }),
  addTodo: (todo) => set((state) => ({ todos: [...state.todos, todo] })),
  updateTodo: (todo) => set((state) => ({
    todos: state.todos.map((t) => t.id === todo.id ? todo : t)
  })),
  removeTodo: (todoId) => set((state) => ({
    todos: state.todos.filter((t) => t.id !== todoId)
  })),
  setSkills: (skills) => set({ skills }),
  setPendingEvents: (events) => set({ pendingEvents: events }),
  addEvent: (event) => set((state) => ({ 
    pendingEvents: [...state.pendingEvents, event] 
  })),
  setTodoToken: (token) => set({ todoToken: token }),
}));
