package main

// Event is Argus's normalized representation of a single row change.
type Event struct {
	Operation string                 // "INSERT" / "UPDATE" / "DELETE"
	Schema    string                 // e.g. "argus_demo"
	Table     string                 // e.g. "orders"
	Before    map[string]interface{} // row data before change (nil for INSERT)
	After     map[string]interface{} // row data after change (nil for DELETE)
}
