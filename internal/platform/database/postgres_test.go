package database

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewPoolUnreachableDatabase(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Port 1 has nothing listening, so the connection is refused imediately
	pool, err := NewPool(ctx, "postgres://app:s3cret-pw@127.0.0.1:1/appdb?sslmode=disable")
	if err == nil {
		pool.Close()
		t.Fatalf("NewPool succeeded against an unreachable database")
	}
	if strings.Contains(err.Error(), "s3cret-pw") {
		t.Errorf("Error leaks the password: %v", err)
	}
}

func TestNewPoolInvalidURL(t *testing.T) {
	if _, err := NewPool(context.Background(), "postgres://app:pw@localhost:notaport/appdb"); err == nil {
		t.Fatalf("NewPool accepted and invalid URL")
	}
}

func TestNewPoolInvalidScheme(t *testing.T) {
	if _, err := NewPool(context.Background(), "mysql://app:pw@localhost:5432/appdb"); err == nil {
		t.Fatalf("NewPool accepted an invalid scheme")
	}
}

// TestNewPoolLiveDatabase runs against the real database from DATABASE_URL.
// It is skipped when the variable is not set, so the plain unit-test run (and
// CI, until it has a database) does not need PostgreSQL.
func TestNewPoolLiveDatabase(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := NewPool(ctx, url)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}

	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Errorf("Ping: %v", err)
	}
}
