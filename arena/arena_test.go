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
