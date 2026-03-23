extends Node

# 自动生成模型并退出

const OUTPUT_DIR = "res://assets/models_generated/"

func _ready():
	print("\n=== Agent-Town 模型生成器 ===\n")
	
	# 等待一帧确保所有类已加载
	call_deferred("_do_generate")

func _do_generate():
	_generate_all()
	
	# 等待保存完成
	await get_tree().create_timer(0.5).timeout
	
	print("\n=== 正在退出 ===\n")
	get_tree().quit(0)

func _generate_all():
	print("=== 开始生成建筑模型 ===")
	print("风格：纪念碑谷（Monument Valley）低多边形")
	print("")
	
	# 确保输出目录存在
	var dir = DirAccess.open("res://assets/")
	if not dir:
		push_error("✗ 无法打开 assets 目录")
		return
	
	if not dir.dir_exists("models_generated"):
		dir.make_dir("models_generated")
		print("✓ 创建目录: ", OUTPUT_DIR)
	
	# 生成政府大厅
	_generate_gov_hall()
	
	print("")
	print("=== 生成完成 ===")
	print("输出目录: ", OUTPUT_DIR)

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
	
	var path = OUTPUT_DIR + "gov_hall.tres"
	var err = ResourceSaver.save(mesh, path)
	
	if err == OK:
		print("  ✓ 已保存: ", path)
		print("  顶点数: ", _get_mesh_vertex_count(mesh))
	else:
		push_error("  ✗ 保存失败: " + path + " (错误码: " + str(err) + ")")

func _get_mesh_vertex_count(mesh: ArrayMesh) -> int:
	if mesh == null or mesh.get_surface_count() == 0:
		return 0
	var arrays = mesh.surface_get_arrays(0)
	if arrays.size() > Mesh.ARRAY_VERTEX:
		return arrays[Mesh.ARRAY_VERTEX].size()
	return 0
