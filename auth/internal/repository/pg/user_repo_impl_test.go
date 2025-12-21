package pg

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func setupDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db, mock
}

func TestCreateUser_Success(t *testing.T) {
	db, mock := setupDB(t)
	repo := New(db)

	mock.ExpectExec(`INSERT INTO users`).
		WithArgs(sqlmock.AnyArg(), "test@mail.com", "hash").
		WillReturnResult(sqlmock.NewResult(1, 1))

	id, err := repo.Create(context.Background(), "test@mail.com", "hash")

	require.NoError(t, err)
	require.NotEmpty(t, id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByEmail_Success(t *testing.T) {
	db, mock := setupDB(t)
	repo := New(db)

	rows := sqlmock.NewRows([]string{"id", "email", "password_hash"}).
		AddRow("id-123", "test@mail.com", "hash")

	mock.ExpectQuery(`SELECT id, email, password_hash FROM users`).
		WithArgs("test@mail.com").
		WillReturnRows(rows)

	u, err := repo.GetByEmail(context.Background(), "test@mail.com")

	require.NoError(t, err)
	require.Equal(t, "id-123", u.ID)
}

func TestGetByID_Success(t *testing.T) {
	db, mock := setupDB(t)
	repo := New(db)

	rows := sqlmock.NewRows([]string{"id", "email", "password_hash"}).
		AddRow("id-123", "test@mail.com", "hash")

	mock.ExpectQuery(`SELECT id, email, password_hash FROM users`).
		WithArgs("id-123").
		WillReturnRows(rows)

	u, err := repo.GetByID(context.Background(), "id-123")

	require.NoError(t, err)
	require.Equal(t, "test@mail.com", u.Email)
}
