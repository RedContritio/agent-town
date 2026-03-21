extends Node

class_name ApiClient

# API 配置从项目设置读取，可通过导出预设或命令行覆盖
var API_BASE_URL: String
var REFRESH_INTERVAL: float

# 认证 Token
var auth_token: String = ""
var current_agent_id: int = -1
var current_agent_name: String = ""

signal world_info_received(info: Dictionary)
signal world_time_received(time_info: Dictionary)
signal agents_received(agents: Array)
signal agent_details_received(agent_id: String, details: Dictionary)
signal map_received(map_data: Dictionary)
signal agent_status_received(status: Dictionary)
signal task_stack_received(stack_data: Dictionary)
signal auth_success(agent_info: Dictionary)
signal auth_failed(error: String)
signal error_occurred(error: String)

var http_client: HTTPClient
var refresh_timer: Timer
var is_requesting := false

func _ready():
	# 从项目设置读取配置，使用默认值
	API_BASE_URL = ProjectSettings.get_setting("custom/api/base_url", "http://localhost:8080/api/v1")
	REFRESH_INTERVAL = ProjectSettings.get_setting("custom/api/refresh_interval", 5.0)
	
	print("API Client initialized with base URL: ", API_BASE_URL)
	
	http_client = HTTPClient.new()
	
	# Create refresh timer
	refresh_timer = Timer.new()
	refresh_timer.wait_time = REFRESH_INTERVAL
	refresh_timer.timeout.connect(_on_refresh_timer)
	add_child(refresh_timer)
	refresh_timer.start()
	
	# Initial fetch
	fetch_world_info()
	fetch_agents()

func _on_refresh_timer():
	if not is_requesting:
		fetch_world_info()
		fetch_agents()
		# 如果已认证，同时刷新任务栈
		if is_authenticated():
			fetch_task_stack()

# ==================== 认证相关 ====================

func is_authenticated() -> bool:
	return auth_token != "" and current_agent_id != -1

func set_auth_token(token: String):
	auth_token = token
	print("Auth token set")

func authenticate_with_token(token: String):
	set_auth_token(token)
	# 验证 token 并获取 agent 信息
	fetch_agent_me()

func clear_auth():
	auth_token = ""
	current_agent_id = -1
	current_agent_name = ""

# ==================== API 请求方法 ====================

func fetch_world_info():
	_make_request("/world/info", HTTPClient.Method.METHOD_GET, func(data):
		world_info_received.emit(data)
	)

func fetch_world_time():
	_make_request("/world/time", HTTPClient.Method.METHOD_GET, func(data):
		world_time_received.emit(data)
	)

func fetch_agents():
	_make_request("/agents", HTTPClient.Method.METHOD_GET, func(data):
		if data is Array:
			agents_received.emit(data)
	)

func fetch_agent_details(agent_id: String):
	_make_request("/agents/" + agent_id, HTTPClient.Method.METHOD_GET, func(data):
		agent_details_received.emit(agent_id, data)
	)

# 旧的 fetch_map 已弃用，使用 fetch_chunks 替代
func fetch_map(x: int, y: int, radius: int):
	# 转换为 chunk 坐标并调用新的 fetch_chunks
	var chunk_size = 32
	var cx = int(floor(float(x) / chunk_size))
	var cy = int(floor(float(y) / chunk_size))
	var cr = int(ceil(float(radius) / chunk_size))
	fetch_chunks(cx, cy, cr)

# 新的 chunk-based 地图获取
func fetch_chunks(cx: int, cy: int, chunk_radius: int):
	var query = "?cx=%d&cy=%d&cr=%d" % [cx, cy, chunk_radius]
	_make_request("/world/map" + query, HTTPClient.Method.METHOD_GET, func(data):
		map_received.emit(data)
	)

# ==================== 认证 API ====================

func fetch_agent_me():
	if not is_authenticated():
		return
	_make_authenticated_request("/agents/me", HTTPClient.Method.METHOD_GET, func(data):
		if data.has("id"):
			current_agent_id = data.get("id", -1)
			current_agent_name = data.get("name", "")
			auth_success.emit(data)
			# 获取成功后拉取任务栈
			fetch_task_stack()
		else:
			auth_failed.emit("Invalid response")
	, func(error):
		auth_failed.emit(error)
	)

func fetch_agent_status():
	if not is_authenticated():
		return
	_make_authenticated_request("/agents/me/status", HTTPClient.Method.METHOD_GET, agent_status_received.emit)

func fetch_task_stack():
	if not is_authenticated():
		return
	_make_authenticated_request("/agents/me/tasks", HTTPClient.Method.METHOD_GET, task_stack_received.emit)

# ==================== 任务操作 API ====================

func create_task(task_type: String, params: Dictionary):
	if not is_authenticated():
		error_occurred.emit("Not authenticated")
		return
	
	var type_id = _task_type_to_id(task_type)
	var body = {
		"type": type_id,
		"params": params
	}
	_make_authenticated_request("/agents/me/tasks", HTTPClient.Method.METHOD_POST, func(data):
		print("Task created: ", data)
		# 刷新任务栈
		fetch_task_stack()
	, Callable(), body)

func pause_task(task_id: String):
	if not is_authenticated():
		return
	var endpoint = "/agents/me/tasks/" + task_id + "/pause"
	_make_authenticated_request(endpoint, HTTPClient.Method.METHOD_POST, func(data):
		print("Task paused: ", task_id)
		fetch_task_stack()
	)

func resume_task(task_id: String):
	if not is_authenticated():
		return
	var endpoint = "/agents/me/tasks/" + task_id + "/resume"
	_make_authenticated_request(endpoint, HTTPClient.Method.METHOD_POST, func(data):
		print("Task resumed: ", task_id)
		fetch_task_stack()
	)

func drop_task(task_id: String):
	if not is_authenticated():
		return
	var endpoint = "/agents/me/tasks/" + task_id
	_make_authenticated_request(endpoint, HTTPClient.Method.METHOD_DELETE, func(data):
		print("Task dropped: ", task_id)
		fetch_task_stack()
	)

# ==================== HTTP 请求工具 ====================

func _make_request(endpoint: String, method: int, callback: Callable):
	var headers = PackedStringArray(["Content-Type: application/json"])
	_perform_request(endpoint, method, headers, callback)

func _make_authenticated_request(endpoint: String, method: int, callback: Callable, error_callback: Callable = Callable(), body: Dictionary = {}):
	if auth_token == "":
		if error_callback:
			error_callback.call("Not authenticated")
		return
	
	var headers = PackedStringArray([
		"Content-Type: application/json",
		"Authorization: Bearer " + auth_token
	])
	_perform_request(endpoint, method, headers, callback, error_callback, body)

func _perform_request(endpoint: String, method: int, headers: PackedStringArray, callback: Callable, error_callback: Callable = Callable(), body: Dictionary = {}):
	is_requesting = true
	
	var http_request = HTTPRequest.new()
	add_child(http_request)
	http_request.request_completed.connect(
		func(result, response_code, response_headers, response_body):
			is_requesting = false
			if result == HTTPRequest.RESULT_SUCCESS and response_code >= 200 and response_code < 300:
				var json = JSON.new()
				var error = json.parse(response_body.get_string_from_utf8())
				if error == OK:
					callback.call(json.data)
				else:
					if error_callback:
						error_callback.call("JSON parse error: " + str(error))
					else:
						error_occurred.emit("JSON parse error: " + str(error))
			else:
				var error_msg = "HTTP error: " + str(response_code)
				if error_callback:
					error_callback.call(error_msg)
				else:
					error_occurred.emit(error_msg)
			http_request.queue_free()
	)
	
	var url = API_BASE_URL + endpoint
	var error: int
	
	if body.is_empty():
		error = http_request.request(url, headers, method)
	else:
		var json_body = JSON.stringify(body)
		error = http_request.request(url, headers, method, json_body)
	
	if error != OK:
		is_requesting = false
		if error_callback:
			error_callback.call("Request failed: " + str(error))
		else:
			error_occurred.emit("Request failed: " + str(error))
		http_request.queue_free()

func _task_type_to_id(task_type: String) -> int:
	match task_type:
		"move": return 0
		"harvest": return 1
		"craft": return 2
		"build": return 3
		"combat": return 4
		_: return 0

func _task_status_to_string(status: int) -> String:
	match status:
		0: return "pending"
		1: return "running"
		2: return "paused"
		3: return "completed"
		4: return "failed"
		_: return "unknown"
