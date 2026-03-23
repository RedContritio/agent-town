@tool
extends Node

# 建筑生成器按钮 - 附加到场景中的任意节点
# 在编辑器中点击"生成"按钮即可运行

@export var generate_button: bool = false :
	set(value):
		if value:
			_generate()
			generate_button = false  # 重置按钮

func _generate():
	print("=== 开始生成建筑模型 ===")
	print("风格：纪念碑谷（Monument Valley）低多边形")
	
	var OUTPUT_DIR = "res://assets/models_generated/"
	
	# 确保输出目录存在
	var dir = DirAccess.open("res://assets/")
	if not dir.dir_exists("models_generated"):
		dir.make_dir("models_generated")
		print("✓ 创建目录: ", OUTPUT_DIR)
	
	# 生成政府大厅
	_generate_gov_hall()
	
	print("")
	print("=== 生成完成 ===")
	print("输出目录: ", OUTPUT_DIR)
	
	# 刷新文件系统
	EditorInterface.get_resource_filesystem().scan()

func _generate_gov_hall() -> void:
	print("生成: gov_hall (政府大厅)")
	
	var generator = GovHallGenerator.new()
	generator.set_size(3.0, 3.0, 2.0)
	
	var colors = BuildingGenerator.get_default_colors("gov_hall")
	generator.set_colors(colors.primary, colors.secondary, colors.accent)
	
	var mesh = generator.generate()
	
	if mesh == null:
		push_error("✗ 生成失败: gov_hall")
		return
	
	var path = "res://assets/models_generated/gov_hall.tres"
	var err = ResourceSaver.save(mesh, path)
	
	if err == OK:
		print("  ✓ 已保存: ", path)
	else:
		push_error("  ✗ 保存失败: " + path)
