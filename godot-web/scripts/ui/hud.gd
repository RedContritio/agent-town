extends Control

class_name HUD

@onready var api_client: ApiClient = get_tree().get_root().get_node("Main/WorldManager/ApiClient")

@onready var agent_count_label: Label = $TopLeftPanel/VBoxContainer/AgentCountLabel
@onready var building_count_label: Label = $TopLeftPanel/VBoxContainer/BuildingCountLabel
@onready var block_count_label: Label = $TopLeftPanel/VBoxContainer/BlockCountLabel
@onready var seed_label: Label = $TopLeftPanel/VBoxContainer/SeedLabel
@onready var status_label: Label = $StatusLabel
@onready var building_legend: VBoxContainer = $TopRightPanel/VBoxContainer/BuildingLegend

var world_info: Dictionary = {}
var agent_count: int = 0
var building_count: int = 0
var block_count: int = 0

func _ready():
	# Connect to API signals
	if api_client:
		api_client.world_info_received.connect(_on_world_info)
		api_client.agents_received.connect(_on_agents_received)
		api_client.error_occurred.connect(_on_error)
		api_client.fetch_world_info()
	
	# Create building legend
	_create_building_legend()
	
	# Initial status
	status_label.text = "Loading..."

func _on_world_info(info: Dictionary):
	world_info = info
	seed_label.text = "Seed: " + info.get("seed", "-")
	agent_count = info.get("agentCount", 0)
	building_count = info.get("buildingCount", 0)
	_update_counts()
	status_label.text = "Connected"
	
	# Fetch map data
	api_client.fetch_map(0, 0, 20)

func _on_agents_received(agents: Array):
	agent_count = agents.size()
	_update_counts()

func _on_error(error: String):
	status_label.text = "Error: " + error

func _update_counts():
	agent_count_label.text = "Agents: " + str(agent_count)
	building_count_label.text = "Buildings: " + str(building_count)
	block_count_label.text = "Blocks: " + str(block_count)

func _create_building_legend():
	# From docs/VISUAL_DESIGN.md - Appendix B: buildingColors
	var types = {
		"home": Color("#d4a574"),      # 暖木/居住感
		"shop": Color("#e8a87c"),      # 暖色零售
		"bank": Color("#4a90d9"),      # 冷蓝金融
		"exchange": Color("#50c878"),  # 绿交易
		"park": Color("#6b8e23"),      # 自然绿
		"office": Color("#8899aa"),    # 灰蓝办公
		"cafe": Color("#c4956a"),      # 暖褐餐饮
	}
	
	for type_name in types.keys():
		var hbox = HBoxContainer.new()
		hbox.add_theme_constant_override("separation", 8)
		
		var color_rect = ColorRect.new()
		color_rect.custom_minimum_size = Vector2(16, 16)
		color_rect.color = types[type_name]
		# Add rounded corners via stylebox
		var rect_style = StyleBoxFlat.new()
		rect_style.bg_color = types[type_name]
		rect_style.corner_radius_top_left = 4
		rect_style.corner_radius_top_right = 4
		rect_style.corner_radius_bottom_left = 4
		rect_style.corner_radius_bottom_right = 4
		color_rect.add_theme_stylebox_override("panel", rect_style)
		hbox.add_child(color_rect)
		
		var label = Label.new()
		label.text = type_name.capitalize()
		label.add_theme_color_override("font_color", Color("#e0e0e0"))
		label.add_theme_font_size_override("font_size", 13)
		hbox.add_child(label)
		
		building_legend.add_child(hbox)
