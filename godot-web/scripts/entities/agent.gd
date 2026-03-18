extends CharacterBody3D

class_name AgentEntity

var agent_id: String
var agent_name: String
var agent_color: Color
var target_position: Vector3

@onready var mesh: MeshInstance3D = $Mesh
@onready var label: Label3D = $Label3D

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
	agent_id = data.get("id", "")
	agent_name = data.get("name", "Unknown")
	
	var pos = data.get("position", {})
	var x = pos.get("x", 0)
	var y = pos.get("z", 0)  # API z is height
	var z = pos.get("y", 0)  # API y is z in Godot
	
	# Place agent on top of ground
	# ground_height is the top of the block (block_y + 0.5 * block_height)
	# Agent should stand on top, so y = ground_height + 0.5 (half of agent height)
	var final_y = max(y, ground_height + 0.5)
	
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
	
	# Update label
	if label:
		label.text = agent_name

func _ready():
	# Bobbing animation
	var tween = create_tween()
	tween.set_loops()
	tween.tween_property(self, "position:y", position.y + 0.1, 0.5)
	tween.tween_property(self, "position:y", position.y, 0.5)

func update_data(data: Dictionary):
	var pos = data.get("position", {})
	target_position = Vector3(
		pos.get("x", 0),
		pos.get("z", 0),
		pos.get("y", 0)
	)
	# Smooth movement could be added here
	position = target_position
