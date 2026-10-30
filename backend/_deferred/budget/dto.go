package budget

import "github.com/google/uuid"

type ApproveBudgetRequest struct {
	ItemIDs []uuid.UUID `json:"item_ids"`
}
