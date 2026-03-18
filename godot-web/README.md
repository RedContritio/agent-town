# Agent Town - Godot Web Frontend

使用 Godot 引擎开发的 3D 网页前端，替代原有的 React + Three.js 方案。

## 优势

- **原生 3D 渲染** - Godot 的 3D 引擎性能更好，渲染效果更稳定
- **强大的 UI 系统** - 使用 Godot 的 Control 节点创建真正的 2D UI 层
- **更好的相机控制** - 原生支持各种相机约束和交互
- **WebAssembly 导出** - 可以导出为网页直接运行

## 项目结构

```
godot-web/
├── project.godot          # Godot 项目配置
├── icon.svg               # 项目图标
├── export_presets.cfg     # 导出配置
├── scenes/                # 场景文件
│   ├── main.tscn          # 主场景
│   ├── entities/          # 实体场景
│   │   ├── agent.tscn     # Agent 角色
│   │   └── building.tscn  # 建筑
│   └── ui/                # UI 场景
│       └── hud.tscn       # HUD 界面
├── scripts/               # GDScript 脚本
│   ├── api_client.gd      # HTTP API 客户端
│   ├── world_manager.gd   # 世界管理器
│   ├── camera_controller.gd  # 相机控制
│   ├── ui/                # UI 脚本
│   │   └── hud.gd
│   └── entities/          # 实体脚本
│       ├── agent.gd
│       └── building.gd
└── assets/                # 资源文件
    ├── models/            # 3D 模型
    ├── textures/          # 贴图
    └── materials/         # 材质
```

## 开发环境配置

1. **安装 Godot 4.3+**
   ```bash
   # 下载 Godot 4.3
   wget https://downloads.tuxfamily.org/godotengine/4.3/Godot_v4.3-stable_linux.x86_64.tar.xz
   tar -xf Godot_v4.3-stable_linux.x86_64.tar.xz
   sudo mv Godot_v4.3-stable_linux.x86_64 /usr/local/bin/godot
   ```

2. **打开项目**
   ```bash
   cd godot-web
   godot project.godot
   ```

## 运行项目

1. **确保后端服务已启动**
   ```bash
   cd ..
   make dev-server  # 启动 Go 后端
   ```

2. **在 Godot 编辑器中运行**
   - 按 F5 或点击"运行项目"按钮

3. **导出 Web 版本**
   - 项目 -> 导出 -> Web
   - 导出到 `../server/cmd/server/web/`
   - 或者使用命令行:
   ```bash
   godot --headless --export-release "Web" ../server/cmd/server/web/index.html
   ```

## 功能特性

### 已实现

- [x] HTTP API 客户端（轮询后端数据）
- [x] 3D 地形渲染（使用 MultiMesh 优化性能）
- [x] Agent 角色显示（带动画和名字标签）
- [x] 建筑显示
- [x] 相机控制（旋转、平移、缩放）
- [x] 相机高度限制（不低于地面）
- [x] HUD 界面（Agent/建筑计数、图例）

### 待实现

- [ ] 点击 Agent/建筑显示详细信息
- [ ] 视野范围显示
- [ ] 地形编辑/交互
- [ ] WebSocket 实时通信
- [ ] 更好的加载界面
- [ ] 错误重连机制

## API 集成

前端通过 HTTP 轮询与后端通信：

```
GET /api/v1/world/info      -> 世界信息
GET /api/v1/world/map       -> 地图数据
GET /api/v1/agents          -> Agent 列表
GET /api/v1/agents/{id}     -> Agent 详情
```

## 相机控制

- **左键拖动**: 旋转视角（不能低于地面）
- **右键拖动**: 平移位置（可以移动到世界各处）
- **滚轮**: 缩放

## 与后端集成

导出 Web 版本后，Go 后端会自动提供静态文件服务：

```go
// server/cmd/server/main.go
fs := http.FileServer(http.Dir("./web"))
http.Handle("/", fs)
```

访问 `http://localhost:8080/` 即可看到 Godot 前端。

## 技术细节

### Label3D 配置

Agent 和建筑的名字使用 `Label3D` 节点：
- `billboard = 1` - 始终面向相机
- `fixed_size = true` - 大小不随距离变化
- `outline_size = 2` - 文字描边，提高可读性

### 性能优化

- 地形使用 `MultiMeshInstance3D` 批量渲染同类型方块
- 每 5 秒轮询一次数据，避免频繁请求
- 使用 `MultiMesh` 减少 Draw Call

## 许可证

MIT License
