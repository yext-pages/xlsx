package xlsx

import (
	"sort"
	"strings"
)

type cellRef map[string]*Cell

type MemoryCellStore struct {
	rows  map[string]cellRef
	cells map[string]*Cell
}

// NewMemoryCellStore returns a pointer to a newly allocated MemoryCellStore
func NewMemoryCellStore() (CellStore, error) {
	cs := &MemoryCellStore{
		rows:  make(map[string]cellRef),
		cells: make(map[string]*Cell),
	}
	return cs, nil
}

//
func (mcs *MemoryCellStore) Close() error {
	return nil
}

//
func (mcs *MemoryCellStore) WriteCell(cell *Cell) error {
	key := cell.key()
	row := cell.Row.makeCellKeyRowPrefix()
	cref, found := mcs.rows[row]
	if !found {
		cref = make(cellRef)
	}
	cref[key] = cell
	mcs.rows[row] = cref
	mcs.cells[cell.key()] = cell
	return nil
}

//
func (mcs *MemoryCellStore) ReadCell(key string) (*Cell, error) {
	cell, found := mcs.cells[key]
	if !found {
		return nil, NewCellNotFoundError(key, "No such cell")
	}

	return cell, nil
}

//
func (mcs *MemoryCellStore) DeleteCell(key string) error {
	delete(mcs.cells, key)
	delete(mcs.rows, keyToRowKey(key))
	return nil
}

//
func (mcs *MemoryCellStore) ForEachInRow(row *Row, cvf CellVisitorFunc) error {
	refs := mcs.rows[row.makeCellKeyRowPrefix()]
	l := row.cellCount
	keys := make([]string, l, l)

	// We have to collect the keys first so we can sort them and
	// visit the cells in order.
	i := 0
	for k, _ := range refs {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for _, k := range keys {
		c, ok := refs[k]
		if ok {
			err := cvf(c)
			if err != nil {
				return err
			}

		}

	}
	return nil
}

// Extract the row key from a provided cell key
func keyToRowKey(key string) string {
	parts := strings.Split(key, ":")
	return parts[0] + ":" + parts[1]
}
