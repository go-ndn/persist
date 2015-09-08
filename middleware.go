package persist

import (
	"github.com/go-ndn/mux"
	"github.com/go-ndn/ndn"
)

type cacher struct {
	ndn.Sender
	ndn.Cache
}

func (c *cacher) SendData(d *ndn.Data) {
	c.Add(d)
	c.Sender.SendData(d)
}

func (c *cacher) Hijack() ndn.Sender {
	return c.Sender
}

func Cacher(file string) mux.Middleware {
	c, err := New(file)
	if err != nil {
		panic(err)
	}
	return func(next mux.Handler) mux.Handler {
		return mux.HandlerFunc(func(w ndn.Sender, i *ndn.Interest) {
			d := c.Get(i)
			if d == nil {
				next.ServeNDN(&cacher{Sender: w, Cache: c}, i)
			} else {
				w.SendData(d)
			}
		})
	}
}
