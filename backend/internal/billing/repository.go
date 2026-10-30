// Package billing owns medical catalog, quotes, invoices, and payments.
package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateProcedure(ctx context.Context, p Procedure) (Procedure, error)
	ListProcedures(ctx context.Context) ([]Procedure, error)
	GetProcedure(ctx context.Context, id string) (Procedure, error)
	UpdateProcedure(ctx context.Context, p Procedure) (Procedure, error)

	CreateQuote(ctx context.Context, q Quote) (Quote, error)
	ListQuotes(ctx context.Context, patientID *string) ([]Quote, error)
	GetQuote(ctx context.Context, id string) (Quote, error)
	UpdateQuote(ctx context.Context, q Quote) (Quote, error)
	BeginQuoteApproval(ctx context.Context) (QuoteApprovalTx, error)

	BeginInvoiceTx(ctx context.Context) (InvoiceTx, error)
	ListInvoicesByPatient(ctx context.Context, patientID string) ([]Invoice, error)
	BeginCashCloseTx(ctx context.Context) (CashCloseTx, error)
}

type QuoteApprovalTx interface {
	LoadQuoteForUpdate(ctx context.Context, quoteID string) (Quote, error)
	SaveQuote(ctx context.Context, quote Quote) error
	Exec(ctx context.Context, query string, args ...any) error
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// CashCloseTx keeps the system totals private until the service has received the count request.
type CashCloseTx interface {
	SystemTotals(ctx context.Context, dayStart, nextDayStart time.Time) (CashTotals, error)
	SaveCashClose(ctx context.Context, close CashClose) (CashClose, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type CashTotals struct {
	Cash      int64
	Transfers int64
}

// InvoiceTx is the billing domain's unit of work for atomic invoice and payment changes.
type InvoiceTx interface {
	CreateInvoice(ctx context.Context, invoice Invoice) (Invoice, error)
	LockInvoice(ctx context.Context, invoiceID string) (InvoiceBalance, error)
	CreatePayment(ctx context.Context, invoiceID string, payment Payment) (Payment, error)
	SetInvoiceStatus(ctx context.Context, invoiceID, status string) error
	LoadInvoice(ctx context.Context, invoiceID string) (Invoice, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type InvoiceBalance struct {
	Total  int64
	Paid   int64
	Status string
}

type invoiceQueryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Procedure operations
func (r *PostgresRepository) CreateProcedure(ctx context.Context, p Procedure) (Procedure, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO procedures (code, name, description, category, base_price, active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id::text, created_at, updated_at
	`, p.Code, p.Name, p.Description, p.Category, p.BasePrice, p.Active).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return Procedure{}, fmt.Errorf("insert procedure: %w", err)
	}
	return p, nil
}

func (r *PostgresRepository) ListProcedures(ctx context.Context) ([]Procedure, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id::text, code, name, description, category, base_price, active, created_at, updated_at
		FROM procedures
		ORDER BY category, code
	`)
	if err != nil {
		return nil, fmt.Errorf("list procedures: %w", err)
	}
	defer rows.Close()

	var procedures []Procedure
	for rows.Next() {
		var p Procedure
		if err := rows.Scan(&p.ID, &p.Code, &p.Name, &p.Description, &p.Category, &p.BasePrice, &p.Active, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan procedure: %w", err)
		}
		procedures = append(procedures, p)
	}
	return procedures, nil
}

func (r *PostgresRepository) GetProcedure(ctx context.Context, id string) (Procedure, error) {
	var p Procedure
	err := r.db.QueryRow(ctx, `
		SELECT id::text, code, name, description, category, base_price, active, created_at, updated_at
		FROM procedures
		WHERE id = $1::uuid
	`, id).Scan(&p.ID, &p.Code, &p.Name, &p.Description, &p.Category, &p.BasePrice, &p.Active, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Procedure{}, ErrProcedureNotFound
		}
		return Procedure{}, fmt.Errorf("get procedure: %w", err)
	}
	return p, nil
}

func (r *PostgresRepository) UpdateProcedure(ctx context.Context, p Procedure) (Procedure, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE procedures SET code=$1, name=$2, description=$3, category=$4, base_price=$5, active=$6, updated_at=now()
		WHERE id=$7::uuid
	`, p.Code, p.Name, p.Description, p.Category, p.BasePrice, p.Active, p.ID)
	if err != nil {
		return Procedure{}, fmt.Errorf("update procedure: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return Procedure{}, ErrProcedureNotFound
	}
	return r.GetProcedure(ctx, p.ID)
}

// Quote operations
func (r *PostgresRepository) CreateQuote(ctx context.Context, q Quote) (Quote, error) {
	itemsJSON, _ := json.Marshal(q.Items)
	err := r.db.QueryRow(ctx, `
		INSERT INTO quotes (patient_id, items, subtotal, discount, total, status)
		VALUES ($1::uuid, $2, $3, $4, $5, $6)
		RETURNING id::text, created_at, updated_at
	`, q.PatientID, itemsJSON, q.Subtotal, q.Discount, q.Total, q.Status).Scan(&q.ID, &q.CreatedAt, &q.UpdatedAt)
	if err != nil {
		return Quote{}, fmt.Errorf("insert quote: %w", err)
	}
	return q, nil
}

func (r *PostgresRepository) ListQuotes(ctx context.Context, patientID *string) ([]Quote, error) {
	query := `
		SELECT id::text, patient_id::text, items, subtotal, discount, total, status, approved_at, created_at, updated_at
		FROM quotes
		WHERE 1=1
	`
	args := []any{}
	argIdx := 1

	if patientID != nil {
		query += fmt.Sprintf(" AND patient_id = $%d::uuid", argIdx)
		args = append(args, *patientID)
		argIdx++
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list quotes: %w", err)
	}
	defer rows.Close()

	var quotes []Quote
	for rows.Next() {
		var q Quote
		var itemsJSON []byte
		if err := rows.Scan(&q.ID, &q.PatientID, &itemsJSON, &q.Subtotal, &q.Discount, &q.Total, &q.Status, &q.ApprovedAt, &q.CreatedAt, &q.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan quote: %w", err)
		}
		if err := json.Unmarshal(itemsJSON, &q.Items); err != nil {
			return nil, fmt.Errorf("decode quote items: %w", err)
		}
		quotes = append(quotes, q)
	}
	return quotes, nil
}

func (r *PostgresRepository) GetQuote(ctx context.Context, id string) (Quote, error) {
	var q Quote
	var itemsJSON []byte
	err := r.db.QueryRow(ctx, `
		SELECT id::text, patient_id::text, items, subtotal, discount, total, status, approved_at, created_at, updated_at
		FROM quotes
		WHERE id = $1::uuid
	`, id).Scan(&q.ID, &q.PatientID, &itemsJSON, &q.Subtotal, &q.Discount, &q.Total, &q.Status, &q.ApprovedAt, &q.CreatedAt, &q.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Quote{}, ErrQuoteNotFound
		}
		return Quote{}, fmt.Errorf("get quote: %w", err)
	}
	if err := json.Unmarshal(itemsJSON, &q.Items); err != nil {
		return Quote{}, fmt.Errorf("decode quote items: %w", err)
	}
	return q, nil
}

func (r *PostgresRepository) UpdateQuote(ctx context.Context, q Quote) (Quote, error) {
	itemsJSON, _ := json.Marshal(q.Items)
	tag, err := r.db.Exec(ctx, `
		UPDATE quotes SET patient_id=$1::uuid, items=$2, subtotal=$3, discount=$4, total=$5, status=$6, approved_at=$7, updated_at=now()
		WHERE id=$8::uuid AND status='draft'
	`, q.PatientID, itemsJSON, q.Subtotal, q.Discount, q.Total, q.Status, q.ApprovedAt, q.ID)
	if err != nil {
		return Quote{}, fmt.Errorf("update quote: %w", err)
	}
	if tag.RowsAffected() == 0 {
		if _, err := r.GetQuote(ctx, q.ID); err != nil {
			return Quote{}, err
		}
		return Quote{}, ErrQuoteNotApprovable
	}
	return r.GetQuote(ctx, q.ID)
}

type postgresQuoteApprovalTx struct{ tx pgx.Tx }

func (r *PostgresRepository) BeginQuoteApproval(ctx context.Context) (QuoteApprovalTx, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin quote approval: %w", err)
	}
	return &postgresQuoteApprovalTx{tx: tx}, nil
}

func (r *postgresQuoteApprovalTx) LoadQuoteForUpdate(ctx context.Context, quoteID string) (Quote, error) {
	var quote Quote
	var itemsJSON []byte
	err := r.tx.QueryRow(ctx, `
		SELECT id::text, patient_id::text, items, subtotal, discount, total, status, approved_at, created_at, updated_at
		FROM quotes WHERE id = $1::uuid FOR UPDATE
	`, quoteID).Scan(&quote.ID, &quote.PatientID, &itemsJSON, &quote.Subtotal, &quote.Discount,
		&quote.Total, &quote.Status, &quote.ApprovedAt, &quote.CreatedAt, &quote.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Quote{}, ErrQuoteNotFound
		}
		return Quote{}, fmt.Errorf("lock quote for approval: %w", err)
	}
	if err := json.Unmarshal(itemsJSON, &quote.Items); err != nil {
		return Quote{}, fmt.Errorf("decode quote items for approval: %w", err)
	}
	return quote, nil
}

func (r *postgresQuoteApprovalTx) SaveQuote(ctx context.Context, quote Quote) error {
	itemsJSON, err := json.Marshal(quote.Items)
	if err != nil {
		return fmt.Errorf("encode quote items: %w", err)
	}
	tag, err := r.tx.Exec(ctx, `
		UPDATE quotes SET items=$1, status=$2, approved_at=$3, updated_at=now()
		WHERE id=$4::uuid
	`, itemsJSON, quote.Status, quote.ApprovedAt, quote.ID)
	if err != nil {
		return fmt.Errorf("save approved quote: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrQuoteNotFound
	}
	return nil
}

func (r *postgresQuoteApprovalTx) Exec(ctx context.Context, query string, args ...any) error {
	_, err := r.tx.Exec(ctx, query, args...)
	return err
}

func (r *postgresQuoteApprovalTx) Commit(ctx context.Context) error { return r.tx.Commit(ctx) }

func (r *postgresQuoteApprovalTx) Rollback(ctx context.Context) error { return r.tx.Rollback(ctx) }

type postgresInvoiceTx struct{ tx pgx.Tx }

func (r *PostgresRepository) BeginInvoiceTx(ctx context.Context) (InvoiceTx, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin invoice transaction: %w", err)
	}
	return &postgresInvoiceTx{tx: tx}, nil
}

func (r *postgresInvoiceTx) CreateInvoice(ctx context.Context, invoice Invoice) (Invoice, error) {
	err := r.tx.QueryRow(ctx, `
		INSERT INTO invoices (patient_id, quote_id, subtotal, tax, total, status, issued_at)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, 'issued', $6)
		RETURNING id::text, created_at, updated_at
	`, invoice.PatientID, invoice.QuoteID, invoice.Subtotal, invoice.Tax, invoice.Total, invoice.IssuedAt).
		Scan(&invoice.ID, &invoice.CreatedAt, &invoice.UpdatedAt)
	if err != nil {
		return Invoice{}, fmt.Errorf("insert invoice: %w", err)
	}
	for index := range invoice.Items {
		item := invoice.Items[index]
		if _, err := r.tx.Exec(ctx, `
			INSERT INTO invoice_items (invoice_id, description, unit_price, quantity, subtotal)
			VALUES ($1::uuid, $2, $3, $4, $5)
		`, invoice.ID, item.Description, item.UnitPrice, item.Quantity, item.Subtotal); err != nil {
			return Invoice{}, fmt.Errorf("insert invoice item: %w", err)
		}
	}
	return invoice, nil
}

func (r *postgresInvoiceTx) LockInvoice(ctx context.Context, invoiceID string) (InvoiceBalance, error) {
	var balance InvoiceBalance
	err := r.tx.QueryRow(ctx, `
		SELECT total, status FROM invoices WHERE id = $1::uuid FOR UPDATE
	`, invoiceID).Scan(&balance.Total, &balance.Status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return InvoiceBalance{}, ErrInvoiceNotFound
		}
		return InvoiceBalance{}, fmt.Errorf("lock invoice for payment: %w", err)
	}
	if err := r.tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0) FROM payments WHERE invoice_id = $1::uuid
	`, invoiceID).Scan(&balance.Paid); err != nil {
		return InvoiceBalance{}, fmt.Errorf("sum invoice payments: %w", err)
	}
	return balance, nil
}

func (r *postgresInvoiceTx) CreatePayment(ctx context.Context, invoiceID string, payment Payment) (Payment, error) {
	err := r.tx.QueryRow(ctx, `
		INSERT INTO payments (invoice_id, amount, method, reference)
		VALUES ($1::uuid, $2, $3, $4)
		RETURNING created_at
	`, invoiceID, payment.Amount, payment.Method, payment.Reference).Scan(&payment.CreatedAt)
	if err != nil {
		return Payment{}, fmt.Errorf("insert invoice payment: %w", err)
	}
	return payment, nil
}

func (r *postgresInvoiceTx) SetInvoiceStatus(ctx context.Context, invoiceID, status string) error {
	if _, err := r.tx.Exec(ctx, `UPDATE invoices SET status=$1, updated_at=now() WHERE id=$2::uuid`, status, invoiceID); err != nil {
		return fmt.Errorf("update invoice payment status: %w", err)
	}
	return nil
}

func (r *postgresInvoiceTx) LoadInvoice(ctx context.Context, invoiceID string) (Invoice, error) {
	return loadInvoice(ctx, r.tx, invoiceID)
}

func (r *postgresInvoiceTx) Commit(ctx context.Context) error { return r.tx.Commit(ctx) }

func (r *postgresInvoiceTx) Rollback(ctx context.Context) error { return r.tx.Rollback(ctx) }

type postgresCashCloseTx struct{ tx pgx.Tx }

func (r *PostgresRepository) BeginCashCloseTx(ctx context.Context) (CashCloseTx, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin cash close transaction: %w", err)
	}
	return &postgresCashCloseTx{tx: tx}, nil
}

func (r *postgresCashCloseTx) SystemTotals(ctx context.Context, dayStart, nextDayStart time.Time) (CashTotals, error) {
	var totals CashTotals
	err := r.tx.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN p.method = 'cash' THEN p.amount ELSE 0 END), 0)::bigint,
			COALESCE(SUM(CASE WHEN p.method IN ('card', 'transfer', 'insurance') THEN p.amount ELSE 0 END), 0)::bigint
		FROM invoices AS i
		JOIN payments AS p ON p.invoice_id = i.id
		WHERE i.status = 'paid' AND i.updated_at >= $1 AND i.updated_at < $2
	`, dayStart, nextDayStart).Scan(&totals.Cash, &totals.Transfers)
	if err != nil {
		return CashTotals{}, fmt.Errorf("calculate cash close totals: %w", err)
	}
	return totals, nil
}

func (r *postgresCashCloseTx) SaveCashClose(ctx context.Context, close CashClose) (CashClose, error) {
	err := r.tx.QueryRow(ctx, `
		INSERT INTO cash_register_closures (
			user_id, closure_date, reported_cash, reported_transfers,
			system_cash, system_transfers, status
		) VALUES ($1::uuid, $2::date, $3, $4, $5, $6, $7)
		RETURNING id::text, created_at
	`, close.UserID, close.ClosureDate, close.ReportedCash, close.ReportedTransfers,
		close.SystemCash, close.SystemTransfers, close.Status).Scan(&close.ID, &close.CreatedAt)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return CashClose{}, ErrCashCloseExists
		}
		return CashClose{}, fmt.Errorf("save cash close: %w", err)
	}
	return close, nil
}

func (r *postgresCashCloseTx) Commit(ctx context.Context) error { return r.tx.Commit(ctx) }

func (r *postgresCashCloseTx) Rollback(ctx context.Context) error { return r.tx.Rollback(ctx) }

func (r *PostgresRepository) ListInvoicesByPatient(ctx context.Context, patientID string) ([]Invoice, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id::text FROM invoices WHERE patient_id = $1::uuid ORDER BY created_at DESC, id
	`, patientID)
	if err != nil {
		return nil, fmt.Errorf("list patient invoices: %w", err)
	}
	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan invoice id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read invoice ids: %w", err)
	}
	invoices := make([]Invoice, 0, len(ids))
	for _, id := range ids {
		invoice, err := r.GetInvoice(ctx, id)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}
	return invoices, nil
}

func (r *PostgresRepository) GetInvoice(ctx context.Context, id string) (Invoice, error) {
	return loadInvoice(ctx, r.db, id)
}

func loadInvoice(ctx context.Context, db invoiceQueryer, id string) (Invoice, error) {
	var invoice Invoice
	err := db.QueryRow(ctx, `
		SELECT id::text, patient_id::text, quote_id::text, subtotal, tax, total, status,
		       issued_at, created_at, updated_at
		FROM invoices WHERE id = $1::uuid
	`, id).Scan(&invoice.ID, &invoice.PatientID, &invoice.QuoteID, &invoice.Subtotal,
		&invoice.Tax, &invoice.Total, &invoice.Status, &invoice.IssuedAt, &invoice.CreatedAt, &invoice.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Invoice{}, ErrInvoiceNotFound
		}
		return Invoice{}, fmt.Errorf("get invoice: %w", err)
	}
	rows, err := db.Query(ctx, `
		SELECT description, quantity, unit_price, subtotal
		FROM invoice_items WHERE invoice_id = $1::uuid ORDER BY id
	`, id)
	if err != nil {
		return Invoice{}, fmt.Errorf("list invoice items: %w", err)
	}
	invoice.Items = make([]InvoiceItem, 0)
	for rows.Next() {
		var item InvoiceItem
		if err := rows.Scan(&item.Description, &item.Quantity, &item.UnitPrice, &item.Subtotal); err != nil {
			rows.Close()
			return Invoice{}, fmt.Errorf("scan invoice item: %w", err)
		}
		invoice.Items = append(invoice.Items, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return Invoice{}, fmt.Errorf("read invoice items: %w", err)
	}
	rows.Close()
	paymentRows, err := db.Query(ctx, `
		SELECT amount, method, reference, created_at FROM payments WHERE invoice_id = $1::uuid ORDER BY created_at, id
	`, id)
	if err != nil {
		return Invoice{}, fmt.Errorf("list invoice payments: %w", err)
	}
	defer paymentRows.Close()
	invoice.Payments = make([]Payment, 0)
	for paymentRows.Next() {
		var payment Payment
		if err := paymentRows.Scan(&payment.Amount, &payment.Method, &payment.Reference, &payment.CreatedAt); err != nil {
			return Invoice{}, fmt.Errorf("scan invoice payment: %w", err)
		}
		invoice.Payments = append(invoice.Payments, payment)
	}
	if err := paymentRows.Err(); err != nil {
		return Invoice{}, fmt.Errorf("read invoice payments: %w", err)
	}
	return invoice, nil
}
