extends StaticBody3D

class_name BuildingEntity

# Debug flag - can be toggled globally to show/hide debug markers
static var DEBUG_SHOW_MARKERS: bool = true

var building_id: String
var building_name: String
var building_type: String

var _mesh: MeshInstance3D
var _roof: MeshInstance3D
var _label: Label3D
var _wall_marker: MeshInstance3D

const TYPE_COLORS = {
	"home": Color("#d4a574"),
	"shop": Color("#ff8c00"),
	"bank": Color("#4a90d9"),
	"exchange": Color("#50c878"),
	"park": Color("#6b8e23"),
	"office": Color("#8899aa"),
	"cafe": Color("#c4956a"),
	"gov_hall": Color("#8b4513"),
	"guide_hall": Color("#4169e1"),
	"quest": Color("#228b22"),
	"house": Color("#d4a574"),
	"farm": Color("#6b8e23"),
	"factory": Color("#8899aa"),
	"storage": Color("#998c80"),
	"town_hall": Color("#8b4513"),
	"quest_board": Color("#228b22"),
}

func _ready():
	pass

func setup(data: Dictionary):
	building_id = data.get("id", "")
	building_name = data.get("name", "Building")
	building_type = data.get("type", "house")
	
	var pos = data.get("anchor", {})
	position = Vector3(
		pos.get("x", 0),
		0,
		pos.get("y", 0)
	)
	
	var width = data.get("width", 3)
	var height = data.get("height", 3)
	var depth = data.get("depth", 3)
	
	# Create building mesh
	_mesh = MeshInstance3D.new()
	_mesh.name = "Mesh"
	add_child(_mesh)
	
	_mesh.mesh = BoxMesh.new()
	_mesh.mesh.size = Vector3(width, height, depth)
	_mesh.position.y = height / 2
	
	var color = TYPE_COLORS.get(building_type, TYPE_COLORS["house"])
	var material = StandardMaterial3D.new()
	material.albedo_color = color
	_mesh.material_override = material
	
	# Create roof
	_roof = MeshInstance3D.new()
	_roof.name = "Roof"
	add_child(_roof)
	
	_roof.mesh = PrismMesh.new()
	_roof.mesh.size = Vector3(width * 0.8, height * 0.5, depth * 0.8)
	_roof.position.y = height + height * 0.25
	_roof.rotation.y = PI / 4
	
	var roof_color = TYPE_COLORS.get(building_type, TYPE_COLORS["house"])
	var roof_material = StandardMaterial3D.new()
	roof_material.albedo_color = roof_color.darkened(0.1)
	_roof.material_override = roof_material
	
	# Create Label3D
	_create_label(height)
	
	# Create debug markers
	_create_debug_markers(width, height, depth)

func _create_label(building_height: float):
	_label = Label3D.new()
	_label.name = "NameLabel"
	add_child(_label)
	
	# Use English fallback for visibility testing
	var display_name = building_name
	if display_name == "政府大厅":
		display_name = "Gov Hall"
	elif display_name == "引导大厅":
		display_name = "Guide Hall"
	elif display_name == "委托处":
		display_name = "Quest"
	elif display_name == "商店":
		display_name = "Shop"
	
	_label.text = display_name
	_label.position = Vector3(0, building_height + 2.0, 0)
	_label.modulate = Color(0.2, 0.9, 1.0, 1.0)
	_label.font_size = 72
	_label.pixel_size = 0.02
	_label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	_label.no_depth_test = true
	_label.visible = true

func _create_debug_markers(width: float, height: float, depth: float):
	var is_target = (building_type == "gov_hall" or building_type == "town_hall")
	var should_show = DEBUG_SHOW_MARKERS and is_target
	
	if not should_show:
		return
	
	# Wall marker - red triangle on north wall (only visible from north)
	_wall_marker = MeshInstance3D.new()
	_wall_marker.name = "WallMarker"
	add_child(_wall_marker)
	
	var triangle_mesh = _create_triangle_mesh()
	_wall_marker.mesh = triangle_mesh
	# 贴在北墙表面（像窗户一样）
	_wall_marker.position = Vector3(0, height * 0.6, -depth/2 + 0.01)
	_wall_marker.rotation = Vector3(-PI/2, 0, 0)
	
	var wall_material = StandardMaterial3D.new()
	wall_material.albedo_color = Color(1, 0, 0)
	wall_material.emission_enabled = true
	wall_material.emission = Color(1, 0, 0)
	wall_material.emission_energy = 3.0
	wall_material.cull_mode = BaseMaterial3D.CULL_DISABLED
	_wall_marker.material_override = wall_material
	_wall_marker.visible = true

func _create_triangle_mesh() -> ArrayMesh:
	var mesh = ArrayMesh.new()
	var arrays = []
	arrays.resize(Mesh.ARRAY_MAX)
	
	var vertices = PackedVector3Array([
		Vector3(0, 0.8, 0),
		Vector3(-0.6, -0.6, 0),
		Vector3(0.6, -0.6, 0),
	])
	
	var normals = PackedVector3Array([
		Vector3(0, 0, 1),
		Vector3(0, 0, 1),
		Vector3(0, 0, 1),
	])
	
	arrays[Mesh.ARRAY_VERTEX] = vertices
	arrays[Mesh.ARRAY_NORMAL] = normals
	arrays[Mesh.ARRAY_INDEX] = PackedInt32Array([0, 1, 2])
	
	mesh.add_surface_from_arrays(Mesh.PRIMITIVE_TRIANGLES, arrays)
	return mesh

static func set_debug_markers_enabled(enabled: bool):
	DEBUG_SHOW_MARKERS = enabled
