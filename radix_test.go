package radix_test

import (
	"fmt"
	"radix"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func newDumper[T any](p func(a ...any)) func(prefix []byte, level int, end bool, values []T) bool {
	m := make(map[int]uint8)
	v := [2]rune{'│', ' '}
	r := [2]rune{'├', '└'}
	return func(prefix []byte, l int, e bool, values []T) bool {
		var u uint8
		if e {
			u = 1
		}
		m[l] = u
		var b strings.Builder
		if len(m) > 1 {
			for i := 1; i < l; i++ {
				b.WriteRune(v[m[i]])
				b.WriteString("  ")
			}
			b.WriteRune(r[m[l]])
			b.WriteRune('─')
		}
		if len(prefix) > 0 {
			b.WriteString("─[")
			b.WriteByte(prefix[0])
			b.WriteString("]:\"")
			b.Write(prefix)
			b.WriteByte('"')
		} else if len(m) > 1 {
			b.WriteString("»┬")
		} else {
			b.WriteRune('┬')
		}
		if len(values) > 0 {
			_, _ = fmt.Fprintf(&b, " = %v", values)
		}
		p(&b)
		return true
	}
}

func ExampleRadix_Dump() {
	t := radix.New[int]()

	t.Insert(1, false, []byte("Pavlov"), []byte("Ivan"))
	t.Insert(2, false, []byte("Pavlov"), []byte("Igor"))
	t.Insert(3, false, []byte("Petrov"), []byte("Ivan"))
	t.Insert(4, false, []byte("Vanina"), []byte("Zina"))
	t.Insert(5, false, []byte("Vanina"), []byte("Zina"))
	t.Insert(6, true, []byte("Vanina"), []byte("Zina"))
	t.Insert(7, false, []byte("Pavlov"), []byte("Oleg"))
	t.Insert(8, false, []byte("Pavlova"), []byte("Zinaida"))
	t.Insert(9, false, []byte("Pushkin"), []byte("Alexander"))
	t.Insert(0, false, []byte("Petras"), []byte("Alex"))
	t.Insert(10, false, []byte("P"), []byte("I"))

	t.Dump(newDumper[int](func(a ...any) { fmt.Println(a...) }))

	fmt.Println("===[]:")
	i := t.Search()
	for i.Next() {
		fmt.Println(i.Get())
	}

	fmt.Println("===[P]:")
	i = t.Search([]byte("P"))
	for i.Next() {
		fmt.Println(i.Get())
	}

	fmt.Println("===[T]:")
	i = t.Search([]byte("T"))
	for i.Next() {
		fmt.Println(i.Get())
	}

	fmt.Println("===[Pavlova]:")
	i = t.Search([]byte("Pavlova"))
	for i.Next() {
		fmt.Println(i.Get())
	}

	fmt.Println("===[Pavlov]:")
	i = t.Search([]byte("Pavlov"))
	for i.Next() {
		fmt.Println(i.Get())
	}

	fmt.Println("===[P, ]:")
	i = t.Search([]byte("P"), nil)
	for i.Next() {
		fmt.Println(i.Get())
	}

	fmt.Println("===[P, , ]:")
	i = t.Search([]byte("P"), nil, nil)
	for i.Next() {
		fmt.Println(i.Get())
	}

	fmt.Println("===[P, X]:")
	i = t.Search([]byte("P"), []byte("X"))
	for i.Next() {
		fmt.Println(i.Get())
	}

	fmt.Println("===[, I]:")
	i = t.Search(nil, []byte("I"))
	for i.Next() {
		fmt.Println(i.Get())
	}

	fmt.Println("===[, Zina]:")
	i = t.Search(nil, []byte("Zina"))
	for i.Next() {
		fmt.Println(i.Get())
	}

	fmt.Println("===[, ]:")
	i = t.Search(nil, nil)
	for i.Next() {
		fmt.Println(i.Get())
	}

	fmt.Println("===[, Alexey]:")
	i = t.Search(nil, []byte("Alexey"))
	for i.Next() {
		fmt.Println(i.Get())
	}

	// Output:
	// ┬
	// ├──[P]:"P"
	// │  ├──[a]:"avlov"
	// │  │  ├──[a]:"a"
	// │  │  │  └─»┬
	// │  │  │     └──[Z]:"Zinaida" = [8]
	// │  │  └─»┬
	// │  │     ├──[I]:"I"
	// │  │     │  ├──[g]:"gor" = [2]
	// │  │     │  └──[v]:"van" = [1]
	// │  │     └──[O]:"Oleg" = [7]
	// │  ├──[e]:"etr"
	// │  │  ├──[a]:"as"
	// │  │  │  └─»┬
	// │  │  │     └──[A]:"Alex" = [0]
	// │  │  └──[o]:"ov"
	// │  │     └─»┬
	// │  │        └──[I]:"Ivan" = [3]
	// │  ├──[u]:"ushkin"
	// │  │  └─»┬
	// │  │     └──[A]:"Alexander" = [9]
	// │  └─»┬
	// │     └──[I]:"I" = [10]
	// └──[V]:"Vanina"
	//    └─»┬
	//       └──[Z]:"Zina" = [4 5]
	// ===[]:
	// [10]
	// [2]
	// [1]
	// [7]
	// [8]
	// [0]
	// [3]
	// [9]
	// [4 5]
	// ===[P]:
	// [10]
	// [2]
	// [1]
	// [7]
	// [8]
	// [0]
	// [3]
	// [9]
	// ===[T]:
	// ===[Pavlova]:
	// [8]
	// ===[Pavlov]:
	// [2]
	// [1]
	// [7]
	// [8]
	// ===[P, ]:
	// [10]
	// [2]
	// [1]
	// [7]
	// [8]
	// [0]
	// [3]
	// [9]
	// ===[P, , ]:
	// ===[P, X]:
	// ===[, I]:
	// [10]
	// [2]
	// [1]
	// [3]
	// ===[, Zina]:
	// [8]
	// [4 5]
	//===[, ]:
	// [10]
	// [2]
	// [1]
	// [7]
	// [8]
	// [0]
	// [3]
	// [9]
	// [4 5]
	// ===[, Alexey]:
	//
}

func ExampleRadix_Walk() {
	t := radix.New[int]()

	for i := 0; i < 100; i++ {
		s := strconv.Itoa(i)
		c := strconv.Itoa(i % 8)
		t.Insert(i, false, []byte("City"+c), []byte("Street"+s))
	}

	t.Walk(newDumper[int](func(a ...any) { fmt.Println(a...) }))

	// Output:
	// ┬
	// └──[C]:"City"
	//    ├──[0]:"0"
	//    │  └─»┬
	//    │     └──[S]:"Street"
	//    │        ├──[0]:"0" = [0]
	//    │        ├──[1]:"16" = [16]
	//    │        ├──[2]:"24" = [24]
	//    │        ├──[3]:"32" = [32]
	//    │        ├──[4]:"4"
	//    │        │  ├──[0]:"0" = [40]
	//    │        │  └──[8]:"8" = [48]
	//    │        ├──[5]:"56" = [56]
	//    │        ├──[6]:"64" = [64]
	//    │        ├──[7]:"72" = [72]
	//    │        ├──[8]:"8" = [8]
	//    │        │  ├──[0]:"0" = [80]
	//    │        │  └──[8]:"8" = [88]
	//    │        └──[9]:"96" = [96]
	//    ├──[1]:"1"
	//    │  └─»┬
	//    │     └──[S]:"Street"
	//    │        ├──[1]:"1" = [1]
	//    │        │  └──[7]:"7" = [17]
	//    │        ├──[2]:"25" = [25]
	//    │        ├──[3]:"33" = [33]
	//    │        ├──[4]:"4"
	//    │        │  ├──[1]:"1" = [41]
	//    │        │  └──[9]:"9" = [49]
	//    │        ├──[5]:"57" = [57]
	//    │        ├──[6]:"65" = [65]
	//    │        ├──[7]:"73" = [73]
	//    │        ├──[8]:"8"
	//    │        │  ├──[1]:"1" = [81]
	//    │        │  └──[9]:"9" = [89]
	//    │        └──[9]:"9" = [9]
	//    │           └──[7]:"7" = [97]
	//    ├──[2]:"2"
	//    │  └─»┬
	//    │     └──[S]:"Street"
	//    │        ├──[1]:"1"
	//    │        │  ├──[0]:"0" = [10]
	//    │        │  └──[8]:"8" = [18]
	//    │        ├──[2]:"2" = [2]
	//    │        │  └──[6]:"6" = [26]
	//    │        ├──[3]:"34" = [34]
	//    │        ├──[4]:"42" = [42]
	//    │        ├──[5]:"5"
	//    │        │  ├──[0]:"0" = [50]
	//    │        │  └──[8]:"8" = [58]
	//    │        ├──[6]:"66" = [66]
	//    │        ├──[7]:"74" = [74]
	//    │        ├──[8]:"82" = [82]
	//    │        └──[9]:"9"
	//    │           ├──[0]:"0" = [90]
	//    │           └──[8]:"8" = [98]
	//    ├──[3]:"3"
	//    │  └─»┬
	//    │     └──[S]:"Street"
	//    │        ├──[1]:"1"
	//    │        │  ├──[1]:"1" = [11]
	//    │        │  └──[9]:"9" = [19]
	//    │        ├──[2]:"27" = [27]
	//    │        ├──[3]:"3" = [3]
	//    │        │  └──[5]:"5" = [35]
	//    │        ├──[4]:"43" = [43]
	//    │        ├──[5]:"5"
	//    │        │  ├──[1]:"1" = [51]
	//    │        │  └──[9]:"9" = [59]
	//    │        ├──[6]:"67" = [67]
	//    │        ├──[7]:"75" = [75]
	//    │        ├──[8]:"83" = [83]
	//    │        └──[9]:"9"
	//    │           ├──[1]:"1" = [91]
	//    │           └──[9]:"9" = [99]
	//    ├──[4]:"4"
	//    │  └─»┬
	//    │     └──[S]:"Street"
	//    │        ├──[1]:"12" = [12]
	//    │        ├──[2]:"2"
	//    │        │  ├──[0]:"0" = [20]
	//    │        │  └──[8]:"8" = [28]
	//    │        ├──[3]:"36" = [36]
	//    │        ├──[4]:"4" = [4]
	//    │        │  └──[4]:"4" = [44]
	//    │        ├──[5]:"52" = [52]
	//    │        ├──[6]:"6"
	//    │        │  ├──[0]:"0" = [60]
	//    │        │  └──[8]:"8" = [68]
	//    │        ├──[7]:"76" = [76]
	//    │        ├──[8]:"84" = [84]
	//    │        └──[9]:"92" = [92]
	//    ├──[5]:"5"
	//    │  └─»┬
	//    │     └──[S]:"Street"
	//    │        ├──[1]:"13" = [13]
	//    │        ├──[2]:"2"
	//    │        │  ├──[1]:"1" = [21]
	//    │        │  └──[9]:"9" = [29]
	//    │        ├──[3]:"37" = [37]
	//    │        ├──[4]:"45" = [45]
	//    │        ├──[5]:"5" = [5]
	//    │        │  └──[3]:"3" = [53]
	//    │        ├──[6]:"6"
	//    │        │  ├──[1]:"1" = [61]
	//    │        │  └──[9]:"9" = [69]
	//    │        ├──[7]:"77" = [77]
	//    │        ├──[8]:"85" = [85]
	//    │        └──[9]:"93" = [93]
	//    ├──[6]:"6"
	//    │  └─»┬
	//    │     └──[S]:"Street"
	//    │        ├──[1]:"14" = [14]
	//    │        ├──[2]:"22" = [22]
	//    │        ├──[3]:"3"
	//    │        │  ├──[0]:"0" = [30]
	//    │        │  └──[8]:"8" = [38]
	//    │        ├──[4]:"46" = [46]
	//    │        ├──[5]:"54" = [54]
	//    │        ├──[6]:"6" = [6]
	//    │        │  └──[2]:"2" = [62]
	//    │        ├──[7]:"7"
	//    │        │  ├──[0]:"0" = [70]
	//    │        │  └──[8]:"8" = [78]
	//    │        ├──[8]:"86" = [86]
	//    │        └──[9]:"94" = [94]
	//    └──[7]:"7"
	//       └─»┬
	//          └──[S]:"Street"
	//             ├──[1]:"15" = [15]
	//             ├──[2]:"23" = [23]
	//             ├──[3]:"3"
	//             │  ├──[1]:"1" = [31]
	//             │  └──[9]:"9" = [39]
	//             ├──[4]:"47" = [47]
	//             ├──[5]:"55" = [55]
	//             ├──[6]:"63" = [63]
	//             ├──[7]:"7" = [7]
	//             │  ├──[1]:"1" = [71]
	//             │  └──[9]:"9" = [79]
	//             ├──[8]:"87" = [87]
	//             └──[9]:"95" = [95]
	//
}

func newTestDumper[T any](t *testing.T) func(prefix []byte, level int, end bool, values []T) bool {
	t.Helper()
	return newDumper[T](t.Log)
}

func TestRadix_Search(t *testing.T) {
	r := radix.New[string]()

	r.Insert("v1", false, []byte("a"))
	r.Insert("v2", false, []byte("ab"))
	r.Insert("v3", false, []byte("abc"))

	r.Insert("d1", false, []byte("123"))
	r.Insert("d2", false, []byte("124"))
	r.Insert("d3", false, []byte("125"))

	r.Insert("s1", false, []byte("user"), []byte("settings"), []byte("theme"))
	r.Insert("s2", false, []byte("user"), []byte("settings"), []byte("font"))
	r.Insert("s3", false, []byte("user"), []byte("profile"))
	r.Insert("s4", false, []byte("user"), []byte("profile"), nil, []byte("size"))

	r.Insert("a2", false, []byte{255, 255, 255, 0})
	r.Insert("a1", false, []byte{255, 0, 255, 255})
	r.Insert("a3", false, []byte("Привет"))

	r.Walk(newTestDumper[string](t))

	type args struct {
		prefixes [][]byte
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		// TODO: Add test cases.
		{
			name: "All under 'a'",
			args: args{prefixes: [][]byte{[]byte("a")}},
			want: []string{"v1", "v2", "v3"},
		},
		{
			name: "All under '\255'",
			args: args{prefixes: [][]byte{{255}}},
			want: []string{"a1", "a2"},
		},
		{
			name: "All under 'При'",
			args: args{prefixes: [][]byte{[]byte("При")}},
			want: []string{"a3"},
		},
		{
			name: "Exact 'ab'",
			args: args{prefixes: [][]byte{[]byte("ab")}},
			want: []string{"v2", "v3"},
		},
		{
			name: "Branching",
			args: args{prefixes: [][]byte{[]byte("12")}},
			want: []string{"d1", "d2", "d3"},
		},
		{
			name: "Specific branch",
			args: args{prefixes: [][]byte{[]byte("124")}},
			want: []string{"d2"},
		},
		{
			name: "Layered deep",
			args: args{prefixes: [][]byte{[]byte("user"), []byte("settings")}},
			want: []string{"s2", "s1"},
		},
		{
			name: "Layered skip middle",
			args: args{prefixes: [][]byte{[]byte("user"), nil, []byte("theme")}},
			want: []string{"s1"},
		},
		{
			name: "Layered skip both",
			args: args{prefixes: [][]byte{nil, nil, []byte("theme")}},
			want: []string{"s1"},
		},
		{
			name: "Layered skip empty",
			args: args{prefixes: [][]byte{nil, nil, nil, []byte("size")}},
			want: []string{"s4"},
		},
		{
			name: "Over-prefixed (not exists)",
			args: args{prefixes: [][]byte{[]byte("a"), []byte("extra")}},
			want: nil,
		},
		{
			name: "Longer search than value",
			args: args{prefixes: [][]byte{[]byte("abc1")}},
			want: nil,
		},
		{
			name: "All for nothing",
			args: args{prefixes: nil},
			want: []string{"d1", "d2", "d3", "v1", "v2", "v3", "s3", "s4", "s2", "s1", "a3", "a1", "a2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			i := r.Search(tt.args.prefixes...)
			for i.Next() {
				got = append(got, i.Get()...)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Search() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkRadix_100(b *testing.B) {
	t := radix.New[int]()

	for i := 0; i < 100; i++ {
		s := strconv.Itoa(i)
		c := strconv.Itoa(i % 8)
		t.Insert(i, false, []byte("City"+c), []byte("Street"+s))
	}

	d := func([]byte, int, bool, []int) bool { return true }

	b.ResetTimer()

	b.Run("Dump", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			t.Dump(d)
		}
	})

	b.Run("Walk", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			t.Walk(d)
		}
	})

	b.Run("Point", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			n := 0
			y := t.Search([]byte("City1"), []byte("Street41"))
			for y.Next() {
				v := y.Get()
				if len(v) != 1 || v[0] != 41 {
					b.Fatal(v)
				}
				n++
			}
			if n != 1 {
				b.Fatal(n, "!=", 1)
			}
		}
	})

	b.Run("Prefix", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			n := 0
			y := t.Search([]byte("City7"), []byte("Street7"))
			for y.Next() {
				v := y.Get()
				if len(v) != 1 {
					b.Fatal(v, 1)
				} else if v[0] > 7 && v[0]/10 != 7 {
					b.Fatal(v, 2)
				} else if v[0] < 8 && v[0] != 7 {
					b.Fatal(v, 3)
				}
				n++
			}
			if n != 3 {
				b.Fatal(n, "!=", 3)
			}
		}
	})

	b.Run("Deep", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			n := 0
			y := t.Search(nil, []byte("Street7"))
			for y.Next() {
				v := y.Get()
				if len(v) != 1 {
					b.Fatal(v, 1)
				} else if v[0] > 7 && v[0]/10 != 7 {
					b.Fatal(v, 2)
				} else if v[0] < 8 && v[0] != 7 {
					b.Fatal(v, 3)
				}
				n++
			}
			if n != 11 {
				b.Fatal(n, "!=", 11)
			}
		}
	})
}

func TestIterator_Next(t *testing.T) {
	if radix.New[int]().Search().Next() {
		t.Fatal(0)
	}
}
