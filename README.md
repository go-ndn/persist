# Simple, fast, and reliable NDN persistent content store

While LRU in-memory cache is usually good enough for simple NDN application, persistent disk-based content store is required if an application has a huge dataset.

It uses [Bolt](https://github.com/boltdb/bolt), and implements `ndn.Cache` interface.
