extends Node

# Debug Controller - WebSocket 连接到 debug_server 实现调试控制

@export var enabled: bool = true
@export var server_url: String = "ws://localhost:8081/godot"

var camera_controller: Node = null
var socket: WebSocketPeer = null
var _connected: bool = false
var _message_queue: Array = []
var _log_file: FileAccess = null

# 预设配置
var _presets = {
	"top": {"distance": 30.0, "azimuth": 0.0, "polar": 10.0},
	"side": {"distance": 25.0, "azimuth": 0.0, "polar": 90.0},
	"north": {"distance": 15.0, "azimuth": -90.0, "polar": 85.0},
	"south": {"distance": 15.0, "azimuth": 90.0, "polar": 85.0},
	"east": {"distance": 15.0, "azimuth": 180.0, "polar": 85.0},
	"west": {"distance": 15.0, "azimuth": 0.0, "polar": 85.0},
}

func _ready():
	if not enabled:
		return
	
	# 延迟初始化，等待相机准备好
	var timer = get_tree().create_timer(1.0)
	timer.timeout.connect(_initialize)

func _initialize():
	print(" Initializing...")
	
	# 延迟查找相机
	var timer = get_tree().create_timer(3.0)
	timer.timeout.connect(_find_camera)
	
	# 同时连接 WebSocket
	_connect_to_server()

func _find_camera():
	var camera = get_viewport().get_camera_3d()
	if camera:
		camera_controller = camera
		if camera_controller.has_method("update_camera_position"):
			print(" Camera controller found, sending ready")
			_send_ready()
			return
	
	print(" Camera not ready, retrying...")
	var timer = get_tree().create_timer(2.0)
	timer.timeout.connect(_find_camera)

func _send_ready():
	_send_message({
		"type": "ready",
		"payload": {
			"camera_ready": true,
			"initial_target": [
				camera_controller.target_position.x,
				camera_controller.target_position.y,
				camera_controller.target_position.z
			]
		}
	})

func _connect_to_server():
	socket = WebSocketPeer.new()
	var err = socket.connect_to_url(server_url)
	if err != OK:
		print(" Failed to connect to server: " + str(err))
		# 重试
		var timer = get_tree().create_timer(3.0)
		timer.timeout.connect(_connect_to_server)
		return
	print(" Connecting to " + server_url)

func _process(delta):
	if socket:
		_socket_poll()
		
		# 检查是否有待处理的消息（调试用）
		if _connected and socket.get_available_packets() > 0:
			print(" Has packets available")

func _socket_poll():
	socket.poll()
	
	var state = socket.get_ready_state()
	
	match state:
		WebSocketPeer.STATE_CONNECTING:
			pass
			
		WebSocketPeer.STATE_OPEN:
			if not _connected:
				_connected = true
				print(" Connected to debug server")
				_flush_message_queue()
			
			# 接收消息
			while socket.get_available_packets() > 0:
				var packet = socket.get_packet()
				var text = packet.get_string_from_utf8()
				print(" RAW PACKET: " + text)
				_handle_message(text)
			
		WebSocketPeer.STATE_CLOSING:
			pass
			
		WebSocketPeer.STATE_CLOSED:
			if _connected:
				_connected = false
				print(" Disconnected from server, reconnecting...")
				var timer = get_tree().create_timer(3.0)
				timer.timeout.connect(_connect_to_server)

func _handle_message(text: String):
	var json = JSON.new()
	var err = json.parse(text)
	if err != OK:
		print(" JSON parse error: ", json.get_error_message())
		return
	
	var msg = json.get_data()
	if typeof(msg) != TYPE_DICTIONARY:
		return
	
	var msg_type = msg.get("type", "")
	var payload = msg.get("payload", {})
	
	print(" Received: ", msg_type)
	
	match msg_type:
		"set_camera":
			print(" Handling set_camera")
			_apply_camera_command(payload)
			
		"set_preset":
			print(" Handling set_preset")
			_apply_preset_command(payload.get("preset", ""))
			
		"capture_screenshot":
			print(" Handling capture_screenshot")
			_capture_screenshot(payload)
			
		"get_camera_info":
			print(" Handling get_camera_info")
			_send_camera_info()
			
		"connected":
			print(" Server acknowledged connection")
		
		_:
			print(" Unknown message type: ", msg_type)

func _apply_camera_command(payload):
	if not camera_controller:
		_send_error("camera_not_ready")
		return
	
	var target = payload.get("target", [0, 1.5, 0])
	var distance = payload.get("distance", 25.0)
	var azimuth = payload.get("azimuth", -45.0)
	var polar = payload.get("polar", 60.0)
	
	camera_controller.target_position = Vector3(target[0], target[1], target[2])
	camera_controller.distance = distance
	camera_controller.azimuth = deg_to_rad(azimuth)
	camera_controller.polar = deg_to_rad(polar)
	camera_controller.update_camera_position()
	
	print(" Camera set: target=", target, " distance=", distance, 
		" azimuth=", azimuth, " polar=", polar)
	
	_send_ack("set_camera")

func _apply_preset_command(preset_name: String):
	if not camera_controller:
		_send_error("camera_not_ready")
		return
	
	if not _presets.has(preset_name):
		_send_error("unknown_preset")
		return
	
	var p = _presets[preset_name]
	camera_controller.distance = p["distance"]
	camera_controller.azimuth = deg_to_rad(p["azimuth"])
	camera_controller.polar = deg_to_rad(p["polar"])
	camera_controller.update_camera_position()
	
	print(" Preset applied: ", preset_name)
	_send_ack("set_preset")

func _capture_screenshot(payload):
	var width = payload.get("width", 1280)
	var height = payload.get("height", 720)
	
	# 获取 viewport 截图
	var viewport = get_viewport()
	var img = viewport.get_texture().get_image()
	
	# 调整大小
	if img.get_width() != width or img.get_height() != height:
		img.resize(width, height)
	
	# 生成带时间戳的文件名
	var datetime = Time.get_datetime_dict_from_system()
	var timestamp = "%04d%02d%02d_%02d%02d%02d" % [
		datetime.year, datetime.month, datetime.day,
		datetime.hour, datetime.minute, datetime.second
	]
	var filename = "agent_town_screenshot_%s.png" % timestamp
	var filepath = "/tmp/%s" % filename
	
	# 保存到 /tmp/
	var err = img.save_png(filepath)
	if err != OK:
		_send_error("failed_to_save_screenshot")
		return
	
	_send_message({
		"type": "screenshot",
		"payload": {
			"width": width,
			"height": height,
			"filepath": filepath
		}
	})
	
	print(" Screenshot saved: ", filepath)

func _send_camera_info():
	if not camera_controller:
		return
	
	_send_message({
		"type": "camera_info",
		"payload": {
			"target": [
				camera_controller.target_position.x,
				camera_controller.target_position.y,
				camera_controller.target_position.z
			],
			"distance": camera_controller.distance,
			"azimuth_deg": rad_to_deg(camera_controller.azimuth),
			"polar_deg": rad_to_deg(camera_controller.polar)
		}
	})

func _send_ack(cmd_type: String):
	_send_message({
		"type": "ack",
		"payload": {"command": cmd_type}
	})

func _send_error(error_msg: String):
	_send_message({
		"type": "error",
		"payload": {"message": error_msg}
	})

func _send_message(data: Dictionary):
	var json_str = JSON.stringify(data)
	print(" Sending: ", json_str)
	
	if _connected and socket:
		var err = socket.send_text(json_str)
		if err != OK:
			print(" Send error: ", err)
		else:
			print(" Sent successfully")
	else:
		print(" Not connected, queuing message")
		_message_queue.append(json_str)

func _flush_message_queue():
	while _message_queue.size() > 0:
		var msg = _message_queue.pop_front()
		socket.send_text(msg)
