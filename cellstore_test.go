package xlsx

import (
	"os"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestCellStore(t *testing.T) {
	c := qt.New(t)

	c.Run("Write and Read Style", func(c *qt.C) {
		cs := newCellStore()
		defer os.RemoveAll(cs.baseDir)

		file := NewFile()
		sheet, _ := file.AddSheet("Sheet1")
		row := sheet.AddRow()
		cell1 := row.AddCell()
		cell1.Value = "A cell!"
		style1 := NewStyle()
		style1.Border = *NewBorder("thin", "thin", "thin", "thin")
		style1.ApplyBorder = true
		cell1.SetStyle(style1)

		err := cs.writeStyle(cell1)
		c.Assert(err, qt.IsNil)

		style2, err := cs.readStyle(cell1.key())
		c.Assert(err, qt.IsNil)

		c.Assert(style1, qt.DeepEquals, style2)
	})

	c.Run("Write and Read DataValidation", func(c *qt.C) {
		var title = "cell"
		var msg = "cell msg"

		cs := newCellStore()
		defer os.RemoveAll(cs.baseDir)

		file := NewFile()
		sheet, _ := file.AddSheet("Sheet1")
		row := sheet.AddRow()
		cell := row.AddCell()

		dd := NewDataValidation(0, 0, 0, 0, true)
		err := dd.SetDropList([]string{"a1", "a2", "a3"})
		c.Assert(err, qt.IsNil)

		dd.SetInput(&title, &msg)
		cell.SetDataValidation(dd)

		err = cs.writeDataValidation(cell)
		c.Assert(err, qt.IsNil)

		dd2, err := cs.readDataValidation(cell.key())
		c.Assert(err, qt.IsNil)
		c.Assert(dd, qt.DeepEquals, dd2)
	})

	c.Run("Write and Read Cell", func(c *qt.C) {
		cs := newCellStore()
		defer os.RemoveAll(cs.baseDir)

		file := NewFile()
		sheet, _ := file.AddSheet("Sheet1")
		row := sheet.AddRow()
		cell := row.AddCell()
		cell.SetString("Blue as blue can be")

		err := cs.writeCell(cell)
		c.Assert(err, qt.IsNil)

		cell2, err := cs.readCell(cell.key())
		c.Assert(err, qt.IsNil)
		c.Assert(cell.String(), qt.Equals, cell2.String())
	})
}
