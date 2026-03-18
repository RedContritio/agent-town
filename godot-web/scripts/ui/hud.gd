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
	var types = {
		"house": Color(0.63, 0.5, 0.38),
		"shop": Color(0.8, 0.6, 0.4),
		"farm": Color(0.5, 0.7, 0.3),
		"factory": Color(0.5, 0.5, 0.55),
		"storage": Color(0.6, 0.55, 0.5),
	}
	
	for type_name in types.keys():
		var hbox = HBoxContainer.new()
		
		var color_rect = ColorRect.new()
		color_rect.custom_minimum_size = Vector2(16, 16)
		color_rect.color = types[type_name]
		hbox.add_child(color_rect)
		
		var label = Label.new()
		label.text = type_name.capitalize()
		hbox.add_child(label)
		
		building_legend.add_child(hbox)
