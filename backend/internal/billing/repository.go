// Package billing owns medical catalog, quotes, invoices, and payments.
package billing

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
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
		WHERE id=$8::uuid
	`, q.PatientID, itemsJSON, q.Subtotal, q.Discount, q.Total, q.Status, q.ApprovedAt, q.ID)
	if err != nil {
		return Quote{}, fmt.Errorf("update quote: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return Quote{}, ErrQuoteNotFound
	}
	return r.GetQuote(ctx, q.ID)
}
