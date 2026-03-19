# ui_debug_overlay.gd
# Attach this to a CanvasLayer in your main scene to enable UI debugging
# Shows: click areas, anchor points, margins, control bounds

extends CanvasLayer

@export var show_click_areas: bool = true
@export var show_control_bounds: bool = true
@export var show_anchor_points: bool = true
@export var log_ui_interactions: bool = true

var _debug_draw_layer: Control
var _interaction_log: Array = []

func _ready():
	name = "UIDebugOverlay"
	layer = 1000  # On top of everything
	
	_debug_draw_layer = Control.new()
	_debug_draw_layer.name = "DebugDrawLayer"
	_debug_draw_layer.set_anchors_preset(Control.PRESET_FULL_RECT)
	_debug_draw_layer.mouse_filter = Control.MOUSE_FILTER_IGNORE
	add_child(_debug_draw_layer)
	
	_debug_draw_layer.draw.connect(_on_debug_draw)
	
	if log_ui_interactions:
		_get_viewport().gui_focus_changed.connect(_on_focus_changed)
		print("[UI Debug] Overlay enabled - Press F12 to toggle visibility")
	
	# Auto-refresh
	set_process(true)

func _process(_delta):
	_debug_draw_layer.queue_redraw()

func _input(event):
	if event is InputEventKey and event.pressed and event.keycode == KEY_F12:
		visible = !visible
		print("[UI Debug] Overlay visibility: ", visible)
	
	if log_ui_interactions and event is InputEventMouseButton and event.pressed:
		_log_mouse_click(event)

func _on_debug_draw():
	if not visible:
		return
	
	var root = get_tree().root
	_traverse_and_draw(root)

func _traverse_and_draw(node: Node):
	if node is Control:
		if show_control_bounds:
			_draw_control_bounds(node)
		if show_anchor_points:
			_draw_anchor_point(node)
		if show_click_areas and (node is Button or node is TextureButton or node is CheckBox):
			_draw_click_area(node)
	
	for child in node.get_children():
		_traverse_and_draw(child)

func _draw_control_bounds(control: Control):
	var rect = control.get_global_rect()
	var color = Color.CYAN if control.visible else Color.GRAY
	color.a = 0.3
	_debug_draw_layer.draw_rect(rect, color, false, 2.0)
	
	# Draw name
	if control.name:
		_debug_draw_layer.draw_string(
			preload("res://assets/fonts/default_font.tres") if ResourceLoader.exists("res://assets/fonts/default_font.tres") else SystemFont.new(),
			rect.position + Vector2(4, 14),
			control.name,
			HORIZONTAL_ALIGNMENT_LEFT,
			-1,
			12,
			Color.YELLOW
		)

func _draw_anchor_point(control: Control):
	var rect = control.get_global_rect()
	var anchor_pos = rect.position + Vector2(
		rect.size.x * control.anchor_right,
		rect.size.y * control.anchor_bottom
	)
	_debug_draw_layer.draw_circle(anchor_pos, 4, Color.RED)
	_debug_draw_layer.draw_line(rect.position, anchor_pos, Color.RED, 1.0, true)

func _draw_click_area(control: Control):
	var rect = control.get_global_rect()
	var color = Color.GREEN
	color.a = 0.2
	_debug_draw_layer.draw_rect(rect, color, true)
	_debug_draw_layer.draw_rect(rect, Color.GREEN, false, 2.0)

func _log_mouse_click(event: InputEventMouseButton):
	var clicked_control = _get_control_at_position(event.global_position)
	if clicked_control:
		var entry = {
			"time": Time.get_time_string_from_system(),
			"position": event.global_position,
			"button": event.button_index,
			"control": clicked_control.name,
			"control_path": clicked_control.get_path(),
			"rect": clicked_control.get_global_rect()
		}
		_interaction_log.append(entry)
		print("[UI Debug] Click: %s at %s on %s (%s)" % [
			event.button_index,
			event.global_position,
			clicked_control.name,
			clicked_control.get_global_rect()
		])

func _get_control_at_position(pos: Vector2) -> Control:
	var root = get_tree().root
	return _find_control_at(root, pos)

func _find_control_at(node: Node, pos: Vector2) -> Control:
	if node is Control:
		var rect = node.get_global_rect()
		if rect.has_point(pos):
			# Check children first (top to bottom)
			for i in range(node.get_child_count() - 1, -1, -1):
				var child = node.get_child(i)
				var found = _find_control_at(child, pos)
				if found:
					return found
			return node
	return null

func _on_focus_changed(control: Control):
	if control:
		print("[UI Debug] Focus changed to: %s (%s)" % [control.name, control.get_path()])

func get_interaction_log() -> Array:
	return _interaction_log.duplicate()

func clear_log():
	_interaction_log.clear()
