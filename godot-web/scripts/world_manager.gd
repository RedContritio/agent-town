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
var terrain_height_map: Dictionary = {}  # "x,z" -> surface height (top of top block)
var pending_agents: Array = []  # Store agents until terrain is loaded

func _ready():
	# Create terrain materials with visual design colors and unique characteristics
	# Colors from docs/VISUAL_DESIGN.md - Appendix B: tileColorMap
	
	# Grass: natural rough surface, matte finish
	terrain_materials["grass"] = _create_material(Color("#5a8f63"), 1.0, 0.0)
	
	# Road: asphalt-like, slightly rough, low specular
	terrain_materials["road"] = _create_material(Color("#c8b08a"), 0.85, 0.1)
	
	# Water: semi-transparent, reflective, smooth
	terrain_materials["water"] = _create_water_material(Color("#4a92b8"))
	
	# Farmland: earthy soil texture, rough
	terrain_materials["farmland"] = _create_material(Color("#9a7820"), 0.95, 0.05)
	
	# Sand: granular, some specular highlights
	terrain_materials["sand"] = _create_material(Color("#d4c088"), 0.7, 0.15)
	
	# Foundation: concrete-like, smooth
	terrain_materials["foundation"] = _create_material(Color("#6a6a6a"), 0.6, 0.1)
	
	# Hill: darker grass, rough
	terrain_materials["hill"] = _create_material(Color("#4a6a4a"), 1.0, 0.0)
	
	# Door: wooden, slight sheen
	terrain_materials["door"] = _create_material(Color("#b89955"), 0.5, 0.2)
	
	# Fence: weathered wood, rough
	terrain_materials["fence"] = _create_material(Color("#8a7254"), 0.9, 0.05)
	
	# Bridge: wood/stone, slightly polished
	terrain_materials["bridge"] = _create_material(Color("#a08050"), 0.55, 0.15)
	
	# Indoor floor: polished wood
	terrain_materials["indoor_floor"] = _create_material(Color("#c4a882"), 0.4, 0.25)
	
	# Indoor wall: plaster-like, smooth matte
	terrain_materials["indoor_wall"] = _create_material(Color("#e8e0d0"), 0.75, 0.05)
	
	# Indoor window: glass-like, transparent and reflective
	terrain_materials["indoor_window"] = _create_glass_material(Color("#add8e6"))
	
	# Dirt/Stone for underground filling
	terrain_materials["dirt"] = _create_material(Color("#5a4d3d"), 0.95, 0.0)
	terrain_materials["stone"] = _create_material(Color("#6a6a6a"), 0.8, 0.05)
	
	# Connect API signals
	api_client.world_info_received.connect(_on_world_info)
	api_client.agents_received.connect(_on_agents_update)
	api_client.map_received.connect(_on_map_received)

func _create_material(color: Color, roughness: float = 0.9, specular: float = 0.0) -> StandardMaterial3D:
	var mat = StandardMaterial3D.new()
	mat.albedo_color = color
	mat.roughness = roughness
	mat.metallic = 0.0
	mat.metallic_specular = specular
	mat.vertex_color_use_as_albedo = false
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_PER_PIXEL
	mat.specular_mode = BaseMaterial3D.SPECULAR_SCHLICK_GGX if specular > 0 else BaseMaterial3D.SPECULAR_DISABLED
	mat.disable_receive_shadows = false
	return mat

func _create_water_material(color: Color) -> StandardMaterial3D:
	var mat = StandardMaterial3D.new()
	mat.albedo_color = Color(color.r, color.g, color.b, 0.65)  # 65% opacity
	mat.roughness = 0.05  # Very smooth for reflections
	mat.metallic = 0.1
	mat.metallic_specular = 0.8  # High specular for water shine
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_PER_PIXEL
	mat.specular_mode = BaseMaterial3D.SPECULAR_SCHLICK_GGX
	mat.disable_receive_shadows = true  # Water receives shadows differently
	# Add some refraction-like effect
	mat.refraction_enabled = true
	mat.refraction_scale = 0.02
	return mat

func _create_glass_material(color: Color) -> StandardMaterial3D:
	var mat = StandardMaterial3D.new()
	mat.albedo_color = Color(color.r, color.g, color.b, 0.4)  # 40% opacity for glass
	mat.roughness = 0.1  # Smooth glass
	mat.metallic = 0.0
	mat.metallic_specular = 0.9  # High specular
	mat.transparency = BaseMaterial3D.TRANSPARENCY_ALPHA
	mat.shading_mode = BaseMaterial3D.SHADING_MODE_PER_PIXEL
	mat.specular_mode = BaseMaterial3D.SPECULAR_SCHLICK_GGX
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
	
	# Build a height map from surface blocks
	# Key: "x,z" -> {height: int, type: string}
	var surface_blocks: Dictionary = {}
	
	for block in blocks:
		var pos = block.get("position", {})
		var x = pos.get("x", 0)
		var y = pos.get("z", 0)  # height (API z is height)
		var z = pos.get("y", 0)  # API y is z in Godot
		var type = block.get("terrainType", "grass")
		
		var key = "%d,%d" % [int(x), int(z)]
		
		# Track the highest block at this x,z position
		if not surface_blocks.has(key):
			surface_blocks[key] = {"x": x, "y": y, "z": z, "type": type}
		else:
			var existing = surface_blocks[key]
			if y > existing.y:
				surface_blocks[key] = {"x": x, "y": y, "z": z, "type": type}
		
		# Store surface height for agent spawning
		var block_top = float(y) + 1.0
		var current_top = terrain_height_map.get(key, -9999.0)
		if block_top > current_top:
			terrain_height_map[key] = block_top
	
	# Now we need to create columns of blocks from bottom up to surface
	# For performance, we'll use MultiMesh for each type
	
	# Collect blocks by type for rendering
	var blocks_by_type: Dictionary = {}
	
	for key in surface_blocks.keys():
		var surface = surface_blocks[key]
		var x = surface.x
		var surface_y = surface.y
		var z = surface.z
		var surface_type = surface.type
		
		# For water, render at level 0 as a surface
		# For other types, render a column from y=0 up to surface_y
		if surface_type == "water":
			# Water is special: render at y=0 (surface level)
			_add_block_to_batch(blocks_by_type, "water", x, 0, z)
		else:
			# For land, render surface block at surface_y
			_add_block_to_batch(blocks_by_type, surface_type, x, surface_y, z)
			
			# Fill below with dirt/stone if above ground level
			# This creates the "column" look
			for fill_y in range(0, surface_y):
				var fill_type = "stone" if fill_y < -1 else "dirt"
				_add_block_to_batch(blocks_by_type, fill_type, x, fill_y, z)
	
	# Create MultiMesh instances for each type
	for type in blocks_by_type.keys():
		var type_blocks = blocks_by_type[type]
		if type_blocks.size() == 0:
			continue
			
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
			var transform = Transform3D()
			transform.origin = Vector3(block.x, block.y + 0.5, block.z)
			multimesh.set_instance_transform(i, transform)
		
		mesh_instance.multimesh = multimesh
		mesh_instance.material_override = terrain_materials.get(type, terrain_materials["grass"])
		terrain_root.add_child(mesh_instance)

func _add_block_to_batch(batch: Dictionary, type: String, x: int, y: int, z: int):
	if not batch.has(type):
		batch[type] = []
	batch[type].append({"x": x, "y": y, "z": z})

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
