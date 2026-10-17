// cryptodata migrates stored plaintext or rotates data encryption keys offline.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"hnu-tickets/internal/platform/config"
	"hnu-tickets/internal/platform/database"
	"hnu-tickets/internal/platform/fieldcrypt"
)

func main() {
	mode := flag.String("mode", "", "migrate (plaintext database) or rotate (encrypted database)")
	flag.Parse()
	if err := run(*mode); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(mode string) error {
	if mode != "migrate" && mode != "rotate" {
		return fmt.Errorf("specify -mode migrate or -mode rotate")
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	pool, err := database.Open(ctx, cfg.Database.URL())
	if err != nil {
		return err
	}
	defer pool.Close()
	n, err := fieldcrypt.Rewrite(ctx, pool, cfg.DataKeys, mode)
	if err != nil {
		return err
	}
	fmt.Printf("Encrypted %d field values.\n", n)
	return nil
}
