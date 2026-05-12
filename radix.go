package radix

import (
	"math/bits"
	"unsafe"
)

type bitset256 [4]uint64

func (b *bitset256) has(i uint8) bool { return b[i>>6]&(1<<(i&63)) != 0 }
func (b *bitset256) set(i uint8)      { b[i>>6] |= 1 << (i & 63) }
func (b *bitset256) pop(i uint8)      { b[i>>6] &^= 1 << (i & 63) }
func (b *bitset256) num(i uint8) int {
	k := i >> 6
	i &= 63
	m := uint64(1)<<i - 1
	n := bits.OnesCount64(b[k] & m)
	for k > 0 {
		k--
		n += bits.OnesCount64(b[k])
	}
	return n
}

const pageSize = 64

type pool[T any] struct {
	pages [][]Radix[T]
	nodes []*Radix[T]
}

func (p *pool[T]) grow() {
	if cap(p.nodes) < pageSize {
		p.nodes = make([]*Radix[T], 0, pageSize)
	}

	page := make([]Radix[T], pageSize)
	p.pages = append(p.pages, page)

	for i := range page {
		p.nodes = append(p.nodes, &page[i])
	}
}

func (p *pool[T]) get() *Radix[T] {
	if len(p.nodes) == 0 {
		p.grow()
	}

	n := p.nodes[len(p.nodes)-1]
	p.nodes = p.nodes[:len(p.nodes)-1]
	n.pool = p

	return n
}

func (p *pool[T]) put(n *Radix[T]) {
	n.prefix = nil
	n.index = bitset256{}
	if cap(n.children) > 64 {
		n.children = nil
	} else {
		n.children = n.children[:0]
	}
	n.values = n.values[:0]
	n.next = nil
	n.pool = p

	p.nodes = append(p.nodes, n)
}

func (p *pool[T]) reset() {
	p.nodes = p.nodes[:0]

	for i := range p.pages {
		for j := range p.pages[i] {
			n := &p.pages[i][j]
			n.zero()
			p.put(n)
		}
	}
}

type Radix[T any] struct {
	prefix   []byte
	index    bitset256
	children []*Radix[T]
	next     *Radix[T]
	values   []T
	pool     *pool[T]
}

func New[T any]() *Radix[T] { return &Radix[T]{pool: &pool[T]{}} }

func (n *Radix[T]) Grow(nodes int) *Radix[T] { // FIXME
	for len(n.pool.pages)*pageSize < nodes {
		n.pool.grow()
	}
	return n
}

func (n *Radix[T]) match(prefix []byte, offset uint32) (bool, int, bool) {
	if uint32(len(prefix)) <= offset {
		return true, 0, true
	}

	i := 0
	prefix = prefix[offset:]
	for i < len(n.prefix) && i < len(prefix) {
		if n.prefix[i] != prefix[i] {
			return false, 0, false
		}
		i++
	}

	if i == len(prefix) {
		return true, i, true
	}

	return true, i, false
}

func (n *Radix[T]) insert(prefix []byte, frames []frame[T], layer uint16, mode uint8) ([]frame[T], *Radix[T]) {
	p := n

	var offset uint32
	for len(prefix) > 0 {
		var mutate uint8

		b := prefix[0]
		i := p.index.num(b)
		if !p.index.has(b) {
			p.index.set(b)
			p.children = append(p.children, nil)
			copy(p.children[i+1:], p.children[i:])
			c := n.pool.get()
			p.children[i] = c
			p.children[i].prefix = prefix
			mutate = mode << 4
		}

		p = p.children[i]
		size := p.common(prefix)
		if size < len(p.prefix) {
			p.split(size)
			mutate |= mode << 2
		} else {
			mutate |= mode
		}

		offset += uint32(size)
		prefix = prefix[size:]

		if frames != nil {
			frames = append(frames, frame[T]{
				n:      p,
				layer:  layer,
				offset: offset,
				mode:   mode + (mutate&8>>3|mutate&4>>1|mutate&1<<1)>>(mutate&48>>3),
			})
		}
	}

	return frames, p
}

func (n *Radix[T]) common(prefix []byte) int {
	i := 0
	for i < len(prefix) && i < len(n.prefix) && prefix[i] == n.prefix[i] {
		i++
	}
	return i
}

func (n *Radix[T]) split(size int) {
	c := n.pool.get()
	c.prefix = n.prefix[size:]
	c.index = n.index
	c.next = n.next

	c.values, n.values = n.values, c.values
	c.children, n.children = n.children, append(c.children[:0], c)

	n.prefix = n.prefix[:size]
	n.index = bitset256{}
	n.index.set(c.prefix[0])
	n.next = nil
}

func (n *Radix[T]) Insert(value T, unique bool, prefixes ...[]byte) bool {
	if len(prefixes) == 0 {
		return false
	}

	p := n

	var i uint16
	for i < uint16(len(prefixes)-1) {
		_, p = p.insert(prefixes[i], nil, i, 2)
		if p.next == nil {
			p.next = p.pool.get()
		}
		p = p.next
		i++
	}
	_, p = p.insert(prefixes[i], nil, i, 1)

	if unique && len(p.values) > 0 {
		return false
	}

	p.values = append(p.values, value)

	return true
}

func (n *Radix[T]) InsertPath(value T, unique bool, prefixes ...[]byte) (Iterator[T], bool) {
	if len(prefixes) == 0 {
		return Iterator[T]{}, false
	}

	p := n

	frames := (&[8]frame[T]{{n: n, mode: 3}})[:1]

	var i uint16
	for i < uint16(len(prefixes)-1) {
		frames, p = p.insert(prefixes[i], frames, i, 2)
		if p.next == nil {
			p.next = n.pool.get()
		}
		p = p.next
		i++
		frames = append(frames, frame[T]{n: p, layer: i, mode: 3})
	}
	frames, p = p.insert(prefixes[i], frames, i, 1)

	if unique && len(p.values) > 0 {
		return Iterator[T]{}, false
	}

	p.values = append(p.values, value)

	return Iterator[T]{frames: frames, prefixes: prefixes}, true
}

type dumper[T any] func(prefix []byte, level uint32, end bool, values []T) bool

func (n *Radix[T]) Dump(yield dumper[T]) bool {
	return n.dump(0, true, yield)
}

func (n *Radix[T]) dump(level uint32, end bool, yield dumper[T]) bool {
	if !yield(n.prefix, level, end, n.values) {
		return false
	}

	level++
	end = n.next == nil

	if len(n.children) > 0 {
		i := 0
		for i < len(n.children)-1 {
			if !n.children[i].dump(level, false, yield) {
				return false
			}
			i++
		}
		if !n.children[i].dump(level, end, yield) {
			return false
		}
	}

	return end || n.next.dump(level, true, yield)
}

func (n *Radix[T]) Walk(yield dumper[T]) bool {
	return n.walk(yield)
}

type step[T any] struct {
	n     *Radix[T]
	level uint32
	end   bool
}

func (n *Radix[T]) walk(yield dumper[T]) bool {
	var (
		p step[T]
		q = append(make([]step[T], 0, 32), step[T]{n: n, end: true})
	)

	for len(q) > 0 {
		p, q = q[len(q)-1], q[:len(q)-1]

		if !yield(p.n.prefix, p.level, p.end, p.n.values) {
			return false
		}

		p.end = true
		p.level++

		if p.n.next != nil {
			q = append(q, step[T]{
				n:     p.n.next,
				level: p.level,
				end:   p.end,
			})
			p.end = false
		}

		i := len(p.n.children)
		if i > 0 {
			i--
			q = append(q, step[T]{
				n:     p.n.children[i],
				level: p.level,
				end:   p.end,
			})
			for i > 0 {
				i--
				q = append(q, step[T]{
					n:     p.n.children[i],
					level: p.level,
					end:   false,
				})
			}
		}
	}

	return true
}

func (n *Radix[T]) Search(prefixes ...[]byte) Iterator[T] {
	return Iterator[T]{
		prefixes: prefixes,
		frames:   (&[8]frame[T]{{n: n}})[:1],
	}
}

func (n *Radix[T]) Foreach(prefixes ...[]byte) func(func(T) bool) {
	i := n.Search(prefixes...)
	return func(yield func(T) bool) {
		for i.Next() {
			for _, v := range i.Get() {
				if !yield(v) {
					return
				}
			}
		}
	}
}

type frame[T any] struct {
	n      *Radix[T]
	offset uint32
	layer  uint16
	mode   uint8
	c      uint8
}

type Iterator[T any] struct {
	frames   []frame[T]
	prefixes [][]byte
}

func (t *Iterator[T]) Next() bool {
	for len(t.frames) > 0 {
		f := &t.frames[len(t.frames)-1]

		var prefix []byte
		if f.layer < uint16(len(t.prefixes)) {
			prefix = t.prefixes[f.layer]
		}

		matched, consumed, end := f.n.match(prefix, f.offset)
		if matched {
			f.offset += uint32(consumed)
			if end {
				switch f.mode {
				case 0:
					f.mode++
					if len(f.n.values) > 0 && f.layer+1 >= uint16(len(t.prefixes)) {
						return true
					}
					fallthrough

				case 1:
					f.mode++
					if f.n.next != nil {
						t.frames = append(t.frames, frame[T]{
							n:     f.n.next,
							layer: f.layer + 1,
						})
						continue
					}
					fallthrough

				case 2:
					m := ^uint64(0) << (f.c & 63)
					switch f.c >> 6 {
					case 0:
						m &= f.n.index[0]
						if m != 0 {
							t.append(0<<6, m, f)
							continue
						}
						m = ^m
						fallthrough
					case 1:
						m &= f.n.index[1]
						if m != 0 {
							t.append(1<<6, m, f)
							continue
						}
						m = ^m
						fallthrough
					case 2:
						m &= f.n.index[2]
						if m != 0 {
							t.append(2<<6, m, f)
							continue
						}
						m = ^m
						fallthrough
					case 3:
						m &= f.n.index[3]
						if m != 0 {
							t.append(3<<6, m, f)
							continue
						}
						fallthrough
					default:
						f.mode++
					}
					fallthrough

				case 3:
				}
			}

			if f.mode != 3 {
				f.mode = 3
				c := prefix[f.offset]
				if f.n.index.has(c) {
					t.frames = append(t.frames, frame[T]{
						n:      f.n.children[f.n.index.num(c)],
						offset: f.offset,
						layer:  f.layer,
					})
					continue
				}
			}
		}

		t.frames = t.frames[:len(t.frames)-1]
	}

	return false
}

func (t *Iterator[T]) append(i uint8, m uint64, f *frame[T]) {
	i += uint8(bits.TrailingZeros64(m))
	f.c = i + 1
	if f.c == 0 {
		f.mode++
	}
	t.frames = append(t.frames, frame[T]{
		n:      f.n.children[f.n.index.num(i)],
		offset: f.offset,
		layer:  f.layer,
	})
}

func (t Iterator[T]) Get() []T {
	if len(t.frames) == 0 {
		return nil
	}
	return t.frames[len(t.frames)-1].n.values
}

func (t Iterator[T]) Up() Iterator[T] {
	if len(t.frames) > 0 {
		i := len(t.frames) - 1
		f := t.frames[i]
		for i > 0 {
			i--
			p := t.frames[i]
			if i > 0 {
				if len(p.n.children) > 0 && p.n.children[0] != f.n && len(p.n.children[0].values) > 0 {
					p.mode = 2
					p.c = p.n.children[0].prefix[0] + 1
					f.n = p.n.children[0]
					f.offset = p.offset + uint32(len(f.n.prefix))
					t.frames = append(t.frames[:i:i], p, f)
					i++
					i++
					break
				} else if len(p.n.values) > 0 {
					i++
					break
				}
			}
			f = p
		}
		t.frames = t.frames[:i]
	}
	return t
}

func (t Iterator[T]) Down() Iterator[T] {
	if len(t.frames) > 0 {
		i := len(t.frames) - 1
		f := &t.frames[i]
		if f.n.next == nil && len(f.n.children) == 0 {
			frames := make([]frame[T], i, cap(t.frames))
			for j := i - 1; j >= 0; j-- {
				frames[j] = t.frames[j]
				if j > 0 && len(f.n.prefix) > 0 {
					frames[j].c = f.n.prefix[0] + 1
					frames[j].mode = 2
				}
				f = &frames[j]
			}
			t.frames = frames
			t.prefixes = nil
		}
		t.Next()
	}
	return t
}

func (t *Iterator[T]) Commit(prefixes [][]byte) {
	for i := range t.frames {
		f := &t.frames[i]
		n := f.n
		if uint16(len(prefixes)) == f.layer {
			break
		}
		if len(n.prefix) == 0 || uint32(len(prefixes[f.layer])) < f.offset {
			continue
		}
		start := f.offset - uint32(len(n.prefix))
		if &n.prefix[0] == &t.prefixes[f.layer][start] && prefixes[f.layer][start] == t.prefixes[f.layer][start] {
			n.prefix = prefixes[f.layer][start:f.offset]
		}
	}
	t.prefixes = prefixes
}

func (t *Iterator[T]) Prefix(prefixes [][]byte) {
	t.prefixes = prefixes
}

func (t *Iterator[T]) Remove() {
	if len(t.frames) == 0 {
		return
	}
	t.merge(-2)
}

func (t *Iterator[T]) Delete(v int) {
	if v < 0 || len(t.frames) == 0 {
		return
	}
	t.merge(v)
}

func (t *Iterator[T]) Rollback() {
	if len(t.frames) == 0 {
		return
	}
	t.merge(-1)
}

func (t *Iterator[T]) merge(v int) {
	i := len(t.frames) - 1
	f := &t.frames[i]

	switch v {
	case -2:
		f.n.zero()
		f.n.values = f.n.values[:0]
	case -1:
		v += len(f.n.values)
		fallthrough
	default:
		if v >= len(f.n.values) {
			return
		}
		v += copy(f.n.values[v:], f.n.values[v+1:])
		clear(f.n.values[v:])
		f.n.values = f.n.values[:v]
	}

	for f.n.empty() && i > 0 {
		p := &t.frames[i-1]
		p.mode = 1
		if p.n.next == f.n {
			p.n.next = nil
		} else {
			p.n.remove(f.n.prefix[0])
		}
		t.frames = t.frames[:i]
		i--
		f = &t.frames[i]
		f.n.merge()
	}
}

func (n *Radix[T]) merge() {
	if n.prefix != nil && len(n.values) == 0 && n.next == nil && len(n.children) == 1 {
		c := n.children[0]
		n.prefix = unsafe.Slice((*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(&c.prefix[0]))-uintptr(len(n.prefix)))), len(n.prefix)+len(c.prefix))
		n.index = c.index
		n.children, c.children = c.children, n.children
		n.values, c.values = c.values, n.values
		n.next = c.next

		n.pool.put(c)
	}
}

func (n *Radix[T]) empty() bool {
	return len(n.values) == 0 && n.next == nil && len(n.children) == 0
}

func (n *Radix[T]) remove(c uint8) {
	i := n.index.num(c)
	n.index.pop(c)
	n.pool.put(n.children[i])
	i += copy(n.children[i:], n.children[i+1:])
	n.children = n.children[:i]
}

func (n *Radix[T]) Reset() {
	n.prefix = nil
	n.children = n.children[:0]
	n.index = bitset256{}
	n.zero()
	n.values = n.values[:0]
	n.next = nil
	n.pool.reset()
}

func (n *Radix[T]) zero() {
	clear(n.values)
}

func (n *Radix[T]) Empty() bool {
	return n.empty()
}

func (n *Radix[T]) siblings(c *Radix[T]) (senior, junior *Radix[T]) {
	i := n.index.num(c.prefix[0])
	if i > 0 {
		senior = n.children[i-1]
	}
	i++
	if i < len(n.children) {
		junior = n.children[i]
	}
	return
}
