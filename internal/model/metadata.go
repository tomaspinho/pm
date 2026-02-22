package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// TaskMetadata holds arbitrary key-value pairs for task metadata.
type TaskMetadata map[string]any

// Scan implements sql.Scanner for database deserialization.
func (m *TaskMetadata) Scan(value any) error {
	if value == nil {
		*m = make(TaskMetadata)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan TaskMetadata: not a byte slice")
	}

	if len(bytes) == 0 {
		*m = make(TaskMetadata)
		return nil
	}

	return json.Unmarshal(bytes, m)
}

// Value implements driver.Valuer for database serialization.
func (m TaskMetadata) Value() (driver.Value, error) {
	if len(m) == 0 {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}

// Get retrieves a value from metadata by key.
func (m TaskMetadata) Get(key string) (any, bool) {
	val, ok := m[key]
	return val, ok
}

// GetString retrieves a string value from metadata.
func (m TaskMetadata) GetString(key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// Set sets a key-value pair in metadata.
func (m TaskMetadata) Set(key string, value any) {
	m[key] = value
}

// Delete removes a key from metadata.
func (m TaskMetadata) Delete(key string) {
	delete(m, key)
}

// Keys returns all keys in metadata.
func (m TaskMetadata) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
