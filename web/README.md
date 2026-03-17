# Agent Town Web

纯前端 3D 世界观察器，使用 React + TypeScript + Three.js 构建。

## 技术栈

- **React 18** - UI 框架
- **TypeScript** - 类型安全
- **Three.js** - 3D 渲染
- **@react-three/fiber** - React Three.js 渲染器
- **@react-three/drei** - Three.js 辅助组件
- **Vite** - 构建工具
- **Zustand** - 状态管理
- **Axios** - HTTP 客户端

## 开发

```bash
# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 构建
npm run build

# 预览生产构建
npm run preview
```

## 项目结构

```
web/
├── src/
│   ├── components/    # React 组件
│   ├── pages/         # 页面组件
│   ├── stores/        # Zustand 状态管理
│   ├── api/           # API 客户端
│   ├── types/         # TypeScript 类型
│   └── utils/         # 工具函数
├── package.json
├── tsconfig.json
├── vite.config.ts
└── index.html
```

## API 连接

Web 通过 HTTP API 连接到 Server (Go 后端)：

- 开发环境: `http://localhost:8080`
- 生产环境: 通过环境变量配置

## 功能

- [ ] 3D 世界渲染
- [ ] Agent 观察
- [ ] TODO 管理
- [ ] 时间控制
- [ ] 快照回放
