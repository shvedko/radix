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
