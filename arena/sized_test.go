package arena

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_size28(t *testing.T) {
	var g granule
	var i uint32
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
	for i := uint32(0); i <= pageGranules+1; i++ {
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

func Benchmark_class14(b *testing.B) {
	for i := 0; i < b.N; i++ {
		class14(uint32(i) & 0x20000)
	}
}

func TestSized_want(t *testing.T) {
	type args struct {
		pid  uint64
		gid  uint16
		size int
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
				pid:  0,
				size: 1,
			},
			want:  0,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:  0,
				size: 1,
			},
			want:  1,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:  0,
				size: 1,
			},
			want:  2,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:  0,
				size: 2,
			},
			want:  4,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:  0,
				size: 4,
			},
			want:  8,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:  0,
				size: 2,
			},
			want:  6,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:  0,
				size: 1,
			},
			want:  3,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:  0,
				size: 8,
			},
			want:  16,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:  0,
				size: 16,
			},
			want:  32,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:  0,
				size: 8,
			},
			want:  24,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:  0,
				size: 32,
			},
			want:  64,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:  0,
				size: 64,
			},
			want:  128,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:  0,
				size: 128,
			},
			want:  256,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:  0,
				size: 256,
			},
			want:  512,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:  0,
				size: 512,
			},
			want:  1024,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:  0,
				size: 1024,
			},
			want:  2048,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:  0,
				size: 2048,
			},
			want:  4096,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:  0,
				size: 4096,
			},
			want:  8192,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:  0,
				size: 8192,
			},
			want:  0,
			want1: false,
		},
		{
			name: "",
			args: args{
				pid:  1,
				size: 8192,
			},
			want:  0,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:  0,
				size: pageGranules,
			},
			want:  0,
			want1: false,
		},
		{
			name: "",
			args: args{
				pid:  1,
				size: pageGranules,
			},
			want:  0,
			want1: false,
		},
		{
			name: "",
			args: args{
				pid:  2,
				size: pageGranules,
			},
			want:  0,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:  2,
				size: 1,
			},
			want:  0,
			want1: false,
		},
		{
			name: "",
			args: args{
				pid:  1,
				size: 1,
			},
			want:  8192,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:  0,
				gid:  0,
				size: 1,
			},
			want:  12,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:  0,
				gid:  11111,
				size: 1,
			},
			want:  12288,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:  0,
				gid:  12345,
				size: 1,
			},
			want:  12345,
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:  0,
				gid:  pageGranules - 600, // 15784
				size: 512,
			},
			want:  15872, // + 512 = 16384, > 15784
			want1: true,
		},
		{
			name: "",
			args: args{
				pid:  0,
				gid:  pageGranules - 1000,
				size: 512,
			},
			want:  0,
			want1: false,
		},
		{
			name: "",
			args: args{
				pid:  0,
				gid:  pageGranules - 1100,
				size: 512,
			},
			want:  15360, // > 15284
			want1: true,
		},
	}
	a := &Sized{
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
			got, ok := a.want(tt.args.pid, tt.args.gid, tt.args.size)
			require.Equal(t, tt.want1, ok)
			require.Equal(t, tt.want, got)
			if ok {
				a.mark2(tt.args.pid, got, tt.args.size)
			}
		})
	}
}

func BenchmarkSized_want(b *testing.B) {
	a := &Sized{
		Linked: Linked{
			bitset0: []uint64{0},
			bitset1: []*bitset256{{}, {}},
			bitset2: []*bitset16k{{}, {}},
			pages:   []*page{{}, {}},
			hint:    0,
		},
		hints: [16]uint64{},
	}
	for i := 0; i < b.N; i++ {
		var pid uint64
		j := i & 15
		switch j {
		case 15:
			a.reset()
			continue
		case 14:
			pid = 1
			fallthrough
		default:
			j = 1 << j
		}
		gid, ok := a.want(pid, 0, j)
		if !ok {
			b.Fatal(j)
		}
		a.mark(pid, gid, true)
	}
}
