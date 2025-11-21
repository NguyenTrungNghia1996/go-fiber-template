package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONB wraps a slice so it can be stored as JSONB in Postgres.
type JSONB[T any] []T

// Value marshals the slice for storage.
func (j JSONB[T]) Value() (driver.Value, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return json.Marshal([]T(j))
}

// Scan unmarshals JSONB data into the slice.
func (j *JSONB[T]) Scan(src interface{}) error {
	if src == nil {
		*j = nil
		return nil
	}
	switch v := src.(type) {
	case string:
		return json.Unmarshal([]byte(v), j)
	case []byte:
		return json.Unmarshal(v, j)
	default:
		return fmt.Errorf("cannot scan type %T into JSONB", src)
	}
}
