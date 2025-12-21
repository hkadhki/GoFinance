package domain

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestBudgetValidate(t *testing.T) {
	tests := []struct {
		name    string
		budget  Budget
		wantErr bool
		field   string
		message string
	}{
		{
			name: "valid budget monthly",
			budget: Budget{
				Category: "food",
				Limit:    decimal.NewFromInt(100),
				Period:   "monthly",
			},
			wantErr: false,
		},
		{
			name: "empty category",
			budget: Budget{
				Category: "",
				Limit:    decimal.NewFromInt(100),
				Period:   "monthly",
			},
			wantErr: true,
			field:   "category",
			message: "must not be empty",
		},
		{
			name: "zero limit",
			budget: Budget{
				Category: "food",
				Limit:    decimal.Zero,
				Period:   "monthly",
			},
			wantErr: true,
			field:   "limit",
			message: "must be positive",
		},
		{
			name: "negative limit",
			budget: Budget{
				Category: "food",
				Limit:    decimal.NewFromInt(-10),
				Period:   "monthly",
			},
			wantErr: true,
			field:   "limit",
			message: "must be positive",
		},
		{
			name: "invalid period",
			budget: Budget{
				Category: "food",
				Limit:    decimal.NewFromInt(100),
				Period:   "yearly",
			},
			wantErr: true,
			field:   "period",
			message: "can be either daily , monthly or weekly",
		},
		{
			name: "empty period allowed",
			budget: Budget{
				Category: "food",
				Limit:    decimal.NewFromInt(100),
				Period:   "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.budget.Validate()

			if !tt.wantErr {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)

			vErr, ok := err.(*ValidationError)
			require.True(t, ok, "error must be ValidationError")

			require.Equal(t, tt.field, vErr.Field)
			require.Equal(t, tt.message, vErr.Message)
		})
	}
}

func TestTransactionValidate(t *testing.T) {
	tests := []struct {
		name    string
		tx      Transaction
		wantErr bool
		field   string
		message string
	}{
		{
			name: "valid transaction",
			tx: Transaction{
				Amount:   decimal.NewFromInt(10),
				Category: "food",
				Date:     time.Now(),
			},
			wantErr: false,
		},
		{
			name: "zero amount",
			tx: Transaction{
				Amount:   decimal.Zero,
				Category: "food",
				Date:     time.Now(),
			},
			wantErr: true,
			field:   "amount",
			message: "must be positive",
		},
		{
			name: "negative amount",
			tx: Transaction{
				Amount:   decimal.NewFromInt(-5),
				Category: "food",
				Date:     time.Now(),
			},
			wantErr: true,
			field:   "amount",
			message: "must be positive",
		},
		{
			name: "empty category",
			tx: Transaction{
				Amount:   decimal.NewFromInt(10),
				Category: "",
				Date:     time.Now(),
			},
			wantErr: true,
			field:   "category",
			message: "must not be empty",
		},
		{
			name: "zero date",
			tx: Transaction{
				Amount:   decimal.NewFromInt(10),
				Category: "food",
				Date:     time.Time{},
			},
			wantErr: true,
			field:   "date",
			message: "must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tx.Validate()

			if !tt.wantErr {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)

			vErr, ok := err.(*ValidationError)
			require.True(t, ok, "error must be ValidationError")

			require.Equal(t, tt.field, vErr.Field)
			require.Equal(t, tt.message, vErr.Message)
		})
	}
}
