package repository

import (
	"database/sql"
	"errors"

	"github.com/RedContritio/agent-town/server/internal/model"
)

// TaskRepository 任务数据访问
type TaskRepository struct {
	DB *sql.DB
}

// NewTaskRepository 创建 TaskRepository
func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{DB: db}
}

// GetNextSeq 获取下一个序列号
func (r *TaskRepository) GetNextSeq(agentID int64) (int, error) {
	var maxSeq sql.NullInt32
	err := r.DB.QueryRow(
		"SELECT MAX(seq) FROM tasks WHERE agent_id=?",
		agentID,
	).Scan(&maxSeq)
	if err != nil {
		return 0, err
	}
	if maxSeq.Valid {
		return int(maxSeq.Int32) + 1, nil
	}
	return 1, nil
}

// Create 创建任务（需要在外层处理事务）
func (r *TaskRepository) Create(tx *sql.Tx, task *model.Task) error {
	_, err := tx.Exec(
		`INSERT INTO tasks (agent_id, seq, type, status, params, stack_depth, created_at, started_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		task.AgentID, task.Seq, task.Type, task.Status, task.Params,
		task.StackDepth, task.CreatedAt, task.StartedAt,
	)
	return err
}

// GetByAgentAndSeq 根据 AgentID 和 Seq 获取任务
func (r *TaskRepository) GetByAgentAndSeq(agentID int64, seq int) (*model.Task, error) {
	task := &model.Task{}
	var startedAt sql.NullInt64
	var completedAt sql.NullInt64
	var errorCode sql.NullInt32
	var result sql.NullString

	err := r.DB.QueryRow(
		`SELECT agent_id, seq, type, status, params, stack_depth, result, error_code,
			created_at, started_at, completed_at
		 FROM tasks WHERE agent_id=? AND seq=?`,
		agentID, seq,
	).Scan(
		&task.AgentID, &task.Seq, &task.Type, &task.Status, &task.Params,
		&task.StackDepth, &result, &errorCode,
		&task.CreatedAt, &startedAt, &completedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if result.Valid {
		task.Result = result.String
	}
	if startedAt.Valid {
		task.StartedAt = &startedAt.Int64
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Int64
	}
	if errorCode.Valid {
		ec := int(errorCode.Int32)
		task.ErrorCode = &ec
	}

	return task, nil
}

// GetStack 获取任务栈（按 stack_depth 排序）
func (r *TaskRepository) GetStack(agentID int64) ([]*model.Task, error) {
	rows, err := r.DB.Query(
		`SELECT agent_id, seq, type, status, params, stack_depth, result, error_code,
			created_at, started_at, completed_at
		 FROM tasks
		 WHERE agent_id=? AND status IN (?, ?, ?)
		 ORDER BY stack_depth ASC`,
		agentID, model.TaskStatusPending, model.TaskStatusRunning, model.TaskStatusPaused,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTasks(rows)
}

// GetByAgentID 获取 Agent 的所有任务
func (r *TaskRepository) GetByAgentID(agentID int64) ([]*model.Task, error) {
	rows, err := r.DB.Query(
		`SELECT agent_id, seq, type, status, params, stack_depth, result, error_code,
			created_at, started_at, completed_at
		 FROM tasks WHERE agent_id=? ORDER BY seq DESC`,
		agentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTasks(rows)
}

// UpdateStatus 更新任务状态
func (r *TaskRepository) UpdateStatus(tx *sql.Tx, agentID int64, seq, status int) error {
	var startedAt interface{}
	if status == model.TaskStatusRunning {
		now := model.NowMillis()
		startedAt = now
	}

	_, err := tx.Exec(
		"UPDATE tasks SET status=?, started_at=? WHERE agent_id=? AND seq=?",
		status, startedAt, agentID, seq,
	)
	return err
}

// UpdateStackDepth 更新栈深度
func (r *TaskRepository) UpdateStackDepth(tx *sql.Tx, agentID int64, seq, depth int) error {
	_, err := tx.Exec(
		"UPDATE tasks SET stack_depth=? WHERE agent_id=? AND seq=?",
		depth, agentID, seq,
	)
	return err
}

// CompleteTask 完成任务
func (r *TaskRepository) CompleteTask(tx *sql.Tx, agentID int64, seq int, result string) error {
	now := model.NowMillis()
	_, err := tx.Exec(
		"UPDATE tasks SET status=?, result=?, completed_at=? WHERE agent_id=? AND seq=?",
		model.TaskStatusCompleted, result, now, agentID, seq,
	)
	return err
}

// CompleteTaskWithResult 完成任务（独立事务版本）
func (r *TaskRepository) CompleteTaskWithResult(agentID int64, seq int, result string) error {
	now := model.NowMillis()
	_, err := r.DB.Exec(
		"UPDATE tasks SET status=?, result=?, completed_at=? WHERE agent_id=? AND seq=?",
		model.TaskStatusCompleted, result, now, agentID, seq,
	)
	return err
}

// FailTaskWithReason 任务失败
func (r *TaskRepository) FailTaskWithReason(agentID int64, seq, errorCode int, reason string) error {
	now := model.NowMillis()
	_, err := r.DB.Exec(
		"UPDATE tasks SET status=?, error_code=?, completed_at=? WHERE agent_id=? AND seq=?",
		model.TaskStatusFailed, errorCode, now, agentID, seq,
	)
	return err
}

// FailTask 标记任务失败
func (r *TaskRepository) FailTask(tx *sql.Tx, agentID int64, seq, errorCode int) error {
	now := model.NowMillis()
	_, err := tx.Exec(
		"UPDATE tasks SET status=?, error_code=?, completed_at=? WHERE agent_id=? AND seq=?",
		model.TaskStatusFailed, errorCode, now, agentID, seq,
	)
	return err
}

// Delete 删除任务
func (r *TaskRepository) Delete(tx *sql.Tx, agentID int64, seq int) error {
	_, err := tx.Exec(
		"DELETE FROM tasks WHERE agent_id=? AND seq=?",
		agentID, seq,
	)
	return err
}

// GetTopTask 获取栈顶任务（depth=0 且活跃）
func (r *TaskRepository) GetTopTask(agentID int64) (*model.Task, error) {
	task := &model.Task{}
	var startedAt sql.NullInt64
	var completedAt sql.NullInt64
	var errorCode sql.NullInt32
	var result sql.NullString

	err := r.DB.QueryRow(
		`SELECT agent_id, seq, type, status, params, stack_depth, result, error_code,
			created_at, started_at, completed_at
		 FROM tasks
		 WHERE agent_id=? AND stack_depth=0 AND status IN (?, ?, ?)
		 LIMIT 1`,
		agentID, model.TaskStatusPending, model.TaskStatusRunning, model.TaskStatusPaused,
	).Scan(
		&task.AgentID, &task.Seq, &task.Type, &task.Status, &task.Params,
		&task.StackDepth, &result, &errorCode,
		&task.CreatedAt, &startedAt, &completedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if result.Valid {
		task.Result = result.String
	}
	if startedAt.Valid {
		task.StartedAt = &startedAt.Int64
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Int64
	}
	if errorCode.Valid {
		ec := int(errorCode.Int32)
		task.ErrorCode = &ec
	}

	return task, nil
}

// IncrementStackDepth 增加所有活跃任务的栈深度
func (r *TaskRepository) IncrementStackDepth(tx *sql.Tx, agentID int64) error {
	_, err := tx.Exec(
		`UPDATE tasks SET stack_depth = stack_depth + 1
		 WHERE agent_id=? AND status IN (?, ?, ?)`,
		agentID, model.TaskStatusPending, model.TaskStatusRunning, model.TaskStatusPaused,
	)
	return err
}

// scanTasks 扫描任务行
func (r *TaskRepository) scanTasks(rows *sql.Rows) ([]*model.Task, error) {
	var tasks []*model.Task
	for rows.Next() {
		task := &model.Task{}
		var startedAt sql.NullInt64
		var completedAt sql.NullInt64
		var errorCode sql.NullInt32
		var result sql.NullString

		err := rows.Scan(
			&task.AgentID, &task.Seq, &task.Type, &task.Status, &task.Params,
			&task.StackDepth, &result, &errorCode,
			&task.CreatedAt, &startedAt, &completedAt,
		)
		if err != nil {
			return nil, err
		}

		if result.Valid {
			task.Result = result.String
		}
		if startedAt.Valid {
			task.StartedAt = &startedAt.Int64
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Int64
		}
		if errorCode.Valid {
			ec := int(errorCode.Int32)
			task.ErrorCode = &ec
		}

		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}
