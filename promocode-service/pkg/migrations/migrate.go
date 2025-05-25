package migrations

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gitlab.com/pisya-dev/promo-code-service/internal/config"
)

func Migrate(config *config.Config) error {
	migraton, err := migrate.New(
		"file://././migrations",
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			config.PostgresUser,
			config.PostgresPassword,
			config.PostgresHost,
			config.PostgresPort,

			config.PostgresDb,
		))
	if err != nil {
		return fmt.Errorf("unable to create migrations: %w", err)
	}

	if err := migraton.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf(":unable to run migrations: %w", err)
	}

	return nil
}
