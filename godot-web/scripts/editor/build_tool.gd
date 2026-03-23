@tool
extends EditorPlugin

# 建筑生成工具 - 添加到编辑器工具栏

const TOOL_NAME = "Generate Buildings"

func _enter_tree():
	add_tool_menu_item(TOOL_NAME, _on_generate_clicked)

func _exit_tree():
	remove_tool_menu_item(TOOL_NAME)

func _on_generate_clicked():
	print("=== 开始生成建筑模型 ===")
	
	var script = load("res://scripts/editor/generate_buildings.gd")
	var instance = script.new()
	instance._run()
