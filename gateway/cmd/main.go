// @title GoFinance Gateway API
// @version 1.0
// @description HTTP Gateway for personal finance microservices

// @contact.name GoFinance

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	authv1 "gateway/auth/v1"
	_ "gateway/docs"
	"gateway/internal/handlers"
	"gateway/internal/middleware"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"

	ledgerv1 "gateway/ledger/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ledgerAddr := os.Getenv("LEDGER_ADDR")
	if ledgerAddr == "" {
		ledgerAddr = "localhost:50051"
	}

	ledgerConn, err := grpc.Dial(
		ledgerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to connect to ledger: %v", err)
	}
	defer ledgerConn.Close()

	ledgerClient := ledgerv1.NewLedgerServiceClient(ledgerConn)

	authAddr := os.Getenv("AUTH_ADDR")
	if authAddr == "" {
		authAddr = "localhost:50052"
	}

	authConn, err := grpc.Dial(
		authAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to connect to auth: %v", err)
	}
	defer authConn.Close()

	authClient := authv1.NewAuthServiceClient(authConn)

	hLedger := handlers.NewHandler(ledgerClient)
	hAuth := handlers.NewAuthHandler(authClient)

	mux := http.NewServeMux()
	auth := http.NewServeMux()

	auth.HandleFunc("/auth/register", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			hAuth.Register(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	auth.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			hAuth.Login(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/transactions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			hLedger.CreateTransaction(w, r)
		case http.MethodGet:
			hLedger.ListTransactions(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/budgets", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			hLedger.CreateBudget(w, r)
		case http.MethodGet:
			hLedger.ListBudget(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/reports/summary", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			hLedger.ReportSummary(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/transactions/bulk", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			hLedger.BulkCreateTransactions(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/transactions/import", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			hLedger.ImportCSV(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/transactions/export", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			hLedger.ExportCSV(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	protected := middleware.NewJWT(authClient)(mux)

	handler := middleware.Logging(
		middleware.TimeoutMiddleware(2 * time.Second)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/auth/login" ||
					r.URL.Path == "/auth/register" ||
					strings.HasPrefix(r.URL.Path, "/swagger/") {

					if strings.HasPrefix(r.URL.Path, "/swagger/") {
						mux.ServeHTTP(w, r)
						return
					}

					auth.ServeHTTP(w, r)
					return
				}
				protected.ServeHTTP(w, r)
			}),
		),
	)

	log.Println("Gateway started on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))

}
