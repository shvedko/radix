package arena

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_size28(t *testing.T) {
	var g granule
	var i int
	for {
		n := put28(&g, i)
		if n == 0 {
			break
		}
		j, h := get28(&g)
		require.Equal(t, n, h, i)
		require.Equal(t, i, j)
		i += i + 1
	}
}

func Test_class14(t *testing.T) {
	var need, fire uint
	for i := uint16(0); i <= pageGranules+1; i++ {
		class, remain, step := class14(i)
		if step == 0 {
			continue
		}
		need += uint(i)
		fire += uint(1<<class) - uint(remain*step)
	}
	if need != 134225920 || fire != 139810136 {
		t.Fatal(need, fire)
	}
	loss := float64(fire-need) / float64(fire) * 100
	if loss > 4 {
		t.Fatal(loss)
	}
}

func TestSized_want(t *testing.T) {
	type args struct {
		pid   uint64
		gid   uint16
		count uint16
	}
	tests := []struct {
		name  string
		args  args
		want  uint16
		want1 bool
	}{
		// TODO: Add test cases.
		{
			name: "",
			args: args{
				pid:   0,
				count: 1,
			},
			want:  0,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:   0,
				count: 1,
			},
			want:  1,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:   0,
				count: 1,
			},
			want:  2,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:   0,
				count: 2,
			},
			want:  4,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:   0,
				count: 4,
			},
			want:  8,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:   0,
				count: 2,
			},
			want:  6,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:   0,
				count: 1,
			},
			want:  3,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:   0,
				count: 8,
			},
			want:  16,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:   0,
				count: 16,
			},
			want:  32,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:   0,
				count: 8,
			},
			want:  24,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:   0,
				count: 32,
			},
			want:  64,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:   0,
				count: 64,
			},
			want:  128,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:   0,
				count: 128,
			},
			want:  256,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:   0,
				count: 256,
			},
			want:  512,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:   0,
				count: 512,
			},
			want:  1024,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:   0,
				count: 1024,
			},
			want:  2048,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:   0,
				count: 2048,
			},
			want:  4096,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:   0,
				count: 4096,
			},
			want:  8192,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:   0,
				count: 8192,
			},
			want:  0,
			want1: false,
		},
		{
			name: "",
			args: args{
				pid:   1,
				count: 8192,
			},
			want:  0,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:   0,
				count: pageGranules,
			},
			want:  0,
			want1: false,
		},
		{
			name: "",
			args: args{
				pid:   1,
				count: pageGranules,
			},
			want:  0,
			want1: false,
		},
		{
			name: "",
			args: args{
				pid:   2,
				count: pageGranules,
			},
			want:  0,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:   2,
				count: 1,
			},
			want:  0,
			want1: false,
		},
		{
			name: "",
			args: args{
				pid:   1,
				count: 1,
			},
			want:  8192,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:   0,
				gid:   0,
				count: 1,
			},
			want:  12,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:   0,
				gid:   11111,
				count: 1,
			},
			want:  12288,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:   0,
				gid:   12345,
				count: 1,
			},
			want:  12345,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:   0,
				gid:   pageGranules - 600, // 15784
				count: 512,
			},
			want:  15872, // + 512 = 16384, > 15784
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:   0,
				gid:   pageGranules - 1000,
				count: 512,
			},
			want:  0,
			want1: false,
		},
		{
			name: "",
			args: args{
				pid:   0,
				gid:   pageGranules - 1100,
				count: 512,
			},
			want:  15360, // > 15284
			want1: true,
		},
	}
	a := Sized{
		Linked: Linked{
			bitset0: []uint64{0},
			bitset1: []*bitset256{{}, {}, {}},
			bitset2: []*bitset16k{{}, {}, {}},
			pages:   []*page{{}, {}, {}},
			hint:    0,
		},
		hints: [16]uint64{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := a.want(tt.args.pid, tt.args.gid, tt.args.count)
			require.Equal(t, tt.want1, ok)
			require.Equal(t, tt.want, got)
			if ok {
				a.mark2(tt.args.pid, got, tt.args.count)
			}
		})
	}

	t.Run("random", func(t *testing.T) {
		b := Sized{
			Linked: Linked{
				bitset0: []uint64{0},
				bitset1: []*bitset256{{}},
				bitset2: []*bitset16k{{}},
				pages:   []*page{{}},
				hint:    0,
			},
			hints: [16]uint64{},
		}
		c := []uint8{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}
		h := [16]uint64{}
		for len(c) > 0 {
			i := random(t, len(c))
			u := c[i]
			x, r, s := class14((1<<u)/2 + 1 + uint16(random(t, 1<<u)/2))
			if x != u {
				t.Fatal(x, u)
			}
			pid, gid := unpack(h[x])
			gid, ok := b.want(pid, gid, 1<<u)
			if ok {
				b.mark2(0, gid, (1<<u)-r*s)
				gid += 1 << u
				gid -= r * s
				gid--
				h[x] = pack(pid, gid)
			} else {
				c = append(c[:i], c[i+1:]...)
			}
		}
		require.Equal(t, newBitset16k(t), b.bitset2[0])
		require.Equal(t, newBitset256(t), b.bitset1[0])
		require.Equal(t, uint64(1), b.bitset0[0])
	})
}

var seed int

func random(t *testing.T, n int) int {
	t.Helper()
	seed = (1103515245*seed + 12345) & 0x7fffffff
	return seed % n
}

func BenchmarkSized_want(b *testing.B) {
	a := Sized{
		Linked: Linked{
			bitset0: []uint64{0},
			bitset1: []*bitset256{{}, {}},
			bitset2: []*bitset16k{{}, {}},
			pages:   []*page{{}, {}},
			hint:    0,
		},
		hints: [16]uint64{},
	}
	var gid uint16
	var ok bool
	for i := 0; i < b.N; i++ {
		var pid uint64
		j := uint16(i & 15)
		switch j {
		case 15:
			a.reset()
			continue
		case 14:
			pid = 1
			gid = 0
			fallthrough
		default:
			j = 1 << j
		}
		gid, ok = a.want(pid, gid, j)
		if !ok {
			b.Fatal(j)
		}
		a.mark(pid, gid, true)
	}
}

func TestSized_write(t *testing.T) {
	type want struct {
		id    uint64
		head  *granule
		hints [16]uint64
	}
	tests := []struct {
		name string
		size int
		want want
	}{
		// TODO: Add test cases.
		{
			name: "",
			size: 0,
			want: want{id: 0, head: &granule{0x00}, hints: [16]uint64{1}},
		}, {
			name: "",
			size: 1,
			want: want{id: 1, head: &granule{0x01}, hints: [16]uint64{2}},
		}, {
			name: "",
			size: 2,
			want: want{id: 2, head: &granule{0x02}, hints: [16]uint64{3}},
		}, {
			name: "",
			size: 7,
			want: want{id: 3, head: &granule{0x07}, hints: [16]uint64{4}},
		}, {
			name: "",
			size: 8,
			want: want{id: 4, head: &granule{0x08}, hints: [16]uint64{4, 6}},
		}, {
			name: "",
			size: 9,
			want: want{id: 6, head: &granule{0x09}, hints: [16]uint64{4, 8}},
		}, {
			name: "",
			size: 17,
			want: want{id: 8, head: &granule{0x11}, hints: [16]uint64{4, 8, 11}},
		}, {
			name: "",
			size: 31,
			want: want{id: 12, head: &granule{0x1f}, hints: [16]uint64{4, 8, 16}},
		}, {
			name: "",
			size: 37,
			want: want{id: 16, head: &granule{0x25}, hints: [16]uint64{4, 8, 16, 21}},
		}, {
			name: "",
			size: 41,
			want: want{id: 24, head: &granule{0x29}, hints: [16]uint64{4, 8, 16, 30}},
		}, {
			name: "",
			size: 49,
			want: want{id: 32, head: &granule{0x31}, hints: [16]uint64{4, 8, 16, 39}},
		}, {
			name: "",
			size: 57,
			want: want{id: 40, head: &granule{0x39}, hints: [16]uint64{4, 8, 16, 48}},
		}, {
			name: "",
			size: 65,
			want: want{id: 48, head: &granule{0x41}, hints: [16]uint64{4, 8, 16, 48, 57}},
		}, {
			name: "",
			size: 75,
			want: want{id: 64, head: &granule{0x4b}, hints: [16]uint64{4, 8, 16, 48, 74}},
		}, {
			name: "",
			size: 83,
			want: want{id: 80, head: &granule{0x53}, hints: [16]uint64{4, 8, 16, 48, 91}},
		}, {
			name: "",
			size: 91,
			want: want{id: 96, head: &granule{0x5b}, hints: [16]uint64{4, 8, 16, 48, 108}},
		}, {
			name: "",
			size: 99,
			want: want{id: 112, head: &granule{0x63}, hints: [16]uint64{4, 8, 16, 48, 125}},
		}, {
			name: "",
			size: 105,
			want: want{id: 128, head: &granule{0x69}, hints: [16]uint64{4, 8, 16, 48, 142}},
		}, {
			name: "",
			size: 113,
			want: want{id: 144, head: &granule{0x71}, hints: [16]uint64{4, 8, 16, 48, 159}},
		}, {
			name: "",
			size: 123,
			want: want{id: 160, head: &granule{0x7b}, hints: [16]uint64{4, 8, 16, 48, 176}},
		}, {
			name: "",
			size: 262140,
			want: want{id: ^uint64(0), hints: [16]uint64{4, 8, 16, 48, 176}},
		},
	}
	var a Sized
	var p [1 << 20]byte
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := a.write(p[:tt.size])
			require.Equal(t, tt.want.id, got)
			require.Equal(t, tt.want.hints, a.hints)
			if got == ^uint64(0) {
				return
			}
			require.Equal(t, tt.want.head, a.granule(unpack(got)))
		})
	}
}

func BenchmarkSized_write(b *testing.B) {
	var a Sized
	var p [8192]byte

	b.SetBytes(8192)
	for i := 0; i < b.N; i++ {
		j := i % 16
		if j == 0 {
			a.reset()
		}
		a.write(p[:])
	}
}

func BenchmarkSized_read(b *testing.B) {
	var a Sized
	var p [8192]byte
	id := a.write(p[:])

	b.SetBytes(8192)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := a.open(id)
		n := c.read(p[:])
		if n != len(p) {
			b.Fatal(n)
		}
	}
}

func TestSized_open(t *testing.T) {
	var a Sized
	var p [1024]byte

	id := a.write(p[:20])
	require.Equal(t, pack(0, 0), id)
	require.Equal(t, [16]uint64{0, 0, 3}, a.hints)

	c := a.open(id)
	require.Equal(t, reader{cursor: cursor{a: &a.Linked, pid: 0, gid: 0, rem: 3, off: 1}, size: 20}, c)

	n := c.read(p[:1])
	require.Equal(t, 1, n)
	require.Equal(t, reader{cursor: cursor{a: &a.Linked, pid: 0, gid: 0, rem: 3, off: 2}, size: 19}, c)

	n = c.read(p[:2])
	require.Equal(t, 2, n)
	require.Equal(t, reader{cursor: cursor{a: &a.Linked, pid: 0, gid: 0, rem: 3, off: 4}, size: 17}, c)

	n = c.read(p[:3])
	require.Equal(t, 3, n)
	require.Equal(t, reader{cursor: cursor{a: &a.Linked, pid: 0, gid: 0, rem: 3, off: 7}, size: 14}, c)

	n = c.read(p[:3])
	require.Equal(t, 3, n)
	require.Equal(t, reader{cursor: cursor{a: &a.Linked, pid: 0, gid: 0, rem: 2, off: 2}, size: 11}, c)

	n = c.read(p[:3])
	require.Equal(t, 3, n)
	require.Equal(t, reader{cursor: cursor{a: &a.Linked, pid: 0, gid: 0, rem: 2, off: 5}, size: 8}, c)

	n = c.read(p[:3])
	require.Equal(t, 3, n)
	require.Equal(t, reader{cursor: cursor{a: &a.Linked, pid: 0, gid: 1, rem: 1, off: 0}, size: 5}, c)

	n = c.read(p[:3])
	require.Equal(t, 3, n)
	require.Equal(t, reader{cursor: cursor{a: &a.Linked, pid: 0, gid: 1, rem: 1, off: 3}, size: 2}, c)

	n = c.read(p[:3])
	require.Equal(t, 2, n)
	require.Equal(t, reader{cursor: cursor{a: &a.Linked, pid: 0, gid: 1, rem: 1, off: 5}, size: 0}, c)

	n = c.read(p[:3])
	require.Equal(t, 0, n)
	require.Equal(t, reader{cursor: cursor{a: &a.Linked, pid: 0, gid: 1, rem: 1, off: 5}, size: 0}, c)

	id = a.write(p[:])
	require.Equal(t, pack(0, 256), id)

	c = a.open(id)
	require.Equal(t, reader{cursor: cursor{a: &a.Linked, pid: 0, gid: 256, rem: 129, off: 2}, size: 1024}, c)

	n = c.read(p[:])
	require.Equal(t, 1024, n)
	require.Equal(t, reader{cursor: cursor{a: &a.Linked, pid: 0, gid: 256, rem: 1, off: 2}, size: 0}, c)

}

func TestSized_free(t *testing.T) {
	var a Sized
	var p [1024]byte
	var q [1024]byte

	for i := range p {
		p[i] = byte(i)
	}

	id := a.Write(p[:20])
	require.Equal(t, pack(0, 0), id)
	require.Equal(t, [16]uint64{0, 0, 3}, a.hints)
	require.Equal(t, &granule{0x14, 0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6}, a.granule(0, 0))

	b := a.Bytes(id)
	require.Equal(t, []byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf, 0x10, 0x11, 0x12, 0x13}, b)

	r := a.Open(id)
	require.NotZero(t, r)

	n, err := r.Read(q[:])
	require.NoError(t, err)
	require.Equal(t, 20, n)
	require.Equal(t, []byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf, 0x10, 0x11, 0x12, 0x13}, q[:n])

	n, err = r.Read(q[:])
	require.ErrorIs(t, err, io.EOF)

	a.Free(id)
	require.Equal(t, Sized{
		Linked: Linked{
			bitset0: []uint64{0},
			bitset1: []*bitset256{{}},
			bitset2: []*bitset16k{{}},
			pages:   []*page{{{0x14, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, {0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e}, {0x0f, 0x10, 0x11, 0x12, 0x13, 0x00, 0x00, 0x00}}},
			hint:    0,
		},
		hints: [16]uint64{},
	}, a)

	a.Free(id)
	require.Equal(t, Sized{
		Linked: Linked{
			bitset0: []uint64{0},
			bitset1: []*bitset256{{}},
			bitset2: []*bitset16k{{}},
			pages:   []*page{{{0x14, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}, {0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e}, {0x0f, 0x10, 0x11, 0x12, 0x13, 0x00, 0x00, 0x00}}},
			hint:    0,
		},
		hints: [16]uint64{},
	}, a)
}
