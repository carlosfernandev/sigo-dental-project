func (s *service) ApproveItems(ctx context.Context, budgetID string, req ApproveBudgetRequest) error {
	id, _ := uuid.Parse(budgetID)
	// (Validaciones omitidas por brevedad)
	budget, _ := s.repo.GetByID(ctx, id)
	return s.repo.ApproveBudgetItems(ctx, id, req.ItemIDs, budget.PatientID)
}