extends StaticBody3D

class_name BuildingEntity

var building_id: String
var building_name: String
var building_type: String

@onready var mesh: MeshInstance3D = $Mesh
@onready var roof: MeshInstance3D = $Roof
@onready var label: Label3D = $Label3D

const TYPE_COLORS = {
	"house": Color(0.63, 0.5, 0.38),
	"shop": Color(0.8, 0.6, 0.4),
	"farm": Color(0.5, 0.7, 0.3),
	"factory": Color(0.5, 0.5, 0.55),
	"storage": Color(0.6, 0.55, 0.5),
}

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
	
	# Update building mesh
	if mesh:
		mesh.mesh = BoxMesh.new()
		mesh.mesh.size = Vector3(width, height, depth)
		mesh.position.y = height / 2
		
		var color = TYPE_COLORS.get(building_type, TYPE_COLORS["house"])
		var material = StandardMaterial3D.new()
		material.albedo_color = color
		mesh.material_override = material
	
	# Update roof
	if roof:
		roof.mesh = PrismMesh.new()
		roof.mesh.size = Vector3(width * 0.8, height * 0.5, depth * 0.8)
		roof.position.y = height + height * 0.25
		roof.rotation.y = PI / 4
		
		var color = TYPE_COLORS.get(building_type, TYPE_COLORS["house"])
		var material = StandardMaterial3D.new()
		material.albedo_color = color.darkened(0.1)
		roof.material_override = material
	
	# Update label
	if label:
		label.text = building_name
		label.position.y = height + 1.5
