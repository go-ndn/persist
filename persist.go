package persist

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/boltdb/bolt"
	"github.com/go-ndn/ndn"
)

type Cache struct {
	*bolt.DB
}

var (
	mainBucket = []byte("main")
)

func New(file string) (c *Cache, err error) {
	db, err := bolt.Open(file, 0600, nil)
	if err != nil {
		return
	}
	c = &Cache{DB: db}
	c.Update(func(tx *bolt.Tx) (err error) {
		_, err = tx.CreateBucketIfNotExists(mainBucket)
		return
	})
	return
}

type entry struct {
	Data *ndn.Data
	Time time.Time
}

func (c *Cache) Add(d *ndn.Data) {
	c.Update(func(tx *bolt.Tx) (err error) {
		buf := new(bytes.Buffer)
		enc := gob.NewEncoder(buf)
		err = enc.Encode(entry{
			Data: d,
			Time: time.Now(),
		})
		if err != nil {
			return
		}
		err = tx.Bucket(mainBucket).Put([]byte(d.Name.String()), buf.Bytes())
		return
	})
}

func (c *Cache) Get(i *ndn.Interest) (match *ndn.Data) {
	c.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(mainBucket).Cursor()
		prefix := []byte(i.Name.String())

		for k, v := c.Seek(prefix); bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var ent entry
			dec := gob.NewDecoder(bytes.NewReader(v))
			dec.Decode(&ent)
			if !i.Selectors.Match(string(k), ent.Data, ent.Time) {
				continue
			}
			if match == nil {
				match = ent.Data
			} else {
				cmp := ent.Data.Name.Compare(match.Name)
				switch i.Selectors.ChildSelector {
				case 0:
					if cmp < 0 {
						match = ent.Data
					}
				case 1:
					if cmp > 0 {
						match = ent.Data
					}
				}
			}
		}
		return nil
	})
	return
}
