package tasks

import (
	"context"
	"fmt"
	"product-service/internal/db"
	"product-service/internal/models"
	"sync"
	"time"

	"github.com/google/uuid"
)

func AutomateCreateBlog() {
	fmt.Println("Executing AutomateCreateBlog at", time.Now())
}

var taskManager *TaskManager

type TaskManager struct {
	tasks map[uuid.UUID]context.CancelFunc
	mu    sync.Mutex
	wg    sync.WaitGroup
}

func InitTaskManager() {
	taskManager = &TaskManager{
		tasks: make(map[uuid.UUID]context.CancelFunc),
	}
}

func RunTask(task *models.Task, job func(context.Context) error) {
	if task == nil {
		return
	}

	// Thêm task vào task manager
	taskManager.mu.Lock()
	ctx, cancel := context.WithCancel(context.Background())
	taskManager.tasks[task.ID] = cancel
	taskManager.mu.Unlock()

	// Chạy task
	go func() {
		fmt.Printf("Task %s started", task.ID)

		// Thêm vào wait group và tạo ticker theo phút
		taskManager.wg.Add(1)
		ticker := time.NewTicker(time.Duration(task.Frequency) * time.Minute)

		for {
			select {
			// Nếu nhận được lệnh dừng, thì chuyển trạng thái về stop
			case <-ctx.Done():
				defer taskManager.wg.Done()
				ticker.Stop()
				fmt.Printf("Task %s stopped", task.ID)
				StopTask(task.ID)
				return
			case <-ticker.C:
				// Nếu task hết thời gian thực thi thì dừng
				if task.Mode == models.TASK_MODE_MINUTES && task.StoppedAt != nil {
					if time.Now().After(*task.StoppedAt) {
						defer taskManager.wg.Done()
						StopTask(task.ID)
						ticker.Stop()
						fmt.Printf("Task %s stopped", task.ID)
						StopTask(task.ID)
						return
					}
				}
				// nếu hàm lỗi thì dừng task
				if err := job(ctx); err != nil {
					defer taskManager.wg.Done()
					fmt.Printf("Job error: %v", err)
					ticker.Stop()
					fmt.Printf("Task %s stopped", task.ID)
					StopTask(task.ID)
					return
				}
			}
		}
	}()
}

func StopTask(taskID uuid.UUID) bool {
	// lock hash table khi dừng
	taskManager.mu.Lock()
	defer taskManager.mu.Unlock()

	if cancel, exists := taskManager.tasks[taskID]; exists {
		cancel()
		var task models.Task
		if err := db.DB.First(&task, "id = ?", taskID).Error; err == nil {
			task.Status = models.TASK_STATUS_STOPPED
			if err := db.DB.Save(&task).Error; err != nil {
				fmt.Printf("Error updating task status: %v", err)
			}
		}
		delete(taskManager.tasks, taskID)
		return true
	}
	return false
}

func StopAllTask() {
	fmt.Println("Stop all task ...")
	taskManager.mu.Lock()
	for id, cancel := range taskManager.tasks {
		cancel()
		var task models.Task
		if err := db.DB.First(&task, "id = ?", id).Error; err == nil {
			task.Status = models.TASK_STATUS_STOPPED
			if err := db.DB.Save(&task).Error; err != nil {
				fmt.Printf("Error updating task status: %v", err)
			}
		}
		delete(taskManager.tasks, id)
	}
	taskManager.mu.Unlock()
	// Đợi cho các task tắt hết
	taskManager.wg.Wait()
}
