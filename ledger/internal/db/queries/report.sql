-- name: ReportSummary :many
SELECT
    category,
    COALESCE(SUM(amount), 0)::DECIMAL(14,2) AS total
FROM expenses
WHERE user_id = $1
  AND date BETWEEN sqlc.arg(from_date) AND sqlc.arg(to_date)
ORDER BY category;