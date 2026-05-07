package arena

import (
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
	var a Sized
	var p [16384]byte

	for i := range p {
		p[i] = byte(i)
	}

	id := a.write(p[:1])
	require.Equal(t, pack(0, 0), id)
	require.Equal(t, &granule{0x01, 0x00, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, a.granule(unpack(id)))

	id = a.write(p[:2])
	require.Equal(t, pack(0, 1), id)
	require.Equal(t, &granule{0x02, 0x00, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0}, a.granule(unpack(id)))

	id = a.write(p[:4])
	require.Equal(t, pack(0, 2), id)
	require.Equal(t, &granule{0x04, 0x00, 0x1, 0x2, 0x3, 0x0, 0x0, 0x0}, a.granule(unpack(id)))

	id = a.write(p[:8192])
	require.Equal(t, pack(0, 2048), id)
	require.Equal(t, &granule{0xa0, 0x00, 0x0, 0x1, 0x2, 0x3, 0x4, 0x5}, a.granule(unpack(id)))

	id = a.write(p[:16384])
	require.Equal(t, pack(0, 4096), id)
	require.Equal(t, &granule{0xc0, 0x40, 0x0, 0x0, 0x1, 0x2, 0x3, 0x4}, a.granule(unpack(id)))

}

func BenchmarkSized_write(b *testing.B) {
	var a Sized
	var p [16384]byte

	for i := 0; i < b.N; i++ {
		j := i % 16
		if j == 0 {
			a.reset()
		}
		a.write(p[:1024])
	}
}
