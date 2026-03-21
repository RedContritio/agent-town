extends PanelContainer

class_name LoginPanel

@onready var api_client: ApiClient = get_tree().get_root().get_node("Main/WorldManager/ApiClient")

@onready var title_label: Label = $VBoxContainer/TitleLabel
@onready var token_input: LineEdit = $VBoxContainer/TokenInput
@onready var login_button: Button = $VBoxContainer/LoginButton
@onready var status_label: Label = $VBoxContainer/StatusLabel
@onready var logout_button: Button = $VBoxContainer/LogoutButton

func _ready():
	# 连接 API 信号
	if api_client:
		api_client.auth_success.connect(_on_auth_success)
		api_client.auth_failed.connect(_on_auth_failed)
	
	# 连接按钮信号
	login_button.pressed.connect(_on_login_pressed)
	logout_button.pressed.connect(_on_logout_pressed)
	token_input.text_submitted.connect(func(text): _on_login_pressed())
	
	# 初始状态
	_update_ui_state()

func _on_login_pressed():
	var token = token_input.text.strip_edges()
	if token == "":
		status_label.text = "Please enter a token"
		status_label.add_theme_color_override("font_color", Color("#ff5555"))
		return
	
	status_label.text = "Authenticating..."
	status_label.add_theme_color_override("font_color", Color("#4fc3f7"))
	
	api_client.authenticate_with_token(token)

func _on_logout_pressed():
	api_client.clear_auth()
	token_input.text = ""
	_update_ui_state()
	status_label.text = "Logged out"
	status_label.add_theme_color_override("font_color", Color("#888888"))

func _on_auth_success(agent_info: Dictionary):
	status_label.text = "Logged in as: " + agent_info.get("name", "Unknown")
	status_label.add_theme_color_override("font_color", Color("#50c878"))
	_update_ui_state()

func _on_auth_failed(error: String):
	status_label.text = "Auth failed: " + error
	status_label.add_theme_color_override("font_color", Color("#ff5555"))

func _update_ui_state():
	if api_client.is_authenticated():
		title_label.text = "Authenticated"
		token_input.hide()
		login_button.hide()
		logout_button.show()
	else:
		title_label.text = "Agent Login"
		token_input.show()
		login_button.show()
		logout_button.hide()

func show_panel():
	show()
	if not api_client.is_authenticated():
		token_input.grab_focus()
