func (r *repository) ApproveBudgetItems(ctx context.Context, budgetID uuid.UUID, itemIDs []uuid.UUID, patientID uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, itemID := range itemIDs {
		// Actualizar item
		_, err := tx.ExecContext(ctx, "UPDATE budget_items SET status = $1 WHERE id = $2 AND budget_id = $3", StatusApproved, itemID, budgetID)
		if err != nil {
			return err
		}

		// Crear tarea agendable
		_, err = tx.ExecContext(ctx, "INSERT INTO schedulable_tasks (patient_id, budget_item_id, status) VALUES ($1, $2, 'pending_scheduling')", patientID, itemID)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}