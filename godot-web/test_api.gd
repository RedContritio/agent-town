# 简单的 API 测试脚本
# 可用于在 Godot 编辑器中运行测试

extends SceneTree

func _init():
	print("=== Godot Web API Test ===")
	
	# 测试任务类型转换
	var api_client_script = load("res://scripts/api_client.gd")
	if api_client_script:
		print("✓ ApiClient script loaded")
	else:
		print("✗ Failed to load ApiClient script")
	
	# 测试 AgentPanel 脚本
	var agent_panel_script = load("res://scripts/ui/agent_panel.gd")
	if agent_panel_script:
		print("✓ AgentPanel script loaded")
	else:
		print("✗ Failed to load AgentPanel script")
	
	# 测试 LoginPanel 脚本
	var login_panel_script = load("res://scripts/ui/login_panel.gd")
	if login_panel_script:
		print("✓ LoginPanel script loaded")
	else:
		print("✗ Failed to load LoginPanel script")
	
	# 测试场景文件
	var main_scene = load("res://scenes/main.tscn")
	if main_scene:
		print("✓ Main scene loaded")
	else:
		print("✗ Failed to load Main scene")
	
	var agent_panel_scene = load("res://scenes/ui/agent_panel.tscn")
	if agent_panel_scene:
		print("✓ AgentPanel scene loaded")
	else:
		print("✗ Failed to load AgentPanel scene")
	
	var login_panel_scene = load("res://scenes/ui/login_panel.tscn")
	if login_panel_scene:
		print("✓ LoginPanel scene loaded")
	else:
		print("✗ Failed to load LoginPanel scene")
	
	print("\n=== Test Complete ===")
	quit()
