package handlers

import (
	"net/http"
	"start-app/models"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserHandler struct {
	DB *gorm.DB
}

func (h *UserHandler) GetAllUsers(c *fiber.Ctx) error {
	var users []models.User
	h.DB.Find(&users)
	return c.Status(http.StatusOK).JSON(users)
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	// Buat struct bantuan untuk menangkap input password
	type SignUpInput struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	input := new(SignUpInput)
	if err := c.BodyParser(input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Format data salah",
		})
	}

	// Validasi: Username dan Email tidak boleh kosong
	if input.Username == "" || input.Email == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Username dan email wajib diisi",
		})
	}

	// Hashing Password
	hashedPasword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal memproses password",
		})
	}

	user := models.User{
		Username: input.Username,
		Email:    input.Email,
		Password: string(hashedPasword),
	}

	if err := h.DB.Create(&user).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal menyimpan user ke database",
		})
	}
	return c.Status(http.StatusCreated).JSON(user)
}

func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var user models.User
	if err := h.DB.First(&user, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "User tidak ditemukan",
		})
	}
	return c.Status(http.StatusOK).JSON(user)
}
