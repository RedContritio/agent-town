class_name BuildingGenerator
extends RefCounted

# 建筑生成器基类 - 纪念碑谷风格
# 子类需要实现 generate() 方法

# 输入参数（由服务端数据决定）
var width: float = 3.0
var height: float = 3.0
var depth: float = 2.0

# 颜色配置（使用纪念碑谷风格调色板）
var primary_color: Color = MeshBuilder.PALETTE["wall_light"]
var secondary_color: Color = MeshBuilder.PALETTE["roof_terracotta"]
var accent_color: Color = MeshBuilder.PALETTE["accent_blue"]

# 设置尺寸
func set_size(w: float, h: float, d: float) -> void:
    width = w
    height = h
    depth = d

# 设置颜色
func set_colors(primary: Color, secondary: Color, accent: Color = Color.WHITE) -> void:
    primary_color = primary
    secondary_color = secondary
    accent_color = accent if accent != Color.WHITE else secondary.lightened(0.2)

# 子类必须实现此方法
func generate() -> ArrayMesh:
    push_error("BuildingGenerator.generate() must be implemented by subclass")
    return null

# 辅助方法：从建筑类型获取默认颜色
static func get_default_colors(building_type: String) -> Dictionary:
    var colors = {
        "primary": MeshBuilder.PALETTE["wall_light"],
        "secondary": MeshBuilder.PALETTE["roof_terracotta"],
        "accent": MeshBuilder.PALETTE["accent_blue"],
    }
    
    match building_type:
        "gov_hall", "town_hall":
            colors.primary = MeshBuilder.PALETTE["wall_cream"]
            colors.secondary = MeshBuilder.PALETTE["roof_terracotta"]
            colors.accent = MeshBuilder.PALETTE["stone_gray"]
        "bank":
            colors.primary = MeshBuilder.PALETTE["wall_light"]
            colors.secondary = MeshBuilder.PALETTE["stone_gray"]
            colors.accent = MeshBuilder.PALETTE["accent_yellow"]
        "shop":
            colors.primary = MeshBuilder.PALETTE["wall_light"]
            colors.secondary = MeshBuilder.PALETTE["roof_pink"]
            colors.accent = MeshBuilder.PALETTE["accent_blue"]
        "home":
            colors.primary = MeshBuilder.PALETTE["wall_cream"]
            colors.secondary = MeshBuilder.PALETTE["roof_terracotta"]
            colors.accent = MeshBuilder.PALETTE["accent_yellow"]
        "cafe":
            colors.primary = MeshBuilder.PALETTE["wall_light"]
            colors.secondary = MeshBuilder.PALETTE["roof_pink"]
            colors.accent = MeshBuilder.PALETTE["accent_yellow"]
        "office":
            colors.primary = MeshBuilder.PALETTE["wall_light"]
            colors.secondary = MeshBuilder.PALETTE["stone_gray"]
            colors.accent = MeshBuilder.PALETTE["accent_blue"]
        "exchange":
            colors.primary = MeshBuilder.PALETTE["wall_cream"]
            colors.secondary = MeshBuilder.PALETTE["roof_terracotta"]
            colors.accent = MeshBuilder.PALETTE["accent_yellow"]
        "park":
            colors.primary = MeshBuilder.PALETTE["wall_light"]
            colors.secondary = MeshBuilder.PALETTE["roof_pink"]
            colors.accent = MeshBuilder.PALETTE["accent_blue"]
        "guide_hall":
            colors.primary = MeshBuilder.PALETTE["wall_light"]
            colors.secondary = MeshBuilder.PALETTE["roof_pink"]
            colors.accent = MeshBuilder.PALETTE["accent_blue"]
        _:
            # 默认颜色
            pass
    
    return colors
