extends PanelContainer

class_name AgentPanel

@onready var api_client: ApiClient = get_tree().get_root().get_node("Main/WorldManager/ApiClient")

# 信息面板
@onready var agent_name_label: Label = $VBoxContainer/InfoSection/AgentNameLabel
@onready var position_label: Label = $VBoxContainer/InfoSection/PositionLabel
@onready var hp_bar: ProgressBar = $VBoxContainer/InfoSection/HPBar
@onready var hp_label: Label = $VBoxContainer/InfoSection/HPBar/HPLabel
@onready var stamina_bar: ProgressBar = $VBoxContainer/InfoSection/StaminaBar
@onready var stamina_label: Label = $VBoxContainer/InfoSection/StaminaBar/StaminaLabel
@onready var balance_label: Label = $VBoxContainer/InfoSection/BalanceLabel

# 任务栈面板
@onready var task_list: VBoxContainer = $VBoxContainer/TaskSection/ScrollContainer/TaskList
@onready var no_auth_label: Label = $VBoxContainer/TaskSection/NoAuthLabel

# 操作按钮
@onready var refresh_button: Button = $VBoxContainer/ButtonSection/RefreshButton
@onready var create_task_button: Button = $VBoxContainer/ButtonSection/CreateTaskButton

var current_tasks: Array = []
var selected_agent_id: int = -1

func _ready():
	hide()
	
	# 连接 API 信号
	if api_client:
		api_client.auth_success.connect(_on_auth_success)
		api_client.auth_failed.connect(_on_auth_failed)
		api_client.agent_status_received.connect(_on_agent_status)
		api_client.task_stack_received.connect(_on_task_stack)
	
	# 连接按钮信号
	refresh_button.pressed.connect(_on_refresh_pressed)
	create_task_button.pressed.connect(_on_create_task_pressed)

func show_for_agent(agent_data: Dictionary):
	selected_agent_id = agent_data.get("id", -1)
	var agent_name = agent_data.get("name", "Unknown")
	
	agent_name_label.text = "Agent: " + agent_name
	_update_position(agent_data.get("position", {}))
	_update_stats(agent_data)
	
	# 检查是否是当前认证的 Agent
	if api_client.is_authenticated() and api_client.current_agent_id == selected_agent_id:
		no_auth_label.hide()
		task_list.show()
		create_task_button.show()
		api_client.fetch_task_stack()
	else:
		no_auth_label.text = "Login with CLI token to view tasks"
		no_auth_label.show()
		task_list.hide()
		create_task_button.hide()
	
	show()

func _update_position(pos: Dictionary):
	var x = pos.get("x", 0)
	var y = pos.get("y", 0)
	position_label.text = "Position: (%d, %d)" % [x, y]

func _update_stats(agent_data: Dictionary):
	var hp = agent_data.get("hp", 0)
	var max_hp = agent_data.get("maxHp", 100)
	hp_bar.max_value = max_hp
	hp_bar.value = hp
	hp_label.text = "%d/%d" % [hp, max_hp]
	
	var stamina = agent_data.get("stamina", 0)
	var max_stamina = agent_data.get("maxStamina", 100)
	stamina_bar.max_value = max_stamina
	stamina_bar.value = stamina
	stamina_label.text = "%d/%d" % [stamina, max_stamina]
	
	var balance = agent_data.get("balance", 0)
	balance_label.text = "Balance: %d" % balance

func _on_auth_success(agent_info: Dictionary):
	print("Agent authenticated: ", agent_info.get("name", ""))
	# 如果当前显示的是这个 agent，刷新任务列表
	if selected_agent_id == agent_info.get("id", -1):
		show_for_agent(agent_info)

func _on_auth_failed(error: String):
	print("Auth failed: ", error)

func _on_agent_status(status: Dictionary):
	_update_stats(status)

func _on_task_stack(stack_data: Dictionary):
	_clear_task_list()
	current_tasks = stack_data.get("stack", [])
	
	if current_tasks.is_empty():
		var empty_label = Label.new()
		empty_label.text = "No active tasks"
		empty_label.add_theme_color_override("font_color", Color("#888888"))
		task_list.add_child(empty_label)
	else:
		for task in current_tasks:
			_create_task_item(task)

func _clear_task_list():
	for child in task_list.get_children():
		child.queue_free()

func _create_task_item(task: Dictionary):
	var task_id = task.get("task_id", "unknown")
	var task_type = _get_task_type_name(task.get("type", 0))
	var status = task.get("status", 0)
	var status_str = _get_status_name(status)
	var depth = task.get("stack_depth", 0)
	
	var hbox = HBoxContainer.new()
	hbox.add_theme_constant_override("separation", 8)
	
	# 深度指示器
	var depth_label = Label.new()
	depth_label.text = str(depth)
	depth_label.custom_minimum_size = Vector2(20, 0)
	depth_label.add_theme_color_override("font_color", Color("#4fc3f7"))
	hbox.add_child(depth_label)
	
	# 任务ID
	var id_label = Label.new()
	id_label.text = task_id
	id_label.custom_minimum_size = Vector2(80, 0)
	id_label.add_theme_color_override("font_color", Color("#e0e0e0"))
	hbox.add_child(id_label)
	
	# 类型
	var type_label = Label.new()
	type_label.text = task_type
	type_label.custom_minimum_size = Vector2(60, 0)
	type_label.add_theme_color_override("font_color", Color("#c0c0c0"))
	hbox.add_child(type_label)
	
	# 状态
	var status_label = Label.new()
	status_label.text = status_str
	status_label.add_theme_color_override("font_color", _get_status_color(status))
	hbox.add_child(status_label)
	
	# 操作按钮
	if depth == 0:  # 栈顶任务
		if status == 1:  # running
			var pause_btn = Button.new()
			pause_btn.text = "Pause"
			pause_btn.pressed.connect(func(): api_client.pause_task(task_id))
			hbox.add_child(pause_btn)
		elif status == 2:  # paused
			var resume_btn = Button.new()
			resume_btn.text = "Resume"
			resume_btn.pressed.connect(func(): api_client.resume_task(task_id))
			hbox.add_child(resume_btn)
		
		var drop_btn = Button.new()
		drop_btn.text = "Drop"
		drop_btn.pressed.connect(func(): api_client.drop_task(task_id))
		hbox.add_child(drop_btn)
	
	task_list.add_child(hbox)

func _on_refresh_pressed():
	if api_client.is_authenticated():
		api_client.fetch_agent_status()
		api_client.fetch_task_stack()

func _on_create_task_pressed():
	# 简单的创建任务对话框
	_create_task_dialog()

func _create_task_dialog():
	var dialog = AcceptDialog.new()
	dialog.title = "Create Task"
	
	var vbox = VBoxContainer.new()
	
	# 任务类型选择
	var type_hbox = HBoxContainer.new()
	var type_label = Label.new()
	type_label.text = "Type:"
	type_hbox.add_child(type_label)
	
	var type_option = OptionButton.new()
	type_option.add_item("Move")
	type_option.add_item("Harvest")
	type_option.add_item("Craft")
	type_option.add_item("Build")
	type_option.select(0)
	type_hbox.add_child(type_option)
	vbox.add_child(type_hbox)
	
	# 参数输入
	var param_hbox = HBoxContainer.new()
	var param_label = Label.new()
	param_label.text = "Params (JSON):"
	param_hbox.add_child(param_label)
	
	var param_input = LineEdit.new()
	param_input.text = '{"dx": 1, "dy": 0}'
	param_input.custom_minimum_size = Vector2(200, 0)
	param_hbox.add_child(param_input)
	vbox.add_child(param_hbox)
	
	dialog.add_child(vbox)
	dialog.add_button("Create", true)
	
	dialog.confirmed.connect(func():
		var task_type = type_option.get_item_text(type_option.selected).to_lower()
		var json = JSON.new()
		var error = json.parse(param_input.text)
		if error == OK:
			api_client.create_task(task_type, json.data)
		else:
			print("Invalid JSON params")
		dialog.queue_free()
	)
	
	get_tree().root.add_child(dialog)
	dialog.popup_centered()

func _get_task_type_name(type_id: int) -> String:
	match type_id:
		0: return "move"
		1: return "harvest"
		2: return "craft"
		3: return "build"
		4: return "combat"
		_: return "unknown"

func _get_status_name(status: int) -> String:
	match status:
		0: return "pending"
		1: return "running"
		2: return "paused"
		3: return "completed"
		4: return "failed"
		_: return "unknown"

func _get_status_color(status: int) -> Color:
	match status:
		0: return Color("#888888")  # pending - gray
		1: return Color("#4fc3f7")  # running - cyan
		2: return Color("#ffaa00")  # paused - orange
		3: return Color("#50c878")  # completed - green
		4: return Color("#ff5555")  # failed - red
		_: return Color("#ffffff")
