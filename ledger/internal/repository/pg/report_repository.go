package pg

import (
	"context"
	"time"

	"ledger/internal/db/sqlc"
	"ledger/internal/domain"

	"github.com/google/uuid"
)

type ReportRepo struct {
	q *sqlc.Queries
}

func NewReportRepo(q *sqlc.Queries) *ReportRepo {
	return &ReportRepo{q: q}
}

func (r *ReportRepo) GetReportSummary(
	ctx context.Context,
	userID uuid.UUID,
	from time.Time,
	to time.Time,
) ([]domain.ReportSummary, error) {

	rows, err := r.q.ReportSummary(ctx, sqlc.ReportSummaryParams{
		UserID:   userID,
		FromDate: from,
		ToDate:   to,
	})
	if err != nil {
		return nil, err
	}

	res := make([]domain.ReportSummary, 0, len(rows))
	for _, row := range rows {
		res = append(res, domain.ReportSummary{
			Category: row.Category,
			Total:    row.Total,
		})
	}
	return res, nil
}
