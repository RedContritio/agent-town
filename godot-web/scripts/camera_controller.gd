extends Camera3D

class_name CameraController

# 相机系统设计：
# - 第三人称视角，相机始终看向目标点
# - 限制相机不能低于地面，不能进入方块内部
# - 避免完全俯视导致的视角奇点

@export var min_distance: float = 5.0     # 最小距离(米)
@export var max_distance: float = 150.0   # 最大距离(米)，允许更远视角
@export var min_height: float = 2.5       # 相机最低高度(米)，高于地面方块
@export var zoom_speed: float = 0.1       # 缩放速度
@export var rotate_speed: float = 0.005   # 旋转速度
@export var pan_speed: float = 0.015      # 平移速度

# 极角限制（弧度）
# 0° = 正上方（会导致视角奇点，禁止）
# 5° = 接近俯视（允许的最小角度）
# 90° = 水平面（地平线，不能再低）
@export var min_polar: float = deg_to_rad(5.0)     # 最垂直（接近俯视但不完全）
@export var max_polar: float = deg_to_rad(90.0)    # 最水平（地平线，不允许仰视）

var target_position: Vector3 = Vector3(0, 1.5, 0)  # 目标点高度1.5米
var distance: float = 25.0                         # 初始距离25米
var azimuth: float = -PI / 4                       # 水平-45度
var polar: float = deg_to_rad(60.0)                # 初始60度（适度俯视）

var is_rotating := false
var is_panning := false
var last_mouse_pos: Vector2

func _ready():
	update_camera_position()

func _input(event):
	# 鼠标滚轮缩放
	if event is InputEventMouseButton:
		if event.button_index == MOUSE_BUTTON_WHEEL_UP:
			distance *= (1.0 - zoom_speed)
			distance = clamp(distance, min_distance, max_distance)
			update_camera_position()
			get_viewport().set_input_as_handled()
			return
		elif event.button_index == MOUSE_BUTTON_WHEEL_DOWN:
			distance *= (1.0 + zoom_speed)
			distance = clamp(distance, min_distance, max_distance)
			update_camera_position()
			get_viewport().set_input_as_handled()
			return
		elif event.button_index == MOUSE_BUTTON_LEFT:
			is_rotating = event.pressed
			if event.pressed:
				last_mouse_pos = event.position
			get_viewport().set_input_as_handled()
			return
		elif event.button_index == MOUSE_BUTTON_RIGHT:
			is_panning = event.pressed
			if event.pressed:
				last_mouse_pos = event.position
			get_viewport().set_input_as_handled()
			return
	
	elif event is InputEventMouseMotion:
		if is_rotating:
			var delta = event.position - last_mouse_pos
			azimuth -= delta.x * rotate_speed
			polar -= delta.y * rotate_speed
			# 限制极角：不能接近完全俯视（避免奇点），不能低于地平线
			polar = clamp(polar, min_polar, max_polar)
			update_camera_position()
			last_mouse_pos = event.position
			get_viewport().set_input_as_handled()
		elif is_panning:
			var delta = event.position - last_mouse_pos
			var right = transform.basis.x
			var forward = -transform.basis.z
			forward.y = 0
			forward = forward.normalized()
			
			var move_speed = pan_speed * (distance / 15.0)
			# 水平拖动：delta.x 正=右，目标点应向左移动（场景向右）
			target_position -= right * delta.x * move_speed
			# 垂直拖动：delta.y 正=下，目标点应向相机前方移动（场景向下）
			# 向上拖动(delta.y负)时，目标点应向相机后方移动
			target_position += forward * delta.y * move_speed
			update_camera_position()
			last_mouse_pos = event.position
			get_viewport().set_input_as_handled()

func update_camera_position():
	# 球坐标计算相机位置
	var sin_polar = sin(polar)
	var cos_polar = cos(polar)
	var sin_azimuth = sin(azimuth)
	var cos_azimuth = cos(azimuth)
	
	var x = target_position.x + distance * sin_polar * cos_azimuth
	var y = target_position.y + distance * cos_polar
	var z = target_position.z + distance * sin_polar * sin_azimuth
	
	# 强制最低高度，确保相机不低于地面
	if y < min_height:
		# 调整距离或极角来满足高度要求
		# 保持水平方向不变，只调整垂直
		var required_cos = (min_height - target_position.y) / distance
		required_cos = clamp(required_cos, cos(max_polar), cos(min_polar))
		polar = acos(required_cos)
		cos_polar = required_cos
		sin_polar = sqrt(1.0 - cos_polar * cos_polar)
		
		# 重新计算位置
		x = target_position.x + distance * sin_polar * cos_azimuth
		y = target_position.y + distance * cos_polar
		z = target_position.z + distance * sin_polar * sin_azimuth
	
	position = Vector3(x, y, z)
	
	# 统一使用 UP 向量，避免 polar 跨越 PI/2 时的突变
	# 使用 quaternion 避免万向节锁
	look_at(target_position, Vector3.UP)

func focus_on_position(pos: Vector3):
	target_position = Vector3(pos.x, 1.5, pos.z)
	update_camera_position()

func get_camera_info() -> Dictionary:
	return {
		"position": position,
		"target": target_position,
		"distance": distance,
		"azimuth_deg": rad_to_deg(azimuth),
		"polar_deg": rad_to_deg(polar),
		"height": position.y
	}

func get_target_position() -> Vector3:
	return target_position

# Set camera to view a specific position from a specific direction
func set_view_from_direction(target_pos: Vector3, direction: String, distance_override: float = 15.0):
	# direction: "north", "south", "east", "west", "top"
	target_position = target_pos
	
	match direction:
		"north":  # View from north (looking south, -Z to +Z)
			azimuth = PI / 2  # 90 degrees, looking south
			polar = PI / 3    # 60 degrees from vertical
		"south":  # View from south (looking north, +Z to -Z)
			azimuth = -PI / 2  # -90 degrees, looking north
			polar = PI / 3
		"east":   # View from east (looking west, +X to -X)
			azimuth = PI  # 180 degrees, looking west
			polar = PI / 3
		"west":   # View from west (looking east, -X to +X)
			azimuth = 0  # 0 degrees, looking east
			polar = PI / 3
		"top":    # Top down view
			azimuth = 0
			polar = PI / 6  # 30 degrees from vertical
		_:
			azimuth = 0
			polar = PI / 3
	
	distance = distance_override
	update_camera_position()

# Quick method to view Gov Hall north wall
func view_gov_hall_north_wall():
	# Gov Hall at (-12, 0, -12), north wall is on -Z side
	# Position camera north of the building, looking south
	set_view_from_direction(Vector3(-12, 1.5, -12), "north", 8.0)
