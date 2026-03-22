extends Panel

# 调试面板 - 控制相机移动到指定位置
# 通过 HTML 按钮点击来控制 Godot 相机

signal focus_requested(pos: Vector3, distance: float)

@onready var info_label: Label = $VBox/InfoLabel

# 预设位置
const POSITIONS = {
	"government_hall": Vector3(-14, 1, -14),
	"guide_hall": Vector3(12, 1, -14),
	"quest_hall": Vector3(-14, 1, 12),
	"shop": Vector3(12, 1, 12),
	"center": Vector3(0, 1, 0),
}

# 默认相机距离
const DEFAULT_DISTANCE = 15.0
const FAR_DISTANCE = 80.0

func _ready():
	# 连接按钮信号
	$VBox/HBox1/BtnGov.pressed.connect(_on_gov_pressed)
	$VBox/HBox1/BtnGuide.pressed.connect(_on_guide_pressed)
	$VBox/HBox1/BtnQuest.pressed.connect(_on_quest_pressed)
	$VBox/HBox1/BtnShop.pressed.connect(_on_shop_pressed)

	$VBox/HBox2/BtnBoundary.pressed.connect(_on_boundary_pressed)
	$VBox/HBox2/BtnCenter.pressed.connect(_on_center_pressed)
	$VBox/HBox2/BtnFar.pressed.connect(_on_far_pressed)

	$VBox/HBox3/BtnZoomOut.pressed.connect(_on_zoom_out_pressed)
	$VBox/HBox3/BtnZoomIn.pressed.connect(_on_zoom_in_pressed)

	_update_info()

func _on_gov_pressed():
	focus_requested.emit(POSITIONS["government_hall"], DEFAULT_DISTANCE)
	_update_info()

func _on_guide_pressed():
	focus_requested.emit(POSITIONS["guide_hall"], DEFAULT_DISTANCE)
	_update_info()

func _on_quest_pressed():
	focus_requested.emit(POSITIONS["quest_hall"], DEFAULT_DISTANCE)
	_update_info()

func _on_shop_pressed():
	focus_requested.emit(POSITIONS["shop"], DEFAULT_DISTANCE)
	_update_info()

func _on_boundary_pressed():
	# 边界处：-20, -20 附近
	focus_requested.emit(Vector3(-20, 1, -20), FAR_DISTANCE)
	_update_info()

func _on_center_pressed():
	focus_requested.emit(POSITIONS["center"], DEFAULT_DISTANCE)
	_update_info()

func _on_far_pressed():
	focus_requested.emit(Vector3(0, 1, 0), FAR_DISTANCE)
	_update_info()

func _on_zoom_out_pressed():
	# 通知外部放大距离
	focus_requested.emit(Vector3(-999, -999, -999), FAR_DISTANCE)  # -999 表示 zoom out

func _on_zoom_in_pressed():
	focus_requested.emit(Vector3(-999, -999, -999), 5.0)  # -999 表示 zoom in

func _update_info():
	info_label.text = "Use buttons above to navigate"

func set_camera_info(pos: Vector3, distance: float):
	info_label.text = "Target: (%.0f, %.0f, %.0f) Dist: %.0f" % [pos.x, pos.y, pos.z, distance]