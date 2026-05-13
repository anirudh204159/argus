package argus

import (
	"database/sql"

	"github.com/go-mysql-org/go-mysql/mysql"
)

// schemaDB is the database connection used for metadata queries (column names, position).
var schemaDB *sql.DB

// columnCache stores resolved column names per "schema.table" so we only query MySQL once per table.
var columnCache = make(map[string][]string)

// getColumnNames returns the ordered column list for a given table, using the cache when possible.
func getColumnNames(schema, table string) ([]string, error) {
	key := schema + "." + table
	if cached, ok := columnCache[key]; ok {
		return cached, nil
	}

	rows, err := schemaDB.Query(`
		SELECT column_name
		FROM information_schema.columns
		WHERE table_schema = ? AND table_name = ?
		ORDER BY ordinal_position`, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			return nil, err
		}
		columns = append(columns, col)
	}

	columnCache[key] = columns
	return columns, nil
}

// getCurrentPosition asks MySQL for its current binlog file and position.
func getCurrentPosition() (mysql.Position, error) {
	var file string
	var position uint32
	var binlogDoDB, binlogIgnoreDB, executedGtidSet sql.NullString

	err := schemaDB.QueryRow("SHOW MASTER STATUS").Scan(&file, &position, &binlogDoDB, &binlogIgnoreDB, &executedGtidSet)
	if err != nil {
		return mysql.Position{}, err
	}

	return mysql.Position{Name: file, Pos: position}, nil
}
