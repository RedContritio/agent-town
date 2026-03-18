extends StaticBody3D

class_name BuildingEntity

var building_id: String
var building_name: String
var building_type: String

@onready var mesh: MeshInstance3D = $Mesh
@onready var roof: MeshInstance3D = $Roof
@onready var label: Label3D = $Label3D

const TYPE_COLORS = {
	# From docs/VISUAL_DESIGN.md - Appendix B: buildingColors
	"home": Color("#d4a574"),      # 暖木/居住感
	"shop": Color("#e8a87c"),      # 暖色、亲和零售感
	"bank": Color("#4a90d9"),      # 偏冷蓝、稳重金融感
	"exchange": Color("#50c878"),  # 清爽绿、交易/流通感
	"park": Color("#6b8e23"),      # 自然绿、开放空间
	"office": Color("#8899aa"),    # 中性灰蓝、办公感
	"cafe": Color("#c4956a"),      # 暖褐、休闲餐饮
	# Legacy mappings for backward compatibility
	"house": Color("#d4a574"),     # -> home
	"farm": Color("#6b8e23"),      # -> park
	"factory": Color("#8899aa"),   # -> office
	"storage": Color("#998c80"),   # warehouse-like
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
