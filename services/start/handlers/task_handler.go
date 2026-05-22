package handlers

import (
	"net/http"
	"start-app/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type TaskHandler struct {
	DB *gorm.DB
}

func (h *TaskHandler) CreateTask(c *fiber.Ctx) error {
	task := new(models.Task)
	if err := c.BodyParser(task); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Format data salah",
		})
	}

	// Cek apakah resource tersedia
	var resource models.Resource
	if err := h.DB.First(&resource, task.ResourceID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Resource tidak ditemukan",
		})
	}

	if resource.Status != "available" {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Resource sedang digunakan atau tidak tersedia",
		})
	}

	// Gunakan Transaction untuk memastikan konsistensi data
	// Jika simpan Task gagal, update status resource juga dibatalkan
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		// Simpan Task
		if err := tx.Create(&task).Error; err != nil {
			return err
		}

		// Update Status Resource menjadi "in_use"
		if err := tx.Model(&resource).Update("status", "in_use").Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal membuat task",
		})
	}

	return c.Status(http.StatusCreated).JSON(task)
}

func (h *TaskHandler) GetAllTasks(c *fiber.Ctx) error {
	var tasks []models.Task
	// Gunakan Preload untuk mengambil data User dan Resource sekaligus
	h.DB.Preload("User").Preload("Resource").Find(&tasks)
	return c.Status(http.StatusOK).JSON(&tasks)
}

func (h *TaskHandler) CompleteTask(c *fiber.Ctx) error {
	id := c.Params("id")
	var task models.Task

	// Cari Task yang dimaksud
	if err := h.DB.First(&task, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Task tidak ditemukan",
		})
	}

	// Gunakan Transaction untuk konsistensi data
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		// Update status task
		if err := tx.Delete(&task).Error; err != nil {
			return err
		}

		// Update Resource kembali ke 'available'
		if err := tx.Model(&models.Resource{}).Where("id = ?", task.ResourceID).Update("status", "available").Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal menyelesaikan task",
		})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Aset telah dikembalikan",
	})
}

func (h *TaskHandler) GetTaskHistory(c *fiber.Ctx) error {
	var tasks []models.Task

	// deleted_at IS NOT NULL memastikan kita hanya mengambil riwayat yang sudah selesai
	err := h.DB.Unscoped().
		Preload("User").
		Preload("Resource").
		Where("deleted_at IS NOT NULL").
		Find(&tasks).Error

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal menarik riwayat pengambalian",
		})
	}

	return c.Status(http.StatusOK).JSON(tasks)
}
