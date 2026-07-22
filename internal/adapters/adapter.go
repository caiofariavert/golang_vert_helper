package adapters

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/gorm"

	"github.com/vert/golang_vert_helper/internal/domain"
)

// MigrationRunner handles database migrations
type MigrationRunner struct {
	db *sql.DB
	m  *migrate.Migrate
}

// NewMigrationRunner creates a new migration runner
func NewMigrationRunner(gormDB *gorm.DB, migrationsPath string) (*MigrationRunner, error) {
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get SQL DB: %w", err)
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration instance: %w", err)
	}

	return &MigrationRunner{
		db: sqlDB,
		m:  m,
	}, nil
}

// Up runs all pending migrations
func (mr *MigrationRunner) Up() error {
	if err := mr.m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration up failed: %w", err)
	}
	return nil
}

// Down reverts all migrations
func (mr *MigrationRunner) Down() error {
	if err := mr.m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration down failed: %w", err)
	}
	return nil
}

// Version returns the current migration version
func (mr *MigrationRunner) Version() (uint, error) {
	version, dirty, err := mr.m.Version()
	if err != nil {
		return 0, fmt.Errorf("failed to get version: %w", err)
	}

	if dirty {
		return 0, fmt.Errorf("migration state is dirty")
	}

	return version, nil
}

// ========== Application Initializer ==========

// ApplicationInitializer handles initialization of the entire application
type ApplicationInitializer struct {
	db    *gorm.DB
	repos *RepositoryFactory
}

// NewApplicationInitializer creates a new application initializer
func NewApplicationInitializer(config *Config) (*ApplicationInitializer, error) {
	// Create database adapter
	dsn := config.GetPostgresDSN()
	dbAdapter, err := NewDatabaseAdapter(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create database adapter: %w", err)
	}

	// Run auto migrations
	if err := dbAdapter.AutoMigrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create repository factory
	repos := NewRepositoryFactory(dbAdapter.DB)

	return &ApplicationInitializer{
		db:    dbAdapter.DB,
		repos: repos,
	}, nil
}

// GetRepositoryFactory returns the repository factory
func (ai *ApplicationInitializer) GetRepositoryFactory() *RepositoryFactory {
	return ai.repos
}

// GetDatabase returns the GORM database connection
func (ai *ApplicationInitializer) GetDatabase() *gorm.DB {
	return ai.db
}

// Close closes the database connection
func (ai *ApplicationInitializer) Close() error {
	sqlDB, err := ai.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Health checks the database connection
func (ai *ApplicationInitializer) Health(ctx context.Context) *domain.HealthCheckResult {
	if err := ai.db.WithContext(ctx).Exec("SELECT 1").Error; err != nil {
		return &domain.HealthCheckResult{
			Status:  domain.HealthStatusUnhealthy,
			Message: fmt.Sprintf("Database health check failed: %v", err),
		}
	}

	return &domain.HealthCheckResult{
		Status:  domain.HealthStatusHealthy,
		Message: "Database is healthy",
	}
}
