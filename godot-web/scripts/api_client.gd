extends Node

class_name ApiClient

const API_BASE_URL := "http://localhost:8080/api/v1"
const REFRESH_INTERVAL := 5.0  # seconds

signal world_info_received(info: Dictionary)
signal world_time_received(time_info: Dictionary)
signal agents_received(agents: Array)
signal agent_details_received(agent_id: String, details: Dictionary)
signal map_received(map_data: Dictionary)
signal error_occurred(error: String)

var http_client: HTTPClient
var refresh_timer: Timer
var is_requesting := false

func _ready():
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

func fetch_world_info():
	_make_request("/world/info", HTTPClient.Method.METHOD_GET, world_info_received.emit)

func fetch_world_time():
	_make_request("/world/time", HTTPClient.Method.METHOD_GET, world_time_received.emit)

func fetch_agents():
	_make_request("/agents", HTTPClient.Method.METHOD_GET, func(data):
		if data is Array:
			agents_received.emit(data)
	)

func fetch_agent_details(agent_id: String):
	_make_request("/agents/" + agent_id, HTTPClient.Method.METHOD_GET, func(data):
		agent_details_received.emit(agent_id, data)
	)

func fetch_map(x: int, y: int, radius: int):
	var query = "?x=%d&y=%d&radius=%d" % [x, y, radius]
	_make_request("/world/map" + query, HTTPClient.Method.METHOD_GET, map_received.emit)

func _make_request(endpoint: String, method: int, callback: Callable):
	is_requesting = true
	
	var http_request = HTTPRequest.new()
	add_child(http_request)
	http_request.request_completed.connect(
		func(result, response_code, headers, body):
			is_requesting = false
			if result == HTTPRequest.RESULT_SUCCESS and response_code == 200:
				var json = JSON.new()
				var error = json.parse(body.get_string_from_utf8())
				if error == OK:
					callback.call(json.data)
				else:
					error_occurred.emit("JSON parse error: " + str(error))
			else:
				error_occurred.emit("HTTP error: " + str(response_code))
			http_request.queue_free()
	)
	
	var headers = PackedStringArray(["Content-Type: application/json"])
	var url = API_BASE_URL + endpoint
	var error = http_request.request(url, headers, method)
	if error != OK:
		is_requesting = false
		error_occurred.emit("Request failed: " + str(error))
		http_request.queue_free()
