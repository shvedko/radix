package radix

import "math/bits"

type bits256 [4]uint64

func (b *bits256) has(i uint8) bool {
	return b[i>>6]&(1<<(i&63)) != 0
}

func (b *bits256) set(i uint8) {
	b[i>>6] |= 1 << (i & 63)
}

func (b *bits256) pop(i uint8) {
	b[i>>6] &^= 1 << (i & 63)
}

type Radix[T any] struct {
	prefix   []byte
	index    bits256
	children [256]*Radix[T]
	next     *Radix[T]
	values   []T
}

func New[T any]() *Radix[T] { return &Radix[T]{} }

func (n *Radix[T]) matchPrefix(prefix []byte, offset int) (bool, int, bool) {
	if len(prefix) == 0 {
		return true, 0, true
	}

	if offset >= len(prefix) {
		return true, 0, true
	}
	prefix = prefix[offset:]

	i := 0
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

func (n *Radix[T]) insert(prefix []byte) *Radix[T] {
	p := n
	for len(prefix) > 0 {
		b := prefix[0]

		c := p.children[b]
		if c == nil {
			c = &Radix[T]{prefix: prefix}
			p.children[b] = c
			p.index.set(b)
			return c
		}

		size := c.commonPrefix(prefix)
		if size < len(c.prefix) {
			c.split(size)
		}

		prefix = prefix[size:]
		p = c
	}
	return p
}

func (n *Radix[T]) commonPrefix(prefix []byte) int {
	i := 0
	for i < len(prefix) && i < len(n.prefix) && prefix[i] == n.prefix[i] {
		i++
	}
	return i
}

func (n *Radix[T]) split(size int) {
	c := &Radix[T]{
		prefix:   n.prefix[size:],
		children: n.children,
		index:    n.index,
		next:     n.next,
		values:   n.values,
	}
	n.children = [256]*Radix[T]{}
	n.children[c.prefix[0]] = c
	n.index = bits256{}
	n.index.set(c.prefix[0])
	n.prefix = n.prefix[:size]
	n.values = nil
	n.next = nil
}

func (n *Radix[T]) Insert(value T, unique bool, fields ...[]byte) bool {
	if len(fields) == 0 {
		return false
	}

	p := n

	for i, field := range fields {
		p = p.insert(field)

		if i < len(fields)-1 {
			if p.next == nil {
				p.next = &Radix[T]{}
			}
			p = p.next
		}
	}

	if unique && len(p.values) > 0 {
		return false
	}

	p.values = append(p.values, value)
	return true
}

type dumper[T any] func(key []byte, prefix []byte, level int, end bool, values []T) bool

func (n *Radix[T]) Dump(yield dumper[T]) bool {
	var key [1]byte
	return n.dump(key[:0], 0, true, yield)
}

func (n *Radix[T]) dump(key []byte, level int, end bool, yield dumper[T]) bool {
	if !yield(key, n.prefix, level, end, n.values) {
		return false
	}

	key = key[:1]
	level++

	var j, k int
	var m uint64
	var l bool
	for k = 3; k >= 0; k-- {
		m = n.index[k]
		if m != 0 {
			b := 63 - bits.LeadingZeros64(m)
			j = b + k<<6
			m &^= 1 << b
			l = true
			break
		}
	}

	for h := 0; h < k; h++ {
		z := n.index[h]
		for z != 0 {
			b := bits.TrailingZeros64(z)
			i := b + h<<6
			z &^= 1 << b

			key[0] = byte(i)
			if !n.children[i].dump(key, level, false, yield) {
				return false
			}
		}
	}

	for m != 0 {
		b := bits.TrailingZeros64(m)
		i := b + k<<6
		m &^= 1 << b

		key[0] = byte(i)
		if !n.children[i].dump(key, level, false, yield) {
			return false
		}
	}

	e := n.next == nil

	if l {
		key[0] = byte(j)
		if !n.children[j].dump(key, level, e, yield) {
			return false
		}
	}

	return e || n.next.dump(key[:0], level, true, yield)
}

func (n *Radix[T]) Walk(yield dumper[T]) bool {
	return n.walk(yield)
}

func (n *Radix[T]) walk(yield dumper[T]) bool {
	type frame struct {
		n     *Radix[T]
		level int
		end   bool
		key   [1]byte
		one   uint8
	}

	var (
		p frame
		q = append(make([]frame, 0, 32), frame{n: n, end: true})
	)

	for len(q) > 0 {
		p, q = q[len(q)-1], q[:len(q)-1]

		if !yield(p.key[:p.one], p.n.prefix, p.level, p.end, p.n.values) {
			return false
		}

		p.end = true
		p.level++

		if p.n.next != nil {
			q = append(q, frame{
				n:     p.n.next,
				level: p.level,
				end:   true,
			})
			p.end = false
		}

		j := len(p.n.children)
		for j > 0 {
			j--
			if p.n.children[j] != nil {
				q = append(q, frame{
					n:     p.n.children[j],
					level: p.level,
					end:   p.end,
					key:   [1]byte{byte(j)},
					one:   1,
				})
				p.end = false
			}
		}
	}

	return true
}
