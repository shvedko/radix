package radix

type Radix[T any] struct {
	prefix   []byte
	children [256]*Radix[T]
	next     *Radix[T]
	values   []T
}

func (n *Radix[T]) matchPrefix(bytes []byte, offset int) (bool, int, bool) {
	if len(bytes) == 0 {
		return true, 0, true
	}

	if offset >= len(bytes) {
		// Мы уже нашли всё слово в родителях, этот узел — часть "хвоста" за пределами поиска
		return true, 0, true
	}
	prefix := bytes[offset:]

	i := 0
	for i < len(n.prefix) && i < len(prefix) {
		if n.prefix[i] != prefix[i] {
			return false, 0, false // Символы разошлись — узел не подходит
		}
		i++
	}

	if i == len(prefix) {
		return true, i, true
	}

	return true, i, false
}

func (n *Radix[T]) insert(bytes []byte) *Radix[T] {
	p := n
	for len(bytes) > 0 {
		first := bytes[0]

		c := p.children[first]
		if c == nil {
			next := &Radix[T]{prefix: bytes}
			p.children[first] = next
			return next
		}

		size := c.commonPrefix(bytes)
		if size < len(c.prefix) {
			//c.split(size)
			c = c.split2(size)
			p.children[first] = c
		}

		bytes = bytes[size:]
		p = c
	}
	return p
}

func (n *Radix[T]) commonPrefix(bytes []byte) int {
	i := 0
	for i < len(bytes) && i < len(n.prefix) && bytes[i] == n.prefix[i] {
		i++
	}
	return i
}

func (n *Radix[T]) split(size int) {
	child := &Radix[T]{
		prefix:   n.prefix[size:],
		children: n.children,
		next:     n.next,
		values:   n.values,
	}

	n.prefix = n.prefix[:size]
	n.children = [256]*Radix[T]{}
	n.children[child.prefix[0]] = child
	n.values = nil
	n.next = nil
}

func (n *Radix[T]) split2(size int) *Radix[T] {
	child := &Radix[T]{
		prefix:   n.prefix[size:],
		children: n.children,
		next:     n.next,
		values:   n.values,
	}

	var children [256]*Radix[T]
	children[child.prefix[0]] = child

	return &Radix[T]{
		prefix:   n.prefix[:size],
		children: children,
		values:   nil,
		next:     nil,
	}
}

func New[T any]() *Radix[T] { return &Radix[T]{} }

func (n *Radix[T]) Insert(item T, unique bool, fields ...[]byte) bool {
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
		return false // Ключ уже занят
	}

	p.values = append(p.values, item)
	return true
}

func (n *Radix[T]) Dump(f func(key []byte, prefix []byte, level int, end bool, values []T) bool) {
	n.dump(nil, 0, true, f)
}

func (n *Radix[T]) dump(key []byte, level int, end bool, f func(key []byte, prefix []byte, level int, end bool, values []T) bool) {
	if n == nil {
		return
	}

	f(key, n.prefix, level, end, n.values)

	j := len(n.children)
	for j > 0 {
		j--
		if n.children[j] != nil {
			break
		}
	}

	for i := 0; i < j; i++ {
		if n.children[i] != nil {
			n.children[i].dump([]byte{byte(i)}, level+1, false, f)
		}
	}

	if n.children[j] != nil {
		n.children[j].dump([]byte{byte(j)}, level+1, n.next == nil, f)
	}

	if n.next != nil {
		n.next.dump(nil, level+1, true, f)
	}
}
