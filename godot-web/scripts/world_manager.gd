extends Node3D

class_name WorldManager

@onready var api_client: ApiClient = $ApiClient
@onready var terrain_root: Node3D = $Terrain
@onready var buildings_root: Node3D = $Buildings
@onready var agents_root: Node3D = $Agents
@onready var name_labels: NameLabelManager = get_node("/root/Main/NameLabels")

const AGENT_SCENE := preload("res://scenes/entities/agent.tscn")
const BUILDING_SCENE := preload("res://scenes/entities/building.tscn")

var terrain_materials: Dictionary = {}
var world_info: Dictionary = {}
var terrain_height_map: Dictionary = {}  # "x,z" -> max_height (top of block)
var pending_agents: Array = []  # Store agents until terrain is loaded

func _ready():
	# Create terrain materials with visual design colors
	# Colors from docs/VISUAL_DESIGN.md - Appendix B: tileColorMap
	terrain_materials["grass"] = _create_material(Color("#5a8f63"))
	terrain_materials["road"] = _create_material(Color("#c8b08a"))
	terrain_materials["water"] = _create_material(Color("#4a92b8"))
	terrain_materials["farmland"] = _create_material(Color("#9a7820"))
	terrain_materials["sand"] = _create_material(Color("#d4c088"))
	terrain_materials["foundation"] = _create_material(Color("#6a6a6a"))
	terrain_materials["hill"] = _create_material(Color("#4a6a4a"))
	terrain_materials["door"] = _create_material(Color("#b89955"))
	terrain_materials["fence"] = _create_material(Color("#8a7254"))
	terrain_materials["bridge"] = _create_material(Color("#a08050"))
	terrain_materials["indoor_floor"] = _create_material(Color("#c4a882"))
	terrain_materials["indoor_wall"] = _create_material(Color("#e8e0d0"))
	terrain_materials["indoor_window"] = _create_material(Color("#add8e6"))
	
	# Connect API signals
	api_client.world_info_received.connect(_on_world_info)
	api_client.agents_received.connect(_on_agents_update)
	api_client.map_received.connect(_on_map_received)

func _create_material(color: Color) -> StandardMaterial3D:
	var mat = StandardMaterial3D.new()
	mat.albedo_color = color
	mat.roughness = 0.9
	mat.vertex_color_use_as_albedo = false
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_PER_PIXEL
	mat.specular_mode = BaseMaterial3D.SPECULAR_DISABLED
	mat.disable_receive_shadows = false
	return mat

func _on_world_info(info: Dictionary):
	world_info = info
	print("World info: ", info)

func _on_agents_update(agents: Array):
	# Clear existing agents and labels
	for child in agents_root.get_children():
		name_labels.remove_label(child)
		child.queue_free()
	
	# If terrain not loaded yet, store agents and wait
	if terrain_height_map.is_empty():
		pending_agents = agents.duplicate()
		return
	
	# Spawn agents now that we have terrain
	_spawn_agents(agents)

func _spawn_agents(agents: Array):
	for agent_data in agents:
		var agent = AGENT_SCENE.instantiate()
		
		# Get agent position from data
		var pos = agent_data.get("position", {})
		var x = pos.get("x", 0)
		var z = pos.get("y", 0)  # y in API is z in Godot
		
		# Get ground height at this position
		var ground_height = _get_ground_height(int(x), int(z))
		
		# Setup agent with adjusted height
		agent.setup_with_height(agent_data, ground_height)
		agents_root.add_child(agent)
		# Add name label
		name_labels.add_label(agent, agent_data.get("name", "Agent"), Color("#4fc3f7"))

func _on_map_received(map_data: Dictionary):
	if map_data.has("blocks"):
		_update_terrain(map_data["blocks"])
		# Now spawn pending agents if any
		if not pending_agents.is_empty():
			_spawn_agents(pending_agents)
			pending_agents.clear()
		
	if map_data.has("buildings"):
		_update_buildings(map_data["buildings"])

func _get_ground_height(x: int, z: int) -> float:
	var key = "%d,%d" % [x, z]
	# Default to 1.0 (single grass block height) if no terrain data
	return terrain_height_map.get(key, 1.0)

func _update_terrain(blocks: Array):
	# Clear existing terrain and height map
	terrain_height_map.clear()
	for child in terrain_root.get_children():
		child.queue_free()
	
	# Group blocks by type
	var blocks_by_type: Dictionary = {}
	for block in blocks:
		var type = block.get("terrainType", "grass")
		if not blocks_by_type.has(type):
			blocks_by_type[type] = []
		blocks_by_type[type].append(block)
	
	# Create MultiMesh for each terrain type (performance optimization)
	for type in blocks_by_type.keys():
		var type_blocks = blocks_by_type[type]
		var mesh_instance = MultiMeshInstance3D.new()
		
		var multimesh = MultiMesh.new()
		multimesh.transform_format = MultiMesh.TRANSFORM_3D
		
		# Create mesh with proper size
		var box_mesh = BoxMesh.new()
		box_mesh.size = Vector3(1, 1, 1)
		multimesh.mesh = box_mesh
		multimesh.instance_count = type_blocks.size()
		
		for i in range(type_blocks.size()):
			var block = type_blocks[i]
			var pos = block.get("position", {})
			var x = pos.get("x", 0)
			var y = pos.get("z", 0)  # height (API z is height)
			var z = pos.get("y", 0)  # API y is z in Godot
			
			# Update height map - keep maximum GROUND TOP height for each x,z
			# Block is 1m tall, centered at y+0.5, so top is at y+1
			var block_top = float(y) + 1.0
			var key = "%d,%d" % [int(x), int(z)]
			var current_height = terrain_height_map.get(key, -9999.0)
			if block_top > current_height:
				terrain_height_map[key] = block_top
			
			var transform = Transform3D()
			transform.origin = Vector3(x, y + 0.5, z)
			multimesh.set_instance_transform(i, transform)
		
		mesh_instance.multimesh = multimesh
		mesh_instance.material_override = terrain_materials.get(type, terrain_materials["grass"])
		terrain_root.add_child(mesh_instance)

func _update_buildings(buildings: Array):
	# Clear existing buildings and labels
	for child in buildings_root.get_children():
		name_labels.remove_label(child)
		child.queue_free()
	
	# Spawn new buildings
	for building_data in buildings:
		var building = BUILDING_SCENE.instantiate()
		building.setup(building_data)
		buildings_root.add_child(building)
		# Add name label
		name_labels.add_label(building, building_data.get("name", "Building"), Color("#4fc3f7"))
