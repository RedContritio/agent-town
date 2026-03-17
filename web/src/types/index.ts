// Agent types
export interface Agent {
  id: string;
  name: string;
  position: Position;
  facing: number;
  hp: number;
  maxHp: number;
  stamina: number;
  maxStamina: number;
  hunger: number;
  maxHunger: number;
  balance: number;
  isOnline: boolean;
  inBattle: boolean;
  battleId?: string;
}

export interface Position {
  x: number;
  y: number;
  z: number;
}

// World types
export interface BlockView {
  position: Position;
  height: number;
  terrainType: string;
  resourceType?: string;
  resourceAmount?: number;
}

export interface BuildingView {
  id: string;
  name: string;
  ownerId: string;
  anchor: Position;
  width: number;
  depth: number;
  height: number;
}

export interface MonsterView {
  id: string;
  name: string;
  type: string;
  position: Position;
}

export interface VisibleArea {
  agentId: string;
  center: Position;
  radius: number;
  blocks: BlockView[];
  agents: Agent[];
  buildings: BuildingView[];
  monsters: MonsterView[];
}

// Todo types
export interface Todo {
  id: string;
  content: string;
  status: 'pending' | 'planning' | 'completed' | 'rejected' | 'delayed';
  priority: number;
  createdAt: string;
  updatedAt: string;
  rejectReason?: string;
}

// Skill types
export interface Skill {
  type: string;
  level: number;
  exp: number;
  expToNext: number;
}

// Event types
export interface Event {
  id: string;
  type: string;
  data: unknown;
  createdAt: string;
}

// World info
export interface WorldInfo {
  id: string;
  name: string;
  seed: string;
  timeSpeed: number;
  currentTime: string;
  agentCount: number;
  buildingCount: number;
}

export interface WorldTime {
  timestamp: string;
  year: number;
  month: number;
  day: number;
  hour: number;
  minute: number;
  season: 'spring' | 'summer' | 'autumn' | 'winter';
  isDaytime: boolean;
}
