// Agent-Town 视觉设计系统
// 基于 "Agent Valley Web — 视觉呈现规范"

// UI 颜色系统（深色 HUD 风格）
export const uiColors = {
  // 背景
  background: '#1a1a2e',      // 深青蓝底
  backgroundDark: '#0f0f1a',  // 更深背景
  
  // 强调色
  primary: '#4fc3f7',         // 青色
  primaryHover: '#29b6f6',    // 悬停态
  primaryGlow: 'rgba(79, 195, 247, 0.3)', // 光晕
  
  // 文字
  textPrimary: '#e0e0e0',     // 主文字
  textSecondary: '#888888',   // 次要文字
  textMuted: '#555555',       // 辅助文字
  
  // 分割线
  border: '#333333',
  borderLight: '#444444',
  borderAccent: 'rgba(79, 195, 247, 0.2)',
  
  // 状态色
  success: '#4caf50',
  danger: '#f44336',
  dangerDark: '#c62828',
  warning: '#ff9800',         // 经济/资源橙
  
  // 特殊
  gold: '#ffd700',            // 房产证等高价值
};

// 3D 世界颜色（温暖自然色调）
export const worldColors = {
  // 地形
  grass: '#5a8f63',           // 草地 - 自然绿
  road: '#c8b08a',            // 路 - 暖土色
  water: '#4a92b8',           // 水 - 清澈蓝
  farmland: '#9a7820',        // 农田 - 金褐色
  sand: '#d4c088',            // 沙地
  foundation: '#6a6a6a',      // 地基
  hill: '#4a7a4a',            // 山丘 - 深绿
  
  // 资源
  tree: '#3d6b3d',            // 树木
  ore: '#708090',             // 矿产
};

// 建筑颜色映射
export const buildingColors: Record<string, string> = {
  town_hall: '#8b7355',   // 棕色 - 市政厅
  bank: '#4a90d9',        // 冷蓝 - 银行
  quest_board: '#50c878', // 清爽绿 - 任务中心
  shop: '#e8a87c',        // 暖橙 - 商店
};

// Agent 颜色（用于区分不同 Agent）
export const agentColors = [
  '#ff6b6b', // 红
  '#4ecdc4', // 青
  '#ffe66d', // 黄
  '#a8e6cf', // 薄荷绿
  '#ff8b94', // 粉红
  '#c7ceea', // 淡紫
  '#ffd3b6', // 桃色
];

// 间距系统
export const spacing = {
  xs: '4px',
  sm: '8px',
  md: '12px',
  lg: '16px',
  xl: '20px',
  xxl: '24px',
  xxxl: '32px',
};

// 圆角系统
export const borderRadius = {
  sm: '4px',
  md: '8px',
  lg: '12px',
  xl: '16px',
  pill: '9999px',
};

// 动效
export const transitions = {
  fast: '0.15s ease',
  normal: '0.25s ease',
  slow: '0.3s ease',
};

// 毛玻璃效果 CSS
export const glassStyle = {
  background: 'rgba(26, 26, 46, 0.9)',
  backdropFilter: 'blur(12px)',
  WebkitBackdropFilter: 'blur(12px)',
  border: `1px solid ${uiColors.borderAccent}`,
  borderRadius: borderRadius.lg,
  boxShadow: '0 4px 24px rgba(0, 0, 0, 0.4)',
};

// HUD 面板样式
export const hudPanelStyle = {
  ...glassStyle,
  color: uiColors.textPrimary,
  padding: spacing.lg,
};

// 按钮样式
export const buttonStyle = {
  background: 'rgba(79, 195, 247, 0.1)',
  border: `1px solid ${uiColors.primary}`,
  color: uiColors.primary,
  padding: `${spacing.sm} ${spacing.lg}`,
  borderRadius: borderRadius.md,
  cursor: 'pointer',
  transition: transitions.normal,
  ':hover': {
    background: uiColors.primary,
    color: uiColors.backgroundDark,
    transform: 'translateY(-2px)',
    boxShadow: `0 4px 12px ${uiColors.primaryGlow}`,
  },
};
