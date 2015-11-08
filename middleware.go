package persist

import "github.com/go-ndn/mux"

func Cacher(file string) mux.Middleware {
	c, err := New(file)
	if err != nil {
		panic(err)
	}
	return mux.RawCacher(c, false)
}
