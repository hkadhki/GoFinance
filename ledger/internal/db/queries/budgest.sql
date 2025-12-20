-- name: UpsertBudget :exec
INSERT INTO budgets (user_id, category, limit_amount, period)
VALUES ($1, $2, $3, $4)
    ON CONFLICT (user_id, category)
DO UPDATE SET
    limit_amount = EXCLUDED.limit_amount,
           period       = EXCLUDED.period;

-- name: ListBudgets :many
SELECT id, user_id, category, limit_amount, period
FROM budgets
WHERE user_id = $1
ORDER BY category;


-- name: GetByCategory :one
SELECT id, user_id, category, limit_amount, period
FROM budgets
WHERE user_id = $1
  AND category = $2;

-- name: ListExpenseCategories :many
SELECT DISTINCT category
FROM expenses
WHERE user_id = $1;