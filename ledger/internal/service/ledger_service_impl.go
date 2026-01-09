package service

import (
	"context"
	"encoding/json"
	"fmt"
	"ledger/internal/cache"
	"ledger/internal/domain"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc/metadata"
)

type ledgerServiceImpl struct {
	cache    cache.Cache
	budgets  domain.BudgetRepository
	expenses domain.ExpenseRepository
	reports  domain.ReportRepository
}

type PeriodRange struct {
	From time.Time
	To   time.Time
}

func (l *ledgerServiceImpl) GetReportSummary(
	ctx context.Context,
	from time.Time,
	to time.Time,
) ([]domain.ReportSummary, error) {

	userID, err := UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf(
		"report:summary:%s:%s:%s",
		userID,
		from.Format("2006-01-02"),
		to.Format("2006-01-02"),
	)

	if data, ok := l.cache.Get(ctx, cacheKey); ok {
		var cached []domain.ReportSummary
		if err := json.Unmarshal(data, &cached); err == nil {
			log.Println("CACHE HIT:", cacheKey)
			return cached, nil
		}
	}

	log.Println("CACHE MISS:", cacheKey)

	categories, err := l.budgets.List(ctx, userID)
	if err != nil {
		return nil, err
	}

	resCh := make(chan domain.ReportSummary, len(categories))
	var wg sync.WaitGroup

	for _, b := range categories {
		category := b.Category
		wg.Add(1)

		go func() {
			defer wg.Done()

			total, err := l.expenses.SumByCategoryAndPeriod(
				ctx,
				userID,
				category,
				from,
				to,
			)
			if err != nil {
				return
			}

			resCh <- domain.ReportSummary{
				Category: category,
				Total:    total,
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resCh)
	}()

	var result []domain.ReportSummary
	for r := range resCh {
		result = append(result, r)
	}

	if data, err := json.Marshal(result); err == nil {
		l.cache.Set(ctx, cacheKey, data, 15*time.Second)
	}

	return result, nil
}
func (l *ledgerServiceImpl) AddTransaction(
	ctx context.Context,
	t domain.Transaction,
) error {

	userID, err := UserIDFromContext(ctx)
	if err != nil {
		return err
	}

	t.UserID = userID

	if err := domain.CheckValid(t); err != nil {
		return err
	}

	budget, err := l.budgets.GetByCategory(ctx, userID, t.Category)
	if err != nil {
		return err
	}
	if budget == nil {
		return domain.ErrBudgetNotFound
	}

	limit := budget.Limit

	pr, err := BudgetPeriodRange(budget.Period, t.Date)
	if err != nil {
		return err
	}

	var spent decimal.Decimal

	if pr == nil {
		// бессрочный бюджет
		spent, err = l.expenses.SumByCategory(ctx, userID, t.Category)
	} else {
		spent, err = l.expenses.SumByCategoryAndPeriod(
			ctx,
			userID,
			t.Category,
			pr.From,
			pr.To,
		)
	}

	if spent.Add(t.Amount).GreaterThan(limit) {
		return &domain.BudgetExceededError{
			Category: t.Category,
			Limit:    limit,
			Current:  spent,
			Amount:   t.Amount,
		}
	}

	if err := l.expenses.Add(ctx, userID, t); err != nil {
		return err
	}

	l.invalidateReportCache(ctx, userID)
	return nil
}

func BudgetPeriodRange(period string, now time.Time) (*PeriodRange, error) {
	switch period {

	case "":
		return nil, nil

	case "daily":
		from := time.Date(
			now.Year(), now.Month(), now.Day(),
			0, 0, 0, 0,
			now.Location(),
		)
		to := from.AddDate(0, 0, 1)
		return &PeriodRange{From: from, To: to}, nil

	case "weekly":
		weekday := int(now.Weekday())
		if weekday == 0 { // Sunday
			weekday = 7
		}

		from := time.Date(
			now.Year(), now.Month(), now.Day(),
			0, 0, 0, 0,
			now.Location(),
		).AddDate(0, 0, -(weekday - 1))

		to := from.AddDate(0, 0, 7)
		return &PeriodRange{From: from, To: to}, nil

	case "monthly":
		from := time.Date(
			now.Year(), now.Month(), 1,
			0, 0, 0, 0,
			now.Location(),
		)
		to := from.AddDate(0, 1, 0)
		return &PeriodRange{From: from, To: to}, nil

	default:
		return nil, fmt.Errorf("unknown budget period: %s", period)
	}
}

func (l *ledgerServiceImpl) ListTransactions(
	ctx context.Context,
) ([]domain.Transaction, error) {

	userID, err := UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return l.expenses.List(ctx, userID)
}

func (l *ledgerServiceImpl) SetBudget(
	ctx context.Context,
	b domain.Budget,
) error {

	userID, err := UserIDFromContext(ctx)
	if err != nil {
		return err
	}
	b.UserID = userID
	if err := domain.CheckValid(b); err != nil {
		return err
	}

	if err := l.budgets.Upsert(ctx, userID, b); err != nil {
		fmt.Println(err)
		return err
	}

	l.invalidateReportCache(ctx, userID)
	l.invalidateBudgetsCache(ctx, userID)

	return nil
}

func (l *ledgerServiceImpl) ListBudgets(
	ctx context.Context,
) ([]domain.Budget, error) {

	userID, err := UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("budgets:all:%s", userID)

	if data, ok := l.cache.Get(ctx, cacheKey); ok {
		var cached []domain.Budget
		if err := json.Unmarshal(data, &cached); err == nil {
			log.Println("CACHE HIT:", cacheKey)
			return cached, nil
		}
	}

	log.Println("CACHE MISS:", cacheKey)

	res, err := l.budgets.List(ctx, userID)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(res); err == nil {
		l.cache.Set(ctx, cacheKey, data, 15*time.Second)
		log.Println("CACHE SET:", cacheKey)
	}

	return res, nil
}

func (l *ledgerServiceImpl) BulkAddTransactions(
	ctx context.Context,
	txs []domain.Transaction,
	workers int,
) (*domain.BulkImportResult, error) {

	userID, err := UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	type job struct {
		index int
		tx    domain.Transaction
	}

	jobs := make(chan job)
	//results := make(chan error)

	var accepted, rejected int64
	var errorsList []domain.BulkImportError
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				j.tx.UserID = userID
				err := l.AddTransaction(ctx, j.tx)
				if err != nil {
					atomic.AddInt64(&rejected, 1)
					mu.Lock()
					errorsList = append(errorsList, domain.BulkImportError{
						Index: j.index,
						Error: err.Error(),
					})
					mu.Unlock()
				} else {
					atomic.AddInt64(&accepted, 1)
				}
			}
		}()
	}

	go func() {
		for i, tx := range txs {
			jobs <- job{i, tx}
		}
		close(jobs)
	}()

	wg.Wait()

	l.invalidateReportCache(ctx, userID)

	return &domain.BulkImportResult{
		Accepted: accepted,
		Rejected: rejected,
		Errors:   errorsList,
	}, nil
}

func New(
	c cache.Cache,
	b domain.BudgetRepository,
	e domain.ExpenseRepository,
	r domain.ReportRepository,
) LedgerService {
	return &ledgerServiceImpl{
		cache:    c,
		budgets:  b,
		expenses: e,
		reports:  r,
	}
}

func startReportHeartbeat(ctx context.Context) func() {
	ticker := time.NewTicker(400 * time.Millisecond)

	done := make(chan struct{})

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Println("report summary calculation in progress")

			case <-ctx.Done():
				log.Println("report summary calculation cancelled")
				return

			case <-done:
				log.Println("report summary calculation finished")
				return
			}
		}
	}()

	return func() {
		close(done)
	}
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, error) {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return uuid.Nil, domain.ErrUnauthenticated
	}

	values := md.Get("user_id")
	if len(values) == 0 {
		return uuid.Nil, domain.ErrUnauthenticated
	}

	userID, err := uuid.Parse(values[0])
	if err != nil {
		return uuid.Nil, domain.ErrUnauthenticated
	}
	return userID, nil
}

func (l *ledgerServiceImpl) invalidateReportCache(
	ctx context.Context,
	userID uuid.UUID,
) {
	pattern := fmt.Sprintf("report:summary:%s:*", userID.String())
	l.cache.DeleteByPattern(ctx, pattern)
}

func (l *ledgerServiceImpl) invalidateBudgetsCache(
	ctx context.Context,
	userID uuid.UUID,
) {
	key := fmt.Sprintf("budgets:all:%s", userID.String())
	l.cache.Delete(ctx, key)
}
