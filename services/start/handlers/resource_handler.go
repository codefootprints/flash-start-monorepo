package handlers

import (
	"net/http"
	"start-app/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ResourceHandler struct {
	DB *gorm.DB
}

func (h *ResourceHandler) GetAllResources(c *fiber.Ctx) error {
	var resources []models.Resource
	h.DB.Find(&resources)
	return c.Status(http.StatusOK).JSON(resources)
}

func (h *ResourceHandler) CreateResource(c *fiber.Ctx) error {
	resource := new(models.Resource)

	if err := c.BodyParser(resource); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Format data salah",
		})
	}
	// Validasi: Nama dan Kategori tidak boleh kosong
	if resource.Name == "" || resource.Category == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Nama aset dan kategori wajib diisi",
		})
	}

	// Set default status
	resource.Status = "available"

	if err := h.DB.Create(&resource).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal menyimpan aset ke database",
		})
	}
	return c.Status(http.StatusOK).JSON(resource)
}

func (h *ResourceHandler) GetResourceByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var resource models.Resource
	if err := h.DB.First(&resource, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Resource tidak ditemukan",
		})
	}
	return c.Status(http.StatusOK).JSON(resource)
}

func (h *ResourceHandler) DeleteResource(c *fiber.Ctx) error {
	id := c.Params("id")
	var resource models.Resource

	// Cari resource-nya dulu
	if err := h.DB.First(&resource, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Resource tidak ditemukan",
		})
	}

	// Cek jika status 'in_use', dilarang hapus
	if resource.Status == "in_use" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Tidak dapat menghapus aset yang sedang dipinjam. Selesaikan peminjaman terlebih dahulu.",
		})
	}

	if err := h.DB.Delete(&models.Resource{}, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Gagal menghapus resource",
		})
	}
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Resource berhasil dihapus",
	})
}
