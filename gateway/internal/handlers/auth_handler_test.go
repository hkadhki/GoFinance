package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	authv1 "gateway/auth/v1"
	ledgerv1 "gateway/ledger/v1"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type mockAuthClient struct {
	authv1.AuthServiceClient
	register func(ctx context.Context, in *authv1.RegisterRequest, opts ...grpc.CallOption) (*authv1.AuthResponse, error)
	login    func(ctx context.Context, in *authv1.LoginRequest, opts ...grpc.CallOption) (*authv1.AuthResponse, error)
}

func (m *mockAuthClient) Register(
	ctx context.Context,
	in *authv1.RegisterRequest,
	opts ...grpc.CallOption,
) (*authv1.AuthResponse, error) {
	return m.register(ctx, in, opts...)
}

func (m *mockAuthClient) Login(
	ctx context.Context,
	in *authv1.LoginRequest,
	opts ...grpc.CallOption,
) (*authv1.AuthResponse, error) {
	return m.login(ctx, in, opts...)
}

type mockLedgerClient struct {
	ledgerv1.LedgerServiceClient
	list func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ledgerv1.ListTransactionsResponse, error)
}

func TestAuthRegister_Success(t *testing.T) {
	client := &mockAuthClient{
		register: func(ctx context.Context, in *authv1.RegisterRequest, _ ...grpc.CallOption) (*authv1.AuthResponse, error) {
			require.Equal(t, "test@mail.com", in.Email)
			require.Equal(t, "secret", in.Password)
			return &authv1.AuthResponse{
				AccessToken: "jwt-token",
			}, nil
		},
	}

	h := NewAuthHandler(client)

	body := `{"email":"test@mail.com","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.Register(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp authResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.Equal(t, "jwt-token", resp.Token)
}
func TestAuthLogin_InvalidJSON(t *testing.T) {
	h := NewAuthHandler(&mockAuthClient{})

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		bytes.NewBufferString(`{bad json}`),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.Login(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}
