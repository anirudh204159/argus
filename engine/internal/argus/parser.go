package argus

import (
	"fmt"

	"github.com/go-mysql-org/go-mysql/replication"
)

// parseRowsEvent converts a raw binlog RowsEvent into one or more Argus Events.
func parseRowsEvent(eventType replication.EventType, e *replication.RowsEvent) ([]Event, error) {
	var op string
	switch eventType {
	case replication.WRITE_ROWS_EVENTv2:
		op = "INSERT"
	case replication.UPDATE_ROWS_EVENTv2:
		op = "UPDATE"
	case replication.DELETE_ROWS_EVENTv2:
		op = "DELETE"
	default:
		return nil, nil
	}

	schema := string(e.Table.Schema)
	table := string(e.Table.Table)

	columns, err := getColumnNames(schema, table)
	if err != nil {
		return nil, err
	}

	var events []Event

	switch op {
	case "INSERT":
		for _, row := range e.Rows {
			events = append(events, Event{
				Operation: op,
				Schema:    schema,
				Table:     table,
				After:     rowToMap(columns, row),
			})
		}
	case "DELETE":
		for _, row := range e.Rows {
			events = append(events, Event{
				Operation: op,
				Schema:    schema,
				Table:     table,
				Before:    rowToMap(columns, row),
			})
		}
	case "UPDATE":
		for i := 0; i < len(e.Rows); i += 2 {
			events = append(events, Event{
				Operation: op,
				Schema:    schema,
				Table:     table,
				Before:    rowToMap(columns, e.Rows[i]),
				After:     rowToMap(columns, e.Rows[i+1]),
			})
		}
	}

	return events, nil
}

// rowToMap pairs column names with values into a map.
func rowToMap(columns []string, row []interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	for i, val := range row {
		name := fmt.Sprintf("col%d", i)
		if i < len(columns) {
			name = columns[i]
		}
		m[name] = val
	}
	return m
}
