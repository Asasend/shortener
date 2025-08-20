package base62

import "testing"

func TestInt2String(t *testing.T) {
	tests := []struct {
		name string
		seq  uint64 // 直接放入参数
		want string
	}{
		{
			name: "zero",
			seq:  0,
			want: "0",
		},
		{
			name: "one",
			seq:  1,
			want: "1",
		},
		{
			name: "ten",
			seq:  10,
			want: "a",
		},
		{
			name: "thirty-six",
			seq:  36,
			want: "A",
		},
		{
			name: "sixty-one",
			seq:  61,
			want: "Z",
		},
		{
			name: "sixty-two",
			seq:  62,
			want: "10",
		},
		{
			name: "large-number",
			seq:  123456,
			want: "w7e",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 直接使用 tt.seq 而不是 tt.args.seq
			if got := Int2String(tt.seq); got != tt.want {
				t.Errorf("Int2String() = %v, want %v", got, tt.want)
			}
		})
	}
}
