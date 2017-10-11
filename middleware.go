package persist

import "github.com/go-ndn/mux"

// Cacher creates a new caching middleware instance
// that uses the given file as the persistent content store.
func Cacher(file string) mux.Middleware {
	c, err := New(file)
	if err != nil {
		panic(err)
	}
	return mux.RawCacher(&mux.CacherOptions{
		Cache: c,
	})
}
