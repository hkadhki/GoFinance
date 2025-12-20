package main

import (
	"context"
	"ledger/internal/cache"
	"ledger/internal/db"
	"ledger/internal/db/sqlc"
	"ledger/internal/repository/pg"
	"ledger/internal/service"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	ledgergrpc "ledger/internal/grpc"
	ledgerv1 "ledger/ledger/v1"

	"github.com/pressly/goose/v3"
	"google.golang.org/grpc"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	database, err := db.Open(ctx)
	if err != nil {
		panic(err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}

	migrationsDir := "./migrations"
	if _, err := os.Stat(migrationsDir); err == nil {
		if err := goose.Up(database, migrationsDir); err != nil {
			log.Printf("Migration error: %v", err)
		}
	}

	cache.Init(ctx)
	q := sqlc.New(database)
	budgetRepo := pg.NewBudgetRepo(q)
	expenseRepo := pg.NewExpenseRepo(q)
	reportRepo := pg.NewReportRepo(q)

	svc := service.New(
		budgetRepo,
		expenseRepo,
		reportRepo,
	)
	closeFn := func() {
		if cache.Client != nil {
			_ = cache.Client.Close()
		}
		_ = database.Close()
	}

	defer closeFn()

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	addr := "0.0.0.0:" + port

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen %s: %v", addr, err)
	}
	grpcServer := grpc.NewServer()

	server := ledgergrpc.NewServer(svc)
	ledgerv1.RegisterLedgerServiceServer(grpcServer, server)

	log.Printf("Ledger gRPC server listening on %s", addr)

	go func() {
		<-ctx.Done()
		log.Println("shutting down gRPC server...")
		grpcServer.GracefulStop()
	}()

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC server stopped: %v", err)
	}
}
