package xlsx

import (
	"fmt"
)

// Row represents a single Row in the current Sheet.
type Row struct {
	Hidden       bool    // Hidden determines whether this Row is hidden or not.
	Sheet        *Sheet  // Sheet is a reference back to the Sheet that this Row is within.
	Height       float64 // Height is the current height of the Row in PostScript Points
	OutlineLevel uint8   // OutlineLevel contains the outline level of this Row.  Used for collapsing.
	isCustom     bool    // isCustom is a flag that is set to true when the Row has been modified
	num          int     // Num hold the positional number of the Row in the Sheet
	cellCount    int     // The current number of cells
}

// SetHeight sets the height of the Row in PostScript Points
func (r *Row) SetHeight(ht float64) {
	r.Height = ht
	r.isCustom = true
}

// SetHeightCM sets the height of the Row in centimetres, inherently converting it to PostScript points.
func (r *Row) SetHeightCM(ht float64) {
	r.Height = ht * 28.3464567 // Convert CM to postscript points
	r.isCustom = true
}

// AddCell adds a new Cell to the Row
func (r *Row) AddCell() *Cell {
	cell := newCell(r, r.cellCount)
	r.cellCount++
	return cell
}

func (r *Row) makeCellKey(colIdx int) string {
	return fmt.Sprintf("%s:%06d:%06d", r.Sheet.Name, r.num, colIdx)
}

func (r *Row) makeCellKeyRowPrefix() string {
	return fmt.Sprintf("%s:%06d", r.Sheet.Name, r.num)
}

// GetCell returns the Cell at a given column index, creating it if it doesn't exist.
func (r *Row) GetCell(colIdx int) *Cell {
	key := r.makeCellKey(colIdx)
	cell, err := r.Sheet.cellStore.ReadCell(key)
	if err != nil {
		if _, ok := err.(*CellNotFoundError); !ok {
			panic(err)
		}
	}
	if cell != nil {
		// We don't persist the row within the cell, but in it's key,
		// so we map the Row pointer back here.
		cell.Row = r
		return cell
	}
	cell = newCell(r, colIdx)
	if colIdx >= r.cellCount {
		r.cellCount = colIdx + 1
	}

	return cell
}

// ForEachCell will call the provided CellVisitorFunc for each
// currently defined cell in the Row.
func (r *Row) ForEachCell(cvf CellVisitorFunc) error {
	fn := func(c *Cell) error {
		c.Row = r
		return cvf(c)
	}

	return r.Sheet.cellStore.ForEachInRow(r, fn)
}
