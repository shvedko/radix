package radix

type Radix[T any] struct {
	prefix   []byte
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
		next:     n.next,
		values:   n.values,
	}
	n.children = [256]*Radix[T]{}
	n.children[c.prefix[0]] = c
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

func (n *Radix[T]) Dump(yield dumper[T]) {
	n.dump(nil, 0, true, yield)
}

func (n *Radix[T]) dump(key []byte, level int, end bool, yield dumper[T]) {
	if n == nil {
		return
	}

	yield(key, n.prefix, level, end, n.values)

	level++

	j := len(n.children)
	for j > 0 {
		j--
		if n.children[j] != nil {
			break
		}
	}

	for i := 0; i < j; i++ {
		n.children[i].dump([]byte{byte(i)}, level, false, yield)
	}

	n.children[j].dump([]byte{byte(j)}, level, n.next == nil, yield)

	n.next.dump(nil, level, true, yield)
}

func (n *Radix[T]) Walk(yield dumper[T]) {
	n.walk(yield)
}

func (n *Radix[T]) walk(yield dumper[T]) {
	if n == nil {
		return
	}

	type frame struct {
		n     *Radix[T]
		level int
		end   bool
		key   [1]byte
		one   uint8
	}

	var (
		p frame
		q = []frame{{n: n, end: true}}
	)

	for len(q) > 0 {
		p, q = q[len(q)-1], q[:len(q)-1]

		yield(p.key[:p.one], p.n.prefix, p.level, p.end, p.n.values)

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
}
