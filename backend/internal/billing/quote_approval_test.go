package billing

import (
	"context"
	"errors"
	"testing"

	"github.com/carlosfpatino/sigo-dental/internal/scheduling"
	"github.com/carlosfpatino/sigo-dental/platform/authn"
)

type quoteApprovalRepositoryFake struct {
	Repository
	quote Quote
	tx    *quoteApprovalTxFake
}

func (r *quoteApprovalRepositoryFake) BeginQuoteApproval(context.Context) (QuoteApprovalTx, error) {
	r.tx = &quoteApprovalTxFake{repo: r}
	return r.tx, nil
}

type quoteApprovalTxFake struct {
	repo       *quoteApprovalRepositoryFake
	committed  bool
	rolledBack bool
}

func (tx *quoteApprovalTxFake) LoadQuoteForUpdate(context.Context, string) (Quote, error) {
	return tx.repo.quote, nil
}

func (tx *quoteApprovalTxFake) SaveQuote(_ context.Context, quote Quote) error {
	tx.repo.quote = quote
	return nil
}

func (*quoteApprovalTxFake) Exec(context.Context, string, ...any) error { return nil }
func (tx *quoteApprovalTxFake) Commit(context.Context) error {
	tx.committed = true
	return nil
}
func (tx *quoteApprovalTxFake) Rollback(context.Context) error {
	tx.rolledBack = true
	return nil
}

type schedulableTaskCreatorFake struct {
	tasks []scheduling.SchedulableTask
	err   error
}

func (c *schedulableTaskCreatorFake) CreateSchedulableTasks(_ context.Context, _ scheduling.Executor, tasks []scheduling.SchedulableTask) error {
	c.tasks = append(c.tasks, tasks...)
	return c.err
}

func TestApproveQuotePartialThenTotal(t *testing.T) {
	quoteID := "123e4567-e89b-42d3-a456-426614174000"
	patientID := "123e4567-e89b-42d3-a456-426614174001"
	quote := Quote{
		ID: quoteID, PatientID: patientID, Status: "draft",
		Items: []QuoteItem{
			{ID: "item-1", ProcedureName: "Cleaning", Status: "pending"},
			{ID: "item-2", ProcedureName: "Filling", Status: "pending"},
		},
	}
	repository := &quoteApprovalRepositoryFake{quote: quote}
	taskCreator := &schedulableTaskCreatorFake{}
	service := NewService(repository, taskCreator)
	principal := authn.Principal{UserID: patientID, Role: authn.RoleReceptionistCashier}

	partial, err := service.ApproveQuote(context.Background(), principal, quoteID, ApproveQuoteRequest{ApprovedItemIDs: []string{"item-1"}})
	if err != nil {
		t.Fatalf("partial approval error = %v", err)
	}
	if partial.Status != "partially_approved" || partial.Items[0].Status != "approved" || partial.Items[1].Status != "pending" {
		t.Fatalf("partial approval result = status %q, items %+v", partial.Status, partial.Items)
	}
	if len(taskCreator.tasks) != 1 || taskCreator.tasks[0].ItemID != "item-1" || !repository.tx.committed {
		t.Fatalf("partial approval tasks = %+v, committed = %v", taskCreator.tasks, repository.tx.committed)
	}

	complete, err := service.ApproveQuote(context.Background(), principal, quoteID, ApproveQuoteRequest{ApprovedItemIDs: []string{"item-2"}})
	if err != nil {
		t.Fatalf("final approval error = %v", err)
	}
	if complete.Status != "approved" || complete.Items[0].Status != "approved" || complete.Items[1].Status != "approved" {
		t.Fatalf("final approval result = status %q, items %+v", complete.Status, complete.Items)
	}
	if len(taskCreator.tasks) != 2 || taskCreator.tasks[1].ItemID != "item-2" || !repository.tx.committed {
		t.Fatalf("final approval tasks = %+v, committed = %v", taskCreator.tasks, repository.tx.committed)
	}
}

func TestApproveQuoteRollsBackWhenTaskCreationFails(t *testing.T) {
	quoteID := "123e4567-e89b-42d3-a456-426614174000"
	patientID := "123e4567-e89b-42d3-a456-426614174001"
	repository := &quoteApprovalRepositoryFake{quote: Quote{
		ID: quoteID, PatientID: patientID, Status: "draft",
		Items: []QuoteItem{{ID: "item-1", ProcedureName: "Cleaning", Status: "pending"}},
	}}
	service := NewService(repository, &schedulableTaskCreatorFake{err: errors.New("insert failed")})
	principal := authn.Principal{UserID: patientID, Role: authn.RoleReceptionistCashier}

	if _, err := service.ApproveQuote(context.Background(), principal, quoteID, ApproveQuoteRequest{ApprovedItemIDs: []string{"item-1"}}); err == nil {
		t.Fatal("ApproveQuote() error = nil, want task creation error")
	}
	if repository.tx.committed || !repository.tx.rolledBack {
		t.Fatalf("transaction state after task creation failed: committed=%v rolled_back=%v", repository.tx.committed, repository.tx.rolledBack)
	}
}
