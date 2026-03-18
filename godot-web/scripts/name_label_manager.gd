extends CanvasLayer

class_name NameLabelManager

@onready var camera: Camera3D = get_node("../Camera3D")

var labels: Dictionary = {}  # node_path -> Label

func _ready():
	pass

func _process(_delta):
	# Update all label positions
	for node_path in labels.keys():
		var node = get_node_or_null(node_path)
		var label = labels[node_path]
		
		if node == null or label == null:
			continue
		
		# Get node position in world space (above the entity)
		var world_pos = node.global_position + Vector3(0, 1.0, 0)
		
		# Project to screen space
		var screen_pos = camera.unproject_position(world_pos)
		
		# Check if behind camera
		if camera.is_position_behind(world_pos):
			label.visible = false
		else:
			label.visible = true
			label.position = screen_pos - label.size / 2

func add_label(node: Node3D, text: String, color: Color = Color.CYAN):
	var label = Label.new()
	label.text = text
	label.add_theme_color_override("font_color", color)
	label.add_theme_font_size_override("font_size", 14)
	
	# Add outline effect
	var style = label.get_theme_stylebox("normal")
	if style == null:
		style = StyleBoxEmpty.new()
	label.add_theme_stylebox_override("normal", style)
	
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
