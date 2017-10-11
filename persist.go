// Package persist implements persistent content store.
package persist

import (
	"crypto/sha256"
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

// New creates a persistent content store.
//
// By design, reading from this store is significantly
// faster than writing new entry.
// This store will not delete any old entry.
// When the file is opened, invoking New again will block.
func New(file string) (ndn.Cache, error) {
	db, err := bolt.Open(file, 0600, nil)
	if err != nil {
		return nil, err
	}
	return &cache{DB: db}, nil
}

type entry struct {
	Data *ndn.Data `tlv:"2"`
	Time time.Time `tlv:"3"`
}

func (c *cache) Add(d *ndn.Data) {
	c.Update(func(tx *bolt.Tx) error {
		h := sha256.New()
		err := d.WriteTo(tlv.NewWriter(h))
		if err != nil {
			return err
		}

		b, err := tlv.Marshal(entry{
			Data: d,
			Time: time.Now(),
		}, 1)
		if err != nil {
			return err
		}

		bucket, err := tx.CreateBucketIfNotExists(mainBucket)
		if err != nil {
			return err
		}
		for _, component := range d.Name.Components {
			bucket, err = bucket.CreateBucketIfNotExists(component)
			if err != nil {
				return err
			}
		}
		return bucket.Put(h.Sum(nil), b)
	})
}

func (c *cache) Get(i *ndn.Interest) (match *ndn.Data) {
	sel := func(v []byte) bool {
		if v == nil {
			return false
		}
		var ent entry
		err := tlv.Unmarshal(v, &ent, 1)
		if err != nil {
			return false
		}
		if !i.Selectors.Match(ent.Data, i.Name.Len()) {
			return false
		}
		if i.Selectors.MustBeFresh && ent.Data.MetaInfo.FreshnessPeriod > 0 &&
			time.Since(ent.Time) > time.Duration(ent.Data.MetaInfo.FreshnessPeriod)*time.Millisecond {
			return false
		}
		match = ent.Data
		return true
	}

	var search func(bucket *bolt.Bucket) bool
	search = func(bucket *bolt.Bucket) bool {
		cursor := bucket.Cursor()

		// determine search order
		var first, next func() ([]byte, []byte)
		if i.Selectors.ChildSelector == 0 {
			first = cursor.First
			next = cursor.Next
		} else {
			first = cursor.Last
			next = cursor.Prev
		}

		// right-most: search nested buckets first
		if i.Selectors.ChildSelector != 0 {
			for k, v := first(); k != nil; k, v = next() {
				if v != nil {
					continue
				}
				if search(bucket.Bucket(k)) {
					return true
				}
			}
		}

		// search non-nested buckets
		for k, v := first(); k != nil; k, v = next() {
			if sel(v) {
				return true
			}
		}

		// left-most: search nested buckets last
		if i.Selectors.ChildSelector == 0 {
			for k, v := first(); k != nil; k, v = next() {
				if v != nil {
					continue
				}
				if search(bucket.Bucket(k)) {
					return true
				}
			}
		}

		return false
	}

	c.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(mainBucket)
		if bucket == nil {
			return nil
		}
		for _, component := range i.Name.Components {
			bucket = bucket.Bucket(component)
			if bucket == nil {
				return nil
			}
		}
		if len(i.Name.ImplicitDigestSHA256) != 0 {
			// directly search for implicit digest
			sel(bucket.Get(i.Name.ImplicitDigestSHA256))
			return nil
		}

		search(bucket)
		return nil
	})
	return
}
