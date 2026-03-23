class_name GovHallGenerator
extends BuildingGenerator

# 政府大厅 - 纪念碑谷风格
# 特征：对称、台阶基座、柱廊、圆顶

func generate() -> ArrayMesh:
    var builder = MeshBuilder.new()
    builder.begin()
    
    # 参数计算（基于纪念碑谷的比例美学）
    var base_height = height * 0.15      # 基座高度占比
    var body_height = height * 0.55      # 主体高度占比
    var roof_height = height * 0.30      # 屋顶高度占比
    
    var step_count = 3
    var step_height = base_height / step_count
    
    # ===== 1. 基座台阶 =====
    # 阶梯式基座，每层比上一层小
    for i in range(step_count):
        var step_scale = 1.0 + (step_count - i) * 0.08  # 底层更大
        var step_w = width * step_scale
        var step_d = depth * step_scale
        var step_y = i * step_height + step_height * 0.5
        
        builder.add_box(
            Vector3(step_w, step_height, step_d),
            Vector3(0, step_y, 0),
            secondary_color.darkened(0.1)  # 深色石材
        )
    
    # ===== 2. 主体建筑 =====
    var body_y = base_height + body_height * 0.5
    builder.add_box(
        Vector3(width * 0.9, body_height, depth * 0.9),
        Vector3(0, body_y, 0),
        primary_color
    )
    
    # ===== 3. 柱廊（正面）=====
    var column_count = 4
    var column_radius = width * 0.06
    var column_height = body_height * 0.7
    var column_y = base_height + column_height * 0.5
    var front_z = -depth * 0.4  # 略在建筑前面
    
    for i in range(column_count):
        var t = (i + 0.5) / column_count  # 0.125, 0.375, 0.625, 0.875
        var col_x = (t - 0.5) * width * 0.7
        
        builder.add_cylinder(
            column_radius,
            column_height,
            8,  # 8边形柱
            Vector3(col_x, column_y, front_z),
            accent_color
        )
    
    # ===== 4. 三角门楣 =====
    var pediment_height = body_height * 0.15
    var pediment_y = base_height + column_height + pediment_height * 0.5
    var pediment_depth = depth * 0.15
    
    builder.add_triangular_prism(
        width * 0.95,
        pediment_height,
        pediment_depth,
        Vector3(0, pediment_y, front_z),
        secondary_color
    )
    
    # ===== 5. 圆顶 =====
    var dome_y = base_height + body_height
    var dome_radius = width * 0.35
    var drum_height = roof_height * 0.3
    var dome_height = roof_height * 0.7
    
    # 圆顶基座（鼓座）
    builder.add_cylinder(
        dome_radius * 0.85,
        drum_height,
        12,
        Vector3(0, dome_y + drum_height * 0.5, 0),
        secondary_color
    )
    
    # 圆顶主体（半球）- 用多层逐渐缩小的圆柱模拟
    var dome_layers = 4
    for i in range(dome_layers):
        var t = i / float(dome_layers)
        var next_t = (i + 1) / float(dome_layers)
        
        var layer_radius = dome_radius * (1.0 - t * t)  # 抛物线收缩
        var next_radius = dome_radius * (1.0 - next_t * next_t)
        
        var layer_y = dome_y + drum_height + t * dome_height
        var layer_h = (next_t - t) * dome_height
        
        # 使用棱台效果（上小下大的圆柱段）
        # 为了简单，用圆柱近似
        var avg_radius = (layer_radius + next_radius) * 0.5
        builder.add_cylinder(
            avg_radius,
            layer_h,
            12,
            Vector3(0, layer_y + layer_h * 0.5, 0),
            secondary_color.lightened(0.1)
        )
    
    # 圆顶顶端装饰
    builder.add_cylinder(
        dome_radius * 0.15,
        dome_height * 0.2,
        6,
        Vector3(0, dome_y + drum_height + dome_height + dome_height * 0.1, 0),
        accent_color
    )
    
    # ===== 6. 侧面装饰窗 =====
    var window_size = Vector3(width * 0.1, height * 0.15, depth * 0.05)
    var window_y = body_y
    var window_x = width * 0.35
    
    # 左右各一个装饰窗
    builder.add_box(window_size, Vector3(window_x, window_y, 0), accent_color.darkened(0.2))
    builder.add_box(window_size, Vector3(-window_x, window_y, 0), accent_color.darkened(0.2))
    
    return builder.build()
