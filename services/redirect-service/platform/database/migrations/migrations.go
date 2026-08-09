package migrations

// //go:embed sql/*.sql
// var migrationFS embed.FS

// // Run applies every pending migration. The DSN must use the pgx5 scheme and
// // name this service's ledger table; see config.MigrationDSN.
// func Run(dsn string) error {
// 	src, err := iofs.New(migrationFS, "sql")
// 	if err != nil {
// 		return fmt.Errorf("loading embedded migrations: %w", err)
// 	}

// 	m, err := migrate.NewWithSourceInstance("iofs", src, dsn)
// 	if err != nil {
// 		return fmt.Errorf("creating migrator: %w", err)
// 	}
// 	defer m.Close()

// 	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
// 		return fmt.Errorf("applying migrations: %w", err)
// 	}
// 	return nil
// }
