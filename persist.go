package persist

import (
	"bytes"
	"time"

	"github.com/boltdb/bolt"
	"github.com/go-ndn/ndn"
	"github.com/go-ndn/tlv"
)

type cache struct {
	*bolt.DB
}

var (
	mainBucket = []byte("main")
)

func New(file string) (c ndn.Cache, err error) {
	db, err := bolt.Open(file, 0600, nil)
	if err != nil {
		return
	}
	db.Update(func(tx *bolt.Tx) (err error) {
		_, err = tx.CreateBucketIfNotExists(mainBucket)
		return
	})
	c = &cache{DB: db}
	return
}

type entry struct {
	Data *ndn.Data `tlv:"2"`
	Time time.Time `tlv:"3"`
}

func (c *cache) Add(d *ndn.Data) {
	c.Update(func(tx *bolt.Tx) (err error) {
		b, err := tlv.MarshalByte(entry{
			Data: d,
			Time: time.Now(),
		}, 1)
		if err != nil {
			return
		}
		err = tx.Bucket(mainBucket).Put([]byte(d.Name.String()), b)
		return
	})
}

func (c *cache) Get(i *ndn.Interest) (match *ndn.Data) {
	c.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(mainBucket).Cursor()
		prefix := []byte(i.Name.String())

		for k, v := c.Seek(prefix); bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var ent entry
			err := tlv.UnmarshalByte(v, &ent, 1)
			if err != nil {
				continue
			}
			if !i.Selectors.Match(string(k), ent.Data) {
				continue
			}
			if i.Selectors.MustBeFresh && time.Since(ent.Time) > time.Duration(ent.Data.MetaInfo.FreshnessPeriod)*time.Millisecond {
				continue
			}
			if i.Selectors.ChildSelector == 0 {
				match = ent.Data
				return nil
			}
			match = ent.Data
		}
		return nil
	})
	return
}
