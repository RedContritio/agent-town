import { create } from 'zustand';
import type { WorldInfo, WorldTime, VisibleArea, Agent, BuildingView, BlockView } from '@/types';

interface WorldState {
  // World data
  worldInfo: WorldInfo | null;
  worldTime: WorldTime | null;
  
  // View data
  visibleArea: VisibleArea | null;
  selectedAgent: Agent | null;
  selectedBuilding: BuildingView | null;
  selectedBlock: BlockView | null;
  
  // View settings
  followAgentId: string | null;
  showGrid: boolean;
  showAgents: boolean;
  showBuildings: boolean;
  
  // Actions
  setWorldInfo: (info: WorldInfo) => void;
  setWorldTime: (time: WorldTime) => void;
  setVisibleArea: (area: VisibleArea) => void;
  setSelectedAgent: (agent: Agent | null) => void;
  setSelectedBuilding: (building: BuildingView | null) => void;
  setSelectedBlock: (block: BlockView | null) => void;
  setFollowAgentId: (agentId: string | null) => void;
  toggleGrid: () => void;
  toggleAgents: () => void;
  toggleBuildings: () => void;
}

export const useWorldStore = create<WorldState>((set) => ({
  worldInfo: null,
  worldTime: null,
  visibleArea: null,
  selectedAgent: null,
  selectedBuilding: null,
  selectedBlock: null,
  followAgentId: null,
  showGrid: true,
  showAgents: true,
  showBuildings: true,
  
  setWorldInfo: (info) => set({ worldInfo: info }),
  setWorldTime: (time) => set({ worldTime: time }),
  setVisibleArea: (area) => set({ visibleArea: area }),
  setSelectedAgent: (agent) => set({ selectedAgent: agent }),
  setSelectedBuilding: (building) => set({ selectedBuilding: building }),
  setSelectedBlock: (block) => set({ selectedBlock: block }),
  setFollowAgentId: (agentId) => set({ followAgentId: agentId }),
  toggleGrid: () => set((state) => ({ showGrid: !state.showGrid })),
  toggleAgents: () => set((state) => ({ showAgents: !state.showAgents })),
  toggleBuildings: () => set((state) => ({ showBuildings: !state.showBuildings })),
}));
