package commands

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/spf13/cobra"
	"github.com/wcatron/query-projects/internal/projects"
	// still needed for withMetrics
)

var LoadCmd = &cobra.Command{
	Use:   "load [flags] <csv1> <csv2> …",
	Short: "Load one or more CSV files into an SQLite database",
	Args:  cobra.MinimumNArgs(0),
	RunE: withMetrics(func(cmd *cobra.Command, args []string) error {
		dbPath, _ := cmd.Flags().GetString("db")
		return CMD_loadCSVs(dbPath, args)
	}),
}

func LoadCmdInit(cmd *cobra.Command) {
	defaultDB := filepath.Join(projects.ResultsFolder, "results.db")
	cmd.Flags().StringP("db", "d", defaultDB, "Path to SQLite database file")
}

// CMD_loadCSVs loads each CSV in files into the SQLite DB.
func CMD_loadCSVs(dbPath string, files []string) error {
	fmt.Printf("Loading %s into %s\n", files, dbPath)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	// force the file to be created and connection verified
	if err := db.Ping(); err != nil {
		return fmt.Errorf("create db: %w", err)
	}
	defer db.Close()

	// If no files were provided on the CLI, load every *.csv in resultsFolder.
	if len(files) == 0 {
		pattern := filepath.Join(projects.ResultsFolder, "*.csv") // e.g. "results/*.csv"
		var err error
		files, err = filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("listing CSVs in %s: %w", projects.ResultsFolder, err)
		}
	}

	for _, f := range files {
		if err := loadSingleCSV(db, f); err != nil {
			return err
		}
	}
	return nil
}

// loadSingleCSV reads one CSV file and inserts it into db.
func loadSingleCSV(db *sql.DB, csvPath string) error {
	file, err := os.Open(csvPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", csvPath, err)
	}

	defer file.Close()

	r := csv.NewReader(file)
	headers, err := r.Read()
	if err != nil {
		return fmt.Errorf("read header of %s: %w", csvPath, err)
	}
	// sanitize headers: trim and replace spaces with underscores
	for i, h := range headers {
		headers[i] = strings.ReplaceAll(strings.TrimSpace(h), " ", "_")
	}

	if err != nil {
		return fmt.Errorf("read header of %s: %w", csvPath, err)
	}
	table := strings.TrimSuffix(filepath.Base(csvPath), filepath.Ext(csvPath))
	if err := ensureTable(db, table, headers); err != nil {
		return err
	}

	// INSERT statement like: INSERT INTO table (c1,c2,…) VALUES (?,?,…)
	placeholders := strings.Repeat("?,", len(headers))
	placeholders = placeholders[:len(placeholders)-1]
	stmtText := fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES (%s)`,
		table, strings.Join(headers, ","), placeholders)
	stmt, err := db.Prepare(stmtText)
	fmt.Printf("SQLITE: %s\n", stmtText)
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	records, err := r.ReadAll()
	if err != nil {
		return fmt.Errorf("read rows of %s: %w", csvPath, err)
	}
	for _, rec := range records {
		// Convert []string to []interface{}
		vals := make([]interface{}, len(rec))
		for i, v := range rec {
			vals[i] = v
		}
		if _, err := stmt.Exec(vals...); err != nil {
			return fmt.Errorf("insert row in %s: %w", table, err)
		}
	}
	fmt.Printf("Loaded %d rows from %s into table %s\n", len(records), csvPath, table)
	return nil
}

// ensureTable creates table if needed
func ensureTable(db *sql.DB, name string, cols []string) error {
	// 1) build the column defs for the first‑time CREATE
	colDefs := make([]string, len(cols))
	for i, c := range cols {
		colDefs[i] = fmt.Sprintf(`"%s" TEXT`, c)
	}
	// add qp_created_at with a default
	colDefs = append(colDefs, `"qp_created_at" TEXT DEFAULT (CURRENT_TIMESTAMP)`)

	_, err := db.Exec(fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS "%s" (%s);`,
		name, strings.Join(colDefs, ",")))
	if err != nil {
		return err
	}

	return nil
}
