package main

type Cache[K comparable, V any] struct {
	data     map[K]V
	capacity int
}

func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	var data map[K]V

	if capacity > 0 {
		data = make(map[K]V, capacity)
	}

	return &Cache[K, V]{
		data:     data,
		capacity: capacity,
	}
}

func (c *Cache[K, V]) Set(key K, value V) {
	if c.capacity <= 0 {
		return
	}

	c.data[key] = value
}

func (c *Cache[K, V]) Get(key K) (any, bool) {
	if c.capacity <= 0 || c.data == nil {
		return nil, false
	}

	val, ok := c.data[key]
	return val, ok
}
