extends CanvasLayer

class_name NameLabelManager

@onready var camera: Camera3D = get_node("../Camera3D")

var labels: Dictionary = {}  # node_path -> Label

# Distance-based scaling settings
const BASE_FONT_SIZE: int = 14          # Base font size at reference distance
const REFERENCE_DISTANCE: float = 10.0   # Distance at which font size equals BASE_FONT_SIZE
const MIN_FONT_SIZE: int = 8             # Minimum font size (smaller than this = hidden)
const MAX_FONT_SIZE: int = 32            # Maximum font size cap
const LABEL_HIDE_DISTANCE: float = 60.0  # Hide labels beyond this distance

func _ready():
	pass

func _process(_delta):
	# Update all label positions and sizes
	for node_path in labels.keys():
		var node = get_node_or_null(node_path)
		var label = labels[node_path]
		
		if node == null or label == null:
			continue
		
		# Get stable node position in world space (ignoring any visual animation)
		var base_position = node.global_position
		
		# Add height offset for label (above the entity)
		var label_height_offset = 0.7
		var world_pos = base_position + Vector3(0, label_height_offset, 0)
		
		# Calculate distance from camera to node
		var distance = camera.global_position.distance_to(world_pos)
		
		# Check if behind camera or too far
		if camera.is_position_behind(world_pos) or distance > LABEL_HIDE_DISTANCE:
			label.visible = false
			continue
		
		# Calculate font size based on distance (perspective scaling)
		# Closer = larger, Farther = smaller
		var scale_factor = REFERENCE_DISTANCE / distance
		var target_font_size = int(BASE_FONT_SIZE * scale_factor)
		
		# Clamp to min/max size
		target_font_size = clamp(target_font_size, MIN_FONT_SIZE, MAX_FONT_SIZE)
		
		# Hide if smaller than minimum (shouldn't happen due to clamp, but for safety)
		if target_font_size < MIN_FONT_SIZE:
			label.visible = false
			continue
		
		# Use scale to adjust label size based on distance
		# This scales both the text AND the background box together
		var label_scale = target_font_size / float(BASE_FONT_SIZE)
		label.scale = Vector2(label_scale, label_scale)
		
		# Project to screen space
		var screen_pos = camera.unproject_position(world_pos)
		
		# Show label and center it (account for scale)
		label.visible = true
		label.position = screen_pos - (label.size * label.scale) / 2

const ACCENT_CYAN = Color("#4fc3f7")

func add_label(node: Node3D, text: String, color: Color = ACCENT_CYAN):
	var label = Label.new()
	label.text = text
	label.add_theme_color_override("font_color", color)
	label.add_theme_font_size_override("font_size", BASE_FONT_SIZE)
	
	# Add semi-transparent white background for low-profile readability
	var style = StyleBoxFlat.new()
	style.bg_color = Color(1, 1, 1, 0.15)  # 15% opacity white
	style.corner_radius_top_left = 4
	style.corner_radius_top_right = 4
	style.corner_radius_bottom_left = 4
	style.corner_radius_bottom_right = 4
	style.content_margin_left = 6
	style.content_margin_right = 6
	style.content_margin_top = 2
	style.content_margin_bottom = 2
	label.add_theme_stylebox_override("normal", style)
	
	# Add subtle shadow for depth
	label.add_theme_constant_override("shadow_offset_x", 1)
	label.add_theme_constant_override("shadow_offset_y", 1)
	label.add_theme_color_override("font_shadow_color", Color(0, 0, 0, 0.5))
	
	add_child(label)
	labels[node.get_path()] = label

func remove_label(node: Node3D):
	var path = node.get_path()
	if labels.has(path):
		labels[path].queue_free()
		labels.erase(path)

func clear_labels():
	for label in labels.values():
		label.queue_free()
	labels.clear()
