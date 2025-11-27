package models

import "time"

// UnitUploadedFile represents an object stored under a unit's namespace.
type UnitUploadedFile struct {
	Key          string    `json:"key"`
	Filename     string    `json:"filename"`
	Size         int64     `json:"size"`
	ContentType  string    `json:"content_type,omitempty"`
	ETag         string    `json:"etag,omitempty"`
	LastModified time.Time `json:"last_modified"`
	PublicURL    string    `json:"public_url,omitempty"`
}
