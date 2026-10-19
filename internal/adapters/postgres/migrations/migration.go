package migrations

import (
	"context"
	"embed"
	"fmt"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// go:embed migrations/*.sql
var migrationFS embed.FS

const migrationLockKey int64 = 8472931

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	// Keep the same PostgreSQL connection for the session-level lock.
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire migration connection: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", migrationLockKey); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}

	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, _ = conn.Exec(
			unlockCtx,
			"SELECT pg_advisory_unlock($1)",
			migrationLockKey,
		)
	}()

	// Track which migrations have already been applied.
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			name TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("create migration ledger: %w", err)
	}

	files, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migration files: %w", err)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	for _, file := range files {
		name := file.Name()

		// Down migrations are never run automatically.
		if file.IsDir() || !strings.HasSuffix(name, ".up.sql") {
			continue
		}

		var applied bool
		err := conn.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM schema_migrations
				WHERE name = $1
			)
		`, name).Scan(&applied)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}

		if applied {
			continue
		}

		sqlBytes, err := migrationFS.ReadFile(path.Join("migrations", name))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		tx, err := conn.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", name, err)
		}

		// Execute the complete SQL file and record it in the same transaction.
		// pgx uses the simple protocol for a query without arguments, allowing
		// multiple SQL statements in one migration file.
		if _, err = tx.Exec(ctx, string(sqlBytes)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("execute migration %s: %w", name, err)
		}

		if _, err = tx.Exec(ctx, `
			INSERT INTO schema_migrations (name)
			VALUES ($1)
		`, name); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record migration %s: %w", name, err)
		}

		if err = tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}
	}

	return nil
}
