package persist

import (
	"github.com/go-ndn/mux"
	"github.com/go-ndn/ndn"
)

type cacher struct {
	ndn.Sender
	cache ndn.Cache
}

func (c *cacher) SendData(d *ndn.Data) {
	c.Sender.SendData(d)
	go c.cache.Add(d)
}

func (c *cacher) Hijack() ndn.Sender {
	return c.Sender
}

func Cacher(file string) mux.Middleware {
	c, _ := New(file)
	return func(next mux.Handler) mux.Handler {
		return mux.HandlerFunc(func(w ndn.Sender, i *ndn.Interest) {
			d := c.Get(i)
			if d == nil {
				next.ServeNDN(&cacher{Sender: w, cache: c}, i)
			} else {
				w.SendData(d)
			}
		})
	}
}