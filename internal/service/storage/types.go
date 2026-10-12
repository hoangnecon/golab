package storage

import "time"

// FileEntry represents an entry inside a directory or a standalone file.
type FileEntry struct {
	Name         string      `json:"name"`
	Path         string      `json:"path"`
	Type         string      `json:"type"` // "file" or "directory"
	Size         *int64      `json:"size,omitempty"`
	Writable     bool        `json:"writable"`
	Created      *time.Time  `json:"created,omitempty"`
	LastModified *time.Time  `json:"last_modified,omitempty"`
	MIMEType     string      `json:"mimetype,omitempty"`
	Format       string      `json:"format,omitempty"` // "text", "base64", "json"
	Content      any         `json:"content,omitempty"`
}

// DirectoryListing represents the content returned when listing a directory.
type DirectoryListing struct {
	Path         string      `json:"path"`
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Writable     bool        `json:"writable"`
	LastModified *time.Time  `json:"last_modified,omitempty"`
	Entries      []FileEntry `json:"entries"`
	TotalEntries int         `json:"total_entries"`
}

// FileReadResult represents the result of reading a file.
type FileReadResult struct {
	Path         string     `json:"path"`
	Name         string     `json:"name"`
	Type         string     `json:"type"`
	Size         int64      `json:"size"`
	Format       string     `json:"format"`
	Content      string     `json:"content"`
	LastModified *time.Time `json:"last_modified,omitempty"`
}

// FileWriteResult represents the outcome of an atomic file write.
type FileWriteResult struct {
	Path         string     `json:"path"`
	Name         string     `json:"name"`
	Size         int64      `json:"size"`
	Type         string     `json:"type"`
	LastModified *time.Time `json:"last_modified,omitempty"`
	Created      bool       `json:"created"`
}
