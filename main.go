package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"task133-structload/internal/clock"
	"task133-structload/internal/httpapi"
	"task133-structload/internal/selfcheck"
	"task133-structload/internal/service"
	"task133-structload/internal/store"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	dbPath := flag.String("db", "structload.db", "SQLite database path")
	migrateOnly := flag.Bool("migrate-only", false, "run schema migration then exit")
	smoke := flag.Bool("smoke-test", false, "run selfcheck scenarios then exit")
	flag.Parse()

	st, err := store.Open(*dbPath)
	if err != nil {
		return err
	}
	defer st.Close()

	if *migrateOnly {
		log.Println("migration complete")
		return nil
	}
	if *smoke {
		// Use a fresh in-file db for the smoke test to keep it self-contained.
		tmpDB := *dbPath + ".smoke"
		os.Remove(tmpDB)
		os.Remove(tmpDB + "-wal")
		os.Remove(tmpDB + "-shm")
		out, err := selfcheck.Run(tmpDB)
		if err != nil {
			fmt.Fprintln(os.Stderr, out)
			return fmt.Errorf("smoke-test failed: %w", err)
		}
		fmt.Println("smoke-test: ok")
		return nil
	}

	svc := service.New(st, clock.Real{})
	handler := httpapi.New(svc)

	srv := &http.Server{
		Addr:              *addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("task133-structload listening on %s (db=%s)", *addr, *dbPath)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-errCh:
		return err
	case sig := <-sigCh:
		log.Printf("received signal %s, shutting down", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(ctx)
	}
}
