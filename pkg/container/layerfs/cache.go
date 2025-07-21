package layerfs

import "github.com/rs/zerolog/log"

type Node[T any] struct {
	key   string
	value T
	next  *Node[T]
	prev  *Node[T]
}

type LRUCache[T any] struct {
	lookup  map[string]*Node[T]
	first   *Node[T]
	last    *Node[T]
	maxSize int
}

func (lc *LRUCache[T]) update(key string) *Node[T] {
	node, ok := lc.lookup[key]
	if !ok {
		return nil
	}
	// remove current from list
	node.prev.next = node.next
	node.next.prev = node.prev
	return node
}

func (lc *LRUCache[T]) evict() {
	old := lc.first.next
	old.next.prev = lc.first
	lc.first.next = old.next
	delete(lc.lookup, old.key)
	log.Debug().Str("key", old.key).Msg("Evicted!")
}

func (lc *LRUCache[T]) add(node *Node[T]) {
	node.prev = lc.last.prev
	node.prev.next = node
	lc.last.prev = node
}

func (lc *LRUCache[T]) Get(key string) *T {
	curr := lc.update(key)
	if curr == nil {
		return nil
	}
	lc.add(curr)
	return &curr.value
}

func (lc *LRUCache[T]) Put(key string, value T) {
	_, ok := lc.lookup[key]
	if ok {
		node := lc.update(key)
		node.value = value
		lc.add(node)
		return
	}
	node := &Node[T]{key: key, value: value}
	if len(lc.lookup) == lc.maxSize {
		lc.evict()
	}
	lc.lookup[key] = node
	lc.add(node)
}

func (lc *LRUCache[T]) GetSize() int { return len(lc.lookup) }

func NewLRUCache[T any](maxSize int) *LRUCache[T] {
	first := Node[T]{}
	last := Node[T]{}
	first.next = &last
	last.prev = &first

	return &LRUCache[T]{
		lookup:  map[string]*Node[T]{},
		first:   &first,
		last:    &last,
		maxSize: maxSize,
	}
}
