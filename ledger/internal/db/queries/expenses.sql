-- name: GetBudgetLimit :one
SELECT limit_amount
FROM budgets
WHERE category = $1
FOR UPDATE;

-- name: GetSumByCategory :one
SELECT COALESCE(SUM(amount), 0)::DECIMAL(14,2)
FROM expenses
WHERE user_id = $1
  AND category = $2;

-- name: InsertExpense :one
INSERT INTO expenses (user_id, amount, category, description, date)
VALUES ($1, $2, $3, $4, $5)
    RETURNING id;

-- name: ListExpenses :many
SELECT id, user_id, amount, category, description, date
FROM expenses
WHERE user_id = $1
ORDER BY date DESC, id DESC;

-- name: SumByCategoryAndPeriod :one
SELECT COALESCE(SUM(amount), 0)::DECIMAL(14,2)
FROM expenses
WHERE user_id = sqlc.arg(user_id)
  AND category = sqlc.arg(category)
  AND date BETWEEN sqlc.arg(from_date) AND sqlc.arg(to_date);