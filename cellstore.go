package xlsx

import "fmt"

// CellStore provides an interface for interacting with backend cell
// storage. This allows us, as required, to persist cells to some
// store instead of holding them in memory.  This tactic allows us a
// degree of control around the characteristics of our programs when
// handling large spreadsheets - we can choose to run more slowly, but
// without exhausting system memory.
type CellStore interface {
	// ReadCell returns a Cell, from the CellStore using a unique
	// key.  When implementing a CellStore you shouldn't need to
	// worry about what this key is, or how it is generated.
	ReadCell(key string) (*Cell, error)
	// WriteCell persists a Cell to the CellStore.  The Cell must
	// be accessible in future via a key string, which you must
	// generate from the Cell using the Cell's key() method.  The
	// structure and procedure of generation of such a key should
	// not be the concern of the CellStore, but if for some reason
	// you find yourself having to escape characters in the key,
	// you must be certain to make this processing symmetrical
	// with the processing in ReadCell.
	WriteCell(cell *Cell) error
	// DeleteCell removes the cell stored with the provided key.
	DeleteCell(key string) error
	// ForEachInRow will visit each Cell in a Row and call the
	// provided CellVisitorFunc with a pointer to that Cell as its
	// only argument.  NOTE: cells that are omitted from the
	// underlying file, and/or have not been persisted in the
	// CellStore will not be visited.
	ForEachInRow(r *Row, cvf CellVisitorFunc) error
	// Close is called when the Sheet containing the CellStore is
	// closed.  This indicates that no further work will happen -
	// your implementation can safely delete its stored data at
	// this point.
	Close() error
}

// CellStoreConstructor defines the signature of a function that will
// be used to return a new instance of the CellStore implmentation,
// you must pass this into
type CellStoreConstructor func() (CellStore, error)

// CellVisitorFunc defines the signature of a function that will be
// called when visiting a Cell using CellStore.ForEachInRow.
type CellVisitorFunc func(c *Cell) error

// CellNotFoundError is an Error that should be returned by a
// CellStore implementation if a call to ReadCell is made with a key
// that doesn't correspond to any persisted Cell.
type CellNotFoundError struct {
	key    string
	reason string
}

// NewCellNotFoundError creates a new CellNotFoundError, capturing the Cell key and the reason this key could not be found.
//
func NewCellNotFoundError(key, reason string) *CellNotFoundError {
	return &CellNotFoundError{key, reason}
}

// Error returns a human-readable description of the failure to find a Cell.  It makes CellNotFoundError comply with the Error interface.
func (cnfe CellNotFoundError) Error() string {
	return fmt.Sprintf("Cell %q not found. %s", cnfe.key, cnfe.reason)
}
