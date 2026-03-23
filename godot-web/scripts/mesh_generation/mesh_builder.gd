class_name MeshBuilder

# 纪念碑谷风格网格构建器
# 特点：硬边几何、纯色平面、清晰轮廓

var _vertices: PackedVector3Array = PackedVector3Array()
var _normals: PackedVector3Array = PackedVector3Array()
var _colors: PackedColorArray = PackedColorArray()
var _indices: PackedInt32Array = PackedInt32Array()
var _index_offset: int = 0

# 纪念碑谷风格调色板（柔和低饱和度）
const PALETTE = {
    "wall_light": Color("#f4e4d4"),    # 暖白墙
    "wall_cream": Color("#e8d5c4"),    # 奶油墙
    "roof_terracotta": Color("#c4a484"), # 陶土屋顶
    "roof_pink": Color("#d4a5a5"),     # 粉屋顶
    "stone_gray": Color("#a8a8a8"),    # 石材灰
    "accent_blue": Color("#8bb8d8"),   # 装饰蓝
    "accent_yellow": Color("#e8d488"), # 装饰黄
    "dark_shadow": Color("#6a6a6a"),   # 阴影色
}

func begin() -> void:
    _vertices.clear()
    _normals.clear()
    _colors.clear()
    _indices.clear()
    _index_offset = 0

# 添加硬边立方体（纪念碑谷风格，无倒角）
func add_box(size: Vector3, position: Vector3, color: Color) -> void:
    var hx = size.x * 0.5
    var hy = size.y * 0.5
    var hz = size.z * 0.5
    
    # 8 个顶点（从底部中心位置偏移）
    var base_pos = position - Vector3(0, hy, 0)  # 调整使 position 是底部中心
    var v = [
        base_pos + Vector3(-hx, 0, -hz),    # 0: 左下后
        base_pos + Vector3(hx, 0, -hz),     # 1: 右下后
        base_pos + Vector3(hx, 0, hz),      # 2: 右下前
        base_pos + Vector3(-hx, 0, hz),     # 3: 左下前
        base_pos + Vector3(-hx, size.y, -hz), # 4: 左上后
        base_pos + Vector3(hx, size.y, -hz),  # 5: 右上后
        base_pos + Vector3(hx, size.y, hz),   # 6: 右上前
        base_pos + Vector3(-hx, size.y, hz),  # 7: 左上前
    ]
    
    # 6 个面，每个面 4 个顶点，2 个三角形
    var faces = [
        [0, 1, 5, 4],  # 后面 (Z-)
        [2, 3, 7, 6],  # 前面 (Z+)
        [0, 4, 7, 3],  # 左面 (X-)
        [1, 2, 6, 5],  # 右面 (X+)
        [4, 5, 6, 7],  # 顶面 (Y+)
        [0, 3, 2, 1],  # 底面 (Y-)
    ]
    
    var face_normals = [
        Vector3(0, 0, -1),  # 后
        Vector3(0, 0, 1),   # 前
        Vector3(-1, 0, 0),  # 左
        Vector3(1, 0, 0),   # 右
        Vector3(0, 1, 0),   # 上
        Vector3(0, -1, 0),  # 下
    ]
    
    for i in range(6):
        _add_quad(v[faces[i][0]], v[faces[i][1]], v[faces[i][2]], v[faces[i][3]], 
                  face_normals[i], color)

# 添加台阶（纪念碑谷核心元素）
func add_steps(width: float, depth: float, step_height: float, step_count: int, 
               position: Vector3, color: Color) -> void:
    for i in range(step_count):
        var step_size = Vector3(width, step_height, depth)
        var step_pos = position + Vector3(0, i * step_height + step_height * 0.5, 0)
        add_box(step_size, step_pos, color)

# 添加圆柱体/棱柱
func add_cylinder(radius: float, height: float, sides: int, position: Vector3, color: Color) -> void:
    var half_height = height * 0.5
    var base_pos = position - Vector3(0, half_height, 0)
    
    # 生成圆周顶点
    var bottom_vertices = []
    var top_vertices = []
    
    for i in range(sides):
        var angle = (i / float(sides)) * TAU
        var x = cos(angle) * radius
        var z = sin(angle) * radius
        bottom_vertices.append(base_pos + Vector3(x, 0, z))
        top_vertices.append(base_pos + Vector3(x, height, z))
    
    # 侧面
    for i in range(sides):
        var next = (i + 1) % sides
        var normal = (bottom_vertices[i] - base_pos).normalized()
        normal.y = 0
        
        _add_quad(bottom_vertices[i], bottom_vertices[next], 
                  top_vertices[next], top_vertices[i], normal, color)
    
    # 顶面和底面（三角扇）
    var center_bottom = base_pos + Vector3(0, 0, 0)
    var center_top = base_pos + Vector3(0, height, 0)
    
    for i in range(sides):
        var next = (i + 1) % sides
        # 底面（向下法线）
        _add_triangle(bottom_vertices[i], bottom_vertices[next], center_bottom, 
                      Vector3(0, -1, 0), color)
        # 顶面（向上法线）
        _add_triangle(top_vertices[next], top_vertices[i], center_top, 
                      Vector3(0, 1, 0), color)

# 添加金字塔/锥体
func add_pyramid(base_size: Vector3, height: float, position: Vector3, color: Color) -> void:
    var hx = base_size.x * 0.5
    var hz = base_size.z * 0.5
    
    var base_pos = position
    var v_bottom = [
        base_pos + Vector3(-hx, 0, -hz),
        base_pos + Vector3(hx, 0, -hz),
        base_pos + Vector3(hx, 0, hz),
        base_pos + Vector3(-hx, 0, hz),
    ]
    var v_top = base_pos + Vector3(0, height, 0)
    
    # 底面
    _add_quad(v_bottom[0], v_bottom[3], v_bottom[2], v_bottom[1], 
              Vector3(0, -1, 0), color)
    
    # 四个侧面
    var side_normals = [
        Vector3(0, 0, -1).normalized(),
        Vector3(1, 0, 0).normalized(),
        Vector3(0, 0, 1).normalized(),
        Vector3(-1, 0, 0).normalized(),
    ]
    
    for i in range(4):
        var next = (i + 1) % 4
        var normal = ((v_bottom[i] - v_top).cross(v_bottom[next] - v_top)).normalized()
        _add_triangle(v_bottom[i], v_bottom[next], v_top, normal, color)

# 添加三角棱柱（用于屋顶、门楣）
func add_triangular_prism(width: float, height: float, depth: float, 
                         position: Vector3, color: Color) -> void:
    var hx = width * 0.5
    var hy = height
    var hz = depth * 0.5
    
    var base_pos = position
    
    # 三棱柱顶点
    var v = [
        base_pos + Vector3(-hx, 0, -hz),   # 0: 左下后
        base_pos + Vector3(hx, 0, -hz),    # 1: 右下后
        base_pos + Vector3(hx, 0, hz),     # 2: 右下前
        base_pos + Vector3(-hx, 0, hz),    # 3: 左下前
        base_pos + Vector3(0, hy, -hz),    # 4: 顶后
        base_pos + Vector3(0, hy, hz),     # 5: 顶前
    ]
    
    # 底面（矩形）
    _add_quad(v[0], v[1], v[2], v[3], Vector3(0, -1, 0), color)
    
    # 两个三角形端面
    _add_triangle(v[0], v[4], v[1], Vector3(0, 0, -1), color)
    _add_triangle(v[3], v[2], v[5], Vector3(0, 0, 1), color)
    
    # 两个斜侧面
    var normal_left = ((v[0] - v[4]).cross(v[3] - v[4])).normalized()
    var normal_right = ((v[1] - v[5]).cross(v[2] - v[5])).normalized()
    
    _add_quad(v[0], v[3], v[5], v[4], normal_left, color)
    _add_quad(v[1], v[4], v[5], v[2], normal_right, color)

# 添加拱门（纪念碑谷常用元素）
func add_arch(width: float, height: float, depth: float, position: Vector3, color: Color) -> void:
    # 简化版：U 形结构
    var pillar_width = width * 0.25
    var pillar_height = height * 0.6
    var top_height = height * 0.4
    var hx = width * 0.5
    var hz = depth * 0.5
    var base_pos = position
    
    # 左柱
    add_box(Vector3(pillar_width, pillar_height, depth), 
            base_pos + Vector3(-hx + pillar_width * 0.5, pillar_height * 0.5, 0), color)
    
    # 右柱
    add_box(Vector3(pillar_width, pillar_height, depth), 
            base_pos + Vector3(hx - pillar_width * 0.5, pillar_height * 0.5, 0), color)
    
    # 顶部横梁
    add_box(Vector3(width, top_height, depth), 
            base_pos + Vector3(0, pillar_height + top_height * 0.5, 0), color)

# 内部方法：添加四边形（2个三角形）
func _add_quad(v0: Vector3, v1: Vector3, v2: Vector3, v3: Vector3, 
               normal: Vector3, color: Color) -> void:
    var start = _vertices.size()
    
    _vertices.append_array([v0, v1, v2, v3])
    _normals.append_array([normal, normal, normal, normal])
    _colors.append_array([color, color, color, color])
    
    # 两个三角形：0-1-2 和 0-2-3
    _indices.append_array([
        start, start + 1, start + 2,
        start, start + 2, start + 3
    ])

# 内部方法：添加三角形
func _add_triangle(v0: Vector3, v1: Vector3, v2: Vector3, 
                   normal: Vector3, color: Color) -> void:
    var start = _vertices.size()
    
    _vertices.append_array([v0, v1, v2])
    _normals.append_array([normal, normal, normal])
    _colors.append_array([color, color, color])
    
    _indices.append_array([start, start + 1, start + 2])

# 构建最终网格
func build() -> ArrayMesh:
    if _vertices.size() == 0:
        push_warning("MeshBuilder: no vertices added")
        return null
    
    var mesh = ArrayMesh.new()
    var arrays = []
    arrays.resize(Mesh.ARRAY_MAX)
    
    arrays[Mesh.ARRAY_VERTEX] = _vertices
    arrays[Mesh.ARRAY_NORMAL] = _normals
    arrays[Mesh.ARRAY_COLOR] = _colors
    arrays[Mesh.ARRAY_INDEX] = _indices
    
    mesh.add_surface_from_arrays(Mesh.PRIMITIVE_TRIANGLES, arrays)
    
    # 创建材质（使用顶点颜色）
    var material = StandardMaterial3D.new()
    material.vertex_color_use_as_albedo = true
    material.shading_mode = BaseMaterial3D.SHADING_MODE_PER_PIXEL
    material.roughness = 0.9  # 柔和无光泽，纪念碑谷风格
    material.metallic = 0.0
    material.specular_mode = BaseMaterial3D.SPECULAR_DISABLED
    material.disable_receive_shadows = false
    
    mesh.surface_set_material(0, material)
    
    return mesh

# 获取当前顶点数（用于调试）
func get_vertex_count() -> int:
    return _vertices.size()
