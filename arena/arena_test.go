package arena

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLinked_alloc(t *testing.T) {
	tests := []struct {
		name  string
		want0 uint64
		want1 uint16
		want3 Linked
	}{
		// TODO: Add test cases.
		{
			name:  "",
			want0: 0,
			want1: 0,
			want3: Linked{
				bitset0: []uint64{0},
				bitset1: []*bitset256{{}},
				bitset2: []*bitset16k{{}},
				pages:   []*page{{}},
			},
		},
		{
			name:  "",
			want0: 1,
			want1: 0,
			want3: Linked{
				bitset0: []uint64{0},
				bitset1: []*bitset256{{}, {}},
				bitset2: []*bitset16k{{}, {}},
				pages:   []*page{{}, {}},
			},
		},
		{
			name:  "",
			want0: 2,
			want1: 0,
			want3: Linked{
				bitset0: []uint64{0},
				bitset1: []*bitset256{{}, {}, {}},
				bitset2: []*bitset16k{{}, {}, {}},
				pages:   []*page{{}, {}, {}},
			},
		},
	}
	a := Linked{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got0, got1 := a.alloc()
			require.Equal(t, tt.want0, got0)
			require.Equal(t, tt.want1, got1)
			require.Equal(t, tt.want3, a)
		})
	}
}

func newBitset16k(t *testing.T, unset ...int) *bitset16k {
	t.Helper()
	var b bitset16k
	for i := range b {
		b[i] = ^uint64(0)
	}
	for _, i := range unset {
		b[(i >> 6)] &^= uint64(1) << (i & 63)
	}
	return &b
}

func newBitset256(t *testing.T, unset ...int) *bitset256 {
	t.Helper()
	var b bitset256
	for i := range b {
		b[i] = ^uint64(0)
	}
	for _, i := range unset {
		b[(i >> 6)] &^= uint64(1) << (i & 63)
	}
	return &b
}

func TestLinked_mark(t *testing.T) {
	type args struct {
		pid      uint64
		gid      uint16
		occupied bool
	}
	tests := []struct {
		name string
		args args
		want Linked
	}{
		// TODO: Add test cases.
		{
			name: "",
			args: args{
				pid:      0,
				gid:      0,
				occupied: false,
			},
			want: Linked{
				bitset0: []uint64{0},
				bitset1: []*bitset256{newBitset256(t, 0)},
				bitset2: []*bitset16k{newBitset16k(t, 0)},
				pages:   []*page{{}},
			},
		},
		{
			name: "",
			args: args{
				pid:      0,
				gid:      0,
				occupied: false,
			},
			want: Linked{
				bitset0: []uint64{0},
				bitset1: []*bitset256{newBitset256(t, 0)},
				bitset2: []*bitset16k{newBitset16k(t, 0)},
				pages:   []*page{{}},
			},
		},
		{
			name: "",
			args: args{
				pid:      0,
				gid:      1,
				occupied: false,
			},
			want: Linked{
				bitset0: []uint64{0},
				bitset1: []*bitset256{newBitset256(t, 0)},
				bitset2: []*bitset16k{newBitset16k(t, 0, 1)},
				pages:   []*page{{}},
			},
		},
		{
			name: "",
			args: args{
				pid:      0,
				gid:      0,
				occupied: true,
			},
			want: Linked{
				bitset0: []uint64{0},
				bitset1: []*bitset256{newBitset256(t, 0)},
				bitset2: []*bitset16k{newBitset16k(t, 1)},
				pages:   []*page{{}},
			},
		},
		{
			name: "",
			args: args{
				pid:      0,
				gid:      0,
				occupied: true,
			},
			want: Linked{
				bitset0: []uint64{0},
				bitset1: []*bitset256{newBitset256(t, 0)},
				bitset2: []*bitset16k{newBitset16k(t, 1)},
				pages:   []*page{{}},
			},
		},
		{
			name: "",
			args: args{
				pid:      0,
				gid:      1,
				occupied: true,
			},
			want: Linked{
				bitset0: []uint64{1},
				bitset1: []*bitset256{newBitset256(t)},
				bitset2: []*bitset16k{newBitset16k(t)},
				pages:   []*page{{}},
			},
		},
		{
			name: "",
			args: args{
				pid:      0,
				gid:      0,
				occupied: false,
			},
			want: Linked{
				bitset0: []uint64{0},
				bitset1: []*bitset256{newBitset256(t, 0)},
				bitset2: []*bitset16k{newBitset16k(t, 0)},
				pages:   []*page{{}},
			},
		},
		{
			name: "",
			args: args{
				pid:      0,
				gid:      0,
				occupied: true,
			},
			want: Linked{
				bitset0: []uint64{1},
				bitset1: []*bitset256{newBitset256(t)},
				bitset2: []*bitset16k{newBitset16k(t)},
				pages:   []*page{{}},
			},
		},
		{
			name: "",
			args: args{
				pid:      0,
				gid:      0,
				occupied: true,
			},
			want: Linked{
				bitset0: []uint64{1},
				bitset1: []*bitset256{newBitset256(t)},
				bitset2: []*bitset16k{newBitset16k(t)},
				pages:   []*page{{}},
			},
		},
		{
			name: "",
			args: args{
				pid:      0,
				gid:      64,
				occupied: false,
			},
			want: Linked{
				bitset0: []uint64{0},
				bitset1: []*bitset256{newBitset256(t, 1)},
				bitset2: []*bitset16k{newBitset16k(t, 64)},
				pages:   []*page{{}},
			},
		},
		{
			name: "",
			args: args{
				pid:      0,
				gid:      0,
				occupied: false,
			},
			want: Linked{
				bitset0: []uint64{0},
				bitset1: []*bitset256{newBitset256(t, 0, 1)},
				bitset2: []*bitset16k{newBitset16k(t, 0, 64)},
				pages:   []*page{{}},
			},
		},
		{
			name: "",
			args: args{
				pid:      0,
				gid:      0,
				occupied: true,
			},
			want: Linked{
				bitset0: []uint64{0},
				bitset1: []*bitset256{newBitset256(t, 1)},
				bitset2: []*bitset16k{newBitset16k(t, 64)},
				pages:   []*page{{}},
			},
		},
		{
			name: "",
			args: args{
				pid:      0,
				gid:      64,
				occupied: true,
			},
			want: Linked{
				bitset0: []uint64{1},
				bitset1: []*bitset256{newBitset256(t)},
				bitset2: []*bitset16k{newBitset16k(t)},
				pages:   []*page{{}},
			},
		},
	}
	a := Linked{
		bitset0: []uint64{1},
		bitset1: []*bitset256{newBitset256(t)},
		bitset2: []*bitset16k{newBitset16k(t)},
		pages:   []*page{{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a.mark(tt.args.pid, tt.args.gid, tt.args.occupied)
			require.Equal(t, tt.want, a)
		})
	}
}

func TestLinked_find(t *testing.T) {
	var a Linked

	pid, gid := a.alloc()
	require.EqualValues(t, 0, pid)
	require.EqualValues(t, 0, gid)

	gid, ok := a.find(pid, gid)
	require.True(t, ok)
	require.EqualValues(t, 0, gid)

	a.mark(pid, gid, ok)

	gid, ok = a.find(pid, gid)
	require.True(t, ok)
	require.EqualValues(t, 1, gid)

	gid, ok = a.find(pid, gid)
	require.True(t, ok)
	require.EqualValues(t, 1, gid)

	a.mark(pid, gid, ok)

	gid, ok = a.find(pid, gid)
	require.True(t, ok)
	require.EqualValues(t, 2, gid)

	a.mark(pid, gid, ok)

	for i := gid + 1; i < pageGranules; i++ {
		gid, ok = a.find(pid, i)
		require.True(t, ok, i)
		require.EqualValues(t, gid, i)

		a.mark(pid, gid, ok)
	}

	require.EqualValues(t, 1, a.bitset0[0])
	require.EqualValues(t, [4]uint64{
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff}, *a.bitset1[0])
	require.EqualValues(t, [256]uint64{
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff,
		0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff}, *a.bitset2[0])

	_, ok = a.find(pid, 0)
	require.False(t, ok)
	_, ok = a.find(pid, pageGranules)
	require.False(t, ok)
}

func TestLinked_next(t *testing.T) {
	var a Linked

	for i := uint64(0); i < 128*pageGranules; i++ {
		pid, gid := a.next(unpack(i))
		require.EqualValues(t, i>>14, pid, i)
		require.EqualValues(t, i&0x3FFF, gid, i)
		a.mark(pid, gid, true)
		require.EqualValues(t, i, pack(pid, gid))
	}

	require.Len(t, a.bitset0, 128/64)
	require.Len(t, a.bitset1, 128)
	require.Len(t, a.bitset2, 128)
	require.Len(t, a.pages, 128)

	_, _, ok := a.scan(0, 0)
	require.False(t, ok)

	pid, gid := a.next(0, 0)
	require.False(t, ok)
	require.EqualValues(t, 128, pid)
	require.EqualValues(t, 0, gid)

	pid, gid, ok = a.scan(0, 0)
	require.True(t, ok)
	require.EqualValues(t, 128, pid)
	require.EqualValues(t, 0, gid)
}

func BenchmarkLinked_find(b *testing.B) {

	b.Run("emptied", func(b *testing.B) {
		a := Linked{
			bitset0: []uint64{0},
			bitset1: []*bitset256{{}},
			bitset2: []*bitset16k{{}},
			pages:   []*page{{}},
		}
		for i := 0; i < b.N; i++ {
			_, ok := a.find(0, uint16(i&0x3FFF))
			if !ok {
				b.Fatal(i, ok)
			}
		}
	})

	b.Run("occupied", func(b *testing.B) {
		a := Linked{
			bitset0: []uint64{0},
			bitset1: []*bitset256{{}},
			bitset2: []*bitset16k{{}},
			pages:   []*page{{}},
		}
		i := pageGranules
		for i > 0 {
			i--
			a.mark(0, uint16(i&0x3FFF), true)
		}
		b.ResetTimer()
		for i = 0; i < b.N; i++ {
			_, ok := a.find(0, uint16(i&0x3FFF))
			if ok {
				b.Fatal(i, ok)
			}
		}
	})

}

func BenchmarkLinked_mark(b *testing.B) {
	x := false
	a := Linked{
		bitset0: []uint64{0},
		bitset1: []*bitset256{{}},
		bitset2: []*bitset16k{{}},
		pages:   []*page{{}},
	}
	for i := 0; i < b.N; i++ {
		gid := uint16(i & 0x3FFF)
		if gid == 0 {
			x = !x
		}
		a.mark(0, gid, x)
	}
}

func TestLinked_Write(t *testing.T) {

	t.Run("End", func(t *testing.T) {
		type want struct {
			i uint64
			g *granule
			b bool
		}
		tests := []struct {
			name string
			args []byte
			want want
		}{
			// TODO: Add test cases.
			{
				name: "",
				args: []byte{0},
				want: want{
					i: 0,
					g: &granule{0xf1, 0x0},
					b: true,
				},
			},
			{
				name: "",
				args: []byte{0, 1},
				want: want{
					i: 1,
					g: &granule{0xf2, 0x0, 0x1},
					b: true,
				},
			},
			{
				name: "",
				args: []byte{0, 1, 2},
				want: want{
					i: 2,
					g: &granule{0xf3, 0x0, 0x1, 0x2},
					b: true,
				},
			},
			{
				name: "",
				args: []byte{0, 1, 2, 3},
				want: want{
					i: 3,
					g: &granule{0xf4, 0x0, 0x1, 0x2, 0x3},
					b: true,
				},
			},
			{
				name: "",
				args: []byte{0, 1, 2, 3, 4},
				want: want{
					i: 4,
					g: &granule{0xf5, 0x0, 0x1, 0x2, 0x3, 0x4},
					b: true,
				},
			}, {
				name: "",
				args: []byte{0, 1, 2, 3, 4, 5},
				want: want{
					i: 5,
					g: &granule{0xf6, 0x0, 0x1, 0x2, 0x3, 0x4, 0x5},
					b: true,
				},
			},
			{
				name: "",
				args: []byte{0, 1, 2, 3, 4, 5, 6},
				want: want{
					i: 6,
					g: &granule{0xf7, 0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6},
					b: true,
				},
			},
		}
		var a Linked
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := a.Write(tt.args)
				require.Equal(t, tt.want.i, got)
				require.Equal(t, tt.want.g, a.granule(unpack(got)))
				require.Equal(t, tt.want.b, a.bit2(unpack(got)))
			})
		}
	})

	t.Run("Stream", func(t *testing.T) {
		type want struct {
			i uint64
			g []*granule
			b bool
		}
		tests := []struct {
			name string
			args []byte
			want want
		}{
			// TODO: Add test cases.
			{
				name: "",
				args: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
				want: want{
					i: 0,
					g: []*granule{
						{0x01, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
						{0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e},
						{0xf6, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}},
					b: true,
				},
			},
			{
				name: "",
				args: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21},
				want: want{
					i: 3,
					g: []*granule{
						{0x01, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
						{0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e},
						{0xf7, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15}},
					b: true,
				},
			},
			{
				name: "",
				args: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22},
				want: want{
					i: 6,
					g: []*granule{
						{0x01, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
						{0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e},
						{0x80, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15}, // T.2
						{0xf1, 0x16}},
					b: true,
				},
			},
			{
				name: "",
				args: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
				want: want{
					i: 10,
					g: []*granule{
						{0x01, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
						{0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e},
						{0xf1, 0xf}},
					b: true,
				},
			},
			{
				name: "",
				args: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14},
				want: want{
					i: 13,
					g: []*granule{
						{0x80, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
						{0x80, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d},
						{0xf1, 0x0e}},
					b: true,
				},
			},
			{
				name: "",
				args: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13},
				want: want{
					i: 16,
					g: []*granule{
						{0x80, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
						{0xf7, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d}},
					b: true,
				},
			},
		}
		var a Linked
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := a.Write(tt.args)
				require.Equal(t, tt.want.i, got)
				for i := range tt.want.g {
					require.Equal(t, tt.want.g[i], a.granule(unpack(got)), i)
					require.Equal(t, tt.want.b, a.bit2(unpack(got)), i)
					got++
				}
			})
		}
	})
}
