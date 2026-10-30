package budget

import (
	"time"
	"uuid"
)

// Añadir a las entidades existentes
const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
)

// Añadir esta nueva entidad
type SchedulableTask struct {
	ID           uuid.UUID `json:"id" db:"id"`
	PatientID    uuid.UUID `json:"patient_id" db:"patient_id"`
	BudgetItemID uuid.UUID `json:"budget_item_id" db:"budget_item_id"`
	Status       string    `json:"status" db:"status"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
