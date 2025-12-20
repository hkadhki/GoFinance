-- +goose Up

CREATE TABLE budgets (
                         id           SERIAL PRIMARY KEY,
                         user_id      UUID NOT NULL,
                         category     TEXT NOT NULL,
                         limit_amount NUMERIC(14,2) NOT NULL
                             CHECK (limit_amount > 0),
                         period       TEXT NOT NULL DEFAULT 'monthly',

                         UNIQUE (user_id, category)
);

CREATE TABLE expenses (
                          id SERIAL PRIMARY KEY,
                          user_id UUID NOT NULL,
                          amount DECIMAL(14,2) NOT NULL CHECK (amount <> 0),
                          category TEXT NOT NULL,
                          description TEXT,
                          date DATE NOT NULL
);

-- +goose Down

DROP TABLE IF EXISTS expenses;
DROP TABLE IF EXISTS budgets;