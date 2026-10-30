package budget

import (
	"github.com/gofiber/fiber/v2"
)

func (h *handler) ApproveItems(c *fiber.Ctx) error {
	var req ApproveBudgetRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Payload inválido"})
	}

	if err := h.service.ApproveItems(c.Context(), c.Params("id"), req); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Aprobado"})
}
