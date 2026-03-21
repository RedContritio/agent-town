extends CharacterBody3D

class_name AgentEntity

var agent_id: String
var agent_name: String
var agent_color: Color
var target_position: Vector3
var agent_data: Dictionary = {}

@onready var mesh: MeshInstance3D = $Mesh
@onready var agent_panel: AgentPanel = get_tree().get_root().get_node("Main/CanvasLayer/AgentPanel")

const COLORS = [
	Color(0.31, 0.76, 0.97),  # Cyan
	Color(0.16, 0.84, 0.45),  # Green
	Color(0.95, 0.77, 0.06),  # Yellow
	Color(0.96, 0.26, 0.35),  # Red
	Color(0.85, 0.39, 0.95),  # Purple
	Color(1.0, 0.6, 0.2),     # Orange
]

func setup(data: Dictionary):
	setup_with_height(data, 0.0)

func setup_with_height(data: Dictionary, ground_height: float):
	agent_data = data
	agent_id = data.get("id", "")
	agent_name = data.get("name", "Unknown")
	
	var pos = data.get("position", {})
	var x = pos.get("x", 0)
	var z = pos.get("y", 0)  # API y is horizontal z in Godot
	
	# Agent only has 2D position (x, y), height is determined by terrain
	# ground_height is the surface height (block top at y=0.5 for surface block)
	# Agent center should be at ground_height + 0.5 (half agent height)
	var final_y = ground_height + 0.5
	
	target_position = Vector3(x, final_y, z)
	position = target_position
	
	# Set color based on index in id
	var color_index = hash(agent_id) % COLORS.size()
	agent_color = COLORS[color_index]
	
	# Update mesh color
	if mesh:
		var material = StandardMaterial3D.new()
		material.albedo_color = agent_color
		mesh.material_override = material
	else:
		# Mesh not ready yet, defer material setup
		call_deferred("_apply_material")

func _apply_material():
	if mesh:
		var material = StandardMaterial3D.new()
		material.albedo_color = agent_color
		mesh.material_override = material
	

func _ready():
	# Apply material if not already set
	if mesh and mesh.material_override == null:
		var material = StandardMaterial3D.new()
		material.albedo_color = agent_color
		mesh.material_override = material
	
	# Bobbing animation - apply to mesh only, not the entire node
	# This ensures name labels (in screen space UI) don't bounce with the model
	_start_bobbing_animation()
	
	# Enable input processing for click detection
	input_event.connect(_on_input_event)
	# Make sure we have collision for clicking
	if not has_node("CollisionShape3D"):
		var collision = CollisionShape3D.new()
		var shape = BoxShape3D.new()
		shape.size = Vector3(0.8, 1.8, 0.8)
		collision.shape = shape
		add_child(collision)

func _on_input_event(camera: Node, event: InputEvent, position: Vector3, normal: Vector3, shape_idx: int):
	if event is InputEventMouseButton:
		if event.button_index == MOUSE_BUTTON_LEFT and event.pressed:
			# Show agent panel with this agent's data
			if agent_panel and not agent_data.is_empty():
				agent_panel.show_for_agent(agent_data)
				print("Selected agent: ", agent_name)

func _start_bobbing_animation():
	if mesh == null:
		return
	var tween = create_tween()
	tween.set_loops()
	# Animate mesh local position instead of node global position
	tween.tween_property(mesh, "position:y", 0.1, 0.5)
	tween.tween_property(mesh, "position:y", 0.0, 0.5)

func update_data(data: Dictionary):
	var pos = data.get("position", {})
	target_position = Vector3(
		pos.get("x", 0),
		pos.get("z", 0),
		pos.get("y", 0)
	)
	# Smooth movement could be added here
	position = target_position
