package store

import (
	"github.com/hoangnecon/golab/internal/domain"
	"github.com/hoangnecon/golab/internal/store/sqlite"
)

// ErrNotFound is returned when a requested row does not exist.
var ErrNotFound = domain.ErrNotFound

// Store is an alias for sqlite.Store implementing domain repositories.
type Store = sqlite.Store

// New creates or opens the SQLite database.
var New = sqlite.New
