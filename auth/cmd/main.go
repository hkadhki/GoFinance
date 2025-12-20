package main

import (
	authpb "auth/auth/v1"
	authgrpc "auth/internal/grpc"
	"auth/internal/repository/pg"
	"auth/internal/service"

	"database/sql"
	"log"
	"net"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"google.golang.org/grpc"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}

	migrationsDir := "./migrations"
	if _, err := os.Stat(migrationsDir); err == nil {
		if err := goose.Up(db, migrationsDir); err != nil {
			log.Printf("Migration error: %v", err)
		}
	}

	repo := pg.New(db)
	svc := service.New(repo)

	grpcServer := grpc.NewServer()

	authHandler := authgrpc.New(svc)
	authpb.RegisterAuthServiceServer(grpcServer, authHandler)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Auth gRPC listening on :50052")
	log.Fatal(grpcServer.Serve(lis))
}
