package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authv1 "gateway/auth/v1"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type mockAuthClient struct {
	validateFn func(ctx context.Context, in *authv1.ValidateRequest, opts ...grpc.CallOption) (*authv1.ValidateResponse, error)
}

func (m *mockAuthClient) Validate(
	ctx context.Context,
	in *authv1.ValidateRequest,
	opts ...grpc.CallOption,
) (*authv1.ValidateResponse, error) {
	return m.validateFn(ctx, in, opts...)
}

func (m *mockAuthClient) Register(context.Context, *authv1.RegisterRequest, ...grpc.CallOption) (*authv1.AuthResponse, error) {
	panic("not used")
}
func (m *mockAuthClient) Login(context.Context, *authv1.LoginRequest, ...grpc.CallOption) (*authv1.AuthResponse, error) {
	panic("not used")
}

func TestNewJWT_OK(t *testing.T) {
	client := &mockAuthClient{
		validateFn: func(ctx context.Context, in *authv1.ValidateRequest, _ ...grpc.CallOption) (*authv1.ValidateResponse, error) {
			require.Equal(t, "valid-token", in.Token)
			return &authv1.ValidateResponse{
				UserId: "user-123",
				Valid:  true,
			}, nil
		},
	}

	handler := NewJWT(client)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		require.True(t, ok)
		require.Equal(t, "user-123", userID)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestNewJWT_MissingHeader(t *testing.T) {
	client := &mockAuthClient{}

	handler := NewJWT(client)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestNewJWT_InvalidToken(t *testing.T) {
	client := &mockAuthClient{
		validateFn: func(ctx context.Context, in *authv1.ValidateRequest, _ ...grpc.CallOption) (*authv1.ValidateResponse, error) {
			return &authv1.ValidateResponse{
				Valid: false,
			}, nil
		},
	}

	handler := NewJWT(client)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestGetUserID_OK(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDKey, "user-42")

	id, ok := GetUserID(ctx)
	require.True(t, ok)
	require.Equal(t, "user-42", id)
}

func TestGetUserID_Missing(t *testing.T) {
	id, ok := GetUserID(context.Background())
	require.False(t, ok)
	require.Empty(t, id)
}

func TestTimeoutMiddleware(t *testing.T) {
	handler := TimeoutMiddleware(10 * time.Millisecond)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
				w.WriteHeader(http.StatusGatewayTimeout)
			case <-time.After(50 * time.Millisecond):
				w.WriteHeader(http.StatusOK)
			}
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusGatewayTimeout, rr.Code)
}

func TestTimeoutMiddleware_Default(t *testing.T) {
	handler := TimeoutMiddleware(0)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestLogging(t *testing.T) {
	handler := Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}
