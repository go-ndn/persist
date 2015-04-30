package persist

import (
	"github.com/go-ndn/mux"
	"github.com/go-ndn/ndn"
)

type cacher struct {
	mux.Sender
	cache ndn.Cache
}

func (c *cacher) SendData(d *ndn.Data) {
	c.Sender.SendData(d)
	c.cache.Add(d)
}

func (c *cacher) Hijack() mux.Sender {
	return c.Sender
}

func Cacher(c *Cache) mux.Middleware {
	return func(next mux.Handler) mux.Handler {
		return mux.HandlerFunc(func(w mux.Sender, i *ndn.Interest) {
			d := c.Get(i)
			if d == nil {
				next.ServeNDN(&cacher{Sender: w, cache: c}, i)
			} else {
				w.SendData(d)
			}
		})
	}
}
