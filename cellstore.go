package xlsx

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"github.com/peterbourgon/diskv"
)

var (
	CellCacheSize uint64 = 1024 * 1024 // 1 MB per sheet
)

type CellStore struct {
	baseDir string
	buf     *bytes.Buffer
	store   *diskv.Diskv
	// enc     *gob.Encoder
	// dec     *gob.Decoder
	enc *json.Encoder
	dec *json.Decoder
}

func (cs *CellStore) writeStyle(c *Cell) error {
	if c.style != nil {
		cs.buf.Reset()
		err := cs.enc.Encode(*c.style)
		if err != nil {
			return err
		}
		return cs.store.WriteStream(c.key()+"-S", cs.buf, true)
	}
	return nil
}

func (cs *CellStore) readStyle(key string) (*Style, error) {
	b, err := cs.store.Read(key + "-S")
	if err != nil {
		return nil, err
	}
	cs.buf.Reset()
	_, err = cs.buf.Write(b)
	if err != nil {
		return nil, err
	}
	dv := &Style{}
	err = cs.dec.Decode(dv)
	if err != nil {
		return nil, err
	}
	return dv, nil
}

//
func (cs *CellStore) writeDataValidation(c *Cell) error {
	if c.DataValidation != nil {
		cs.buf.Reset()
		err := cs.enc.Encode(*c.DataValidation)
		if err != nil {
			return err
		}
		return cs.store.WriteStream(c.key()+"-DV", cs.buf, true)
	}
	return nil
}

func (cs *CellStore) readDataValidation(key string) (*xlsxDataValidation, error) {
	b, err := cs.store.Read(key + "-DV")
	if err != nil {
		return nil, err
	}
	cs.buf.Reset()
	_, err = cs.buf.Write(b)
	if err != nil {
		return nil, err
	}
	dv := &xlsxDataValidation{}
	err = cs.dec.Decode(dv)
	if err != nil {
		return nil, err
	}
	return dv, nil

}

func (cs *CellStore) writeCell(c *Cell) error {
	cs.buf.Reset()
	err := cs.enc.Encode(c)
	if err != nil {
		return err
	}
	key := c.key()
	return cs.store.WriteStream(key, cs.buf, true)
}

func (cs *CellStore) WriteCell(c *Cell) error {
	err := cs.writeDataValidation(c)
	if err != nil {
		return err
	}
	err = cs.writeStyle(c)
	if err != nil {
		return err
	}
	return cs.writeCell(c)
}

//
func (cs *CellStore) readCell(key string) (*Cell, error) {
	b, err := cs.store.Read(key)
	if err != nil {
		return nil, err
	}
	cs.buf.Reset()
	_, err = cs.buf.Write(b)
	if err != nil {
		return nil, err
	}
	c := &Cell{}
	err = cs.dec.Decode(c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (cs *CellStore) ReadCell(key string) (*Cell, error) {
	c, err := cs.readCell(key)
	if err != nil {
		if _, ok := err.(*os.PathError); !ok {
			// PathError occurs when the cell doesn't
			// exist in the diskv store yet, which is
			// fine.  Anything else is a disaster.
			return nil, err
		}
		return nil, nil
	}
	dv, err := cs.readDataValidation(key)
	if err != nil {
		if _, ok := err.(*os.PathError); !ok {
			// PathError occurs when the cell doesn't
			// exist in the diskv store yet, which is
			// fine.  Anything else is a disaster.
			return nil, err
		}
	}
	c.DataValidation = dv
	s, err := cs.readStyle(key)
	if err != nil {
		if _, ok := err.(*os.PathError); !ok {
			// PathError occurs when the cell doesn't
			// exist in the diskv store yet, which is
			// fine.  Anything else is a disaster.
			return nil, err
		}
	}
	c.style = s

	return c, nil
}

type CellVisitorFunc func(c *Cell) error

//
func (cs *CellStore) ForEach(cvf CellVisitorFunc) error {
	for key := range cs.store.Keys(nil) {
		c, err := cs.ReadCell(key)
		if err != nil {
			return err
		}
		err = cvf(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cs *CellStore) ForEachInRow(r *Row, cvf CellVisitorFunc) error {
	pref := r.makeCellKeyRowPrefix()
	for key := range cs.store.KeysPrefix(pref, nil) {
		c, err := cs.ReadCell(key)
		if err != nil {
			return err
		}
		err = cvf(c)
		if err != nil {
			return err
		}
	}
	return nil

}

func newCellStore() *CellStore {
	cs := &CellStore{
		buf: bytes.NewBuffer([]byte{}),
	}
	dir, err := ioutil.TempDir("", "cellstore")
	if err != nil {
		return nil
	}
	cs.baseDir = dir
	cs.store = diskv.New(diskv.Options{
		BasePath: dir,
		// Transform:    cellTransform,
		CacheSizeMax: CellCacheSize,
	})
	// cs.enc = gob.NewEncoder(cs.buf)
	// cs.dec = gob.NewDecoder(cs.buf)
	cs.enc = json.NewEncoder(cs.buf)
	cs.dec = json.NewDecoder(cs.buf)
	return cs
}

func cellTransform(s string) []string {
	return strings.Split(s, ":")
}
