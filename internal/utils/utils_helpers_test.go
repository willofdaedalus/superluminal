package utils

import (
	"reflect"
	"testing"
)

func TestRLEncode(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "simple rle",
			args: args{
				data: []byte("wwwwbwwww"),
			},
			want: []byte("w4bw4"),
		},
		{
			name: "single character input",
			args: args{
				data: []byte("a"),
			},
			want: []byte("a"),
		},
		{
			name: "no input (empty slice)",
			args: args{
				data: []byte{},
			},
			want: nil,
		},
		{
			name: "all unique characters",
			args: args{
				data: []byte("abcdef"),
			},
			want: []byte("abcdef"),
		},
		{
			name: "long run of one character",
			args: args{
				data: []byte("zzzzzzzzzzzzzzzzzzzz"), // 20 'z's
			},
			want: []byte("z20"),
		},
		{
			name: "run length exceeding 9 (checks string conversion)",
			args: args{
				data: []byte("aaaaaaaaaa"), // 10 'a's
			},
			want: []byte("a10"),
		},
		{
			name: "mixed short and long runs",
			args: args{
				data: []byte("aaabbcddddddee"),
			},
			want: []byte("a3b2cd6e2"),
		},
		{
			name: "digits in input (should not interfere with encoding)",
			args: args{
				data: []byte("1112223333"),
			},
			want: []byte("13" + "23" + "34"), // Encodes as '1' three times, '2' three times, '3' four times
		},
		{
			name: "alternating characters",
			args: args{
				data: []byte("ababababab"),
			},
			want: []byte("ababababab"),
		},
		{
			name: "mixed",
			args: args{
				data: []byte("aabaabbbabc"),
			},
			want: []byte("a2ba2b3abc"),
		},
		{
			name: "whitespace characters",
			args: args{
				data: []byte("    \t\t\n\n\n"),
			},
			want: []byte(" 4\t2\n3"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RLEncode(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RLEncode() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

func TestRLDecode(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "simple rle",
			args: args{
				data: []byte("w4bw4"),
			},
			want: []byte("wwwwbwwww"),
		},
		{
			name: "single character input",
			args: args{
				data: []byte("a"),
			},
			want: []byte("a"),
		},
		{
			name: "no input (empty slice)",
			args: args{
				data: []byte{},
			},
			want: nil,
		},
		{
			name: "all unique characters",
			args: args{
				data: []byte("abcdef"),
			},
			want: []byte("abcdef"),
		},
		{
			name: "long run of one character",
			args: args{
				data: []byte("z20"), // 20 'z's
			},
			want: []byte("zzzzzzzzzzzzzzzzzzzz"),
		},
		{
			name: "run length exceeding 9 (checks string conversion)",
			args: args{
				data: []byte("a10"), // 10 'a's
			},
			want: []byte("aaaaaaaaaa"),
		},
		{
			name: "mixed short and long runs",
			args: args{
				data: []byte("a3b2cd6e2"),
			},
			want: []byte("aaabbcddddddee"),
		},
		// {
		// 	name: "digits in input (should not interfere with encoding)",
		// 	args: args{
		// 		data: []byte("1112223333"),
		// 	},
		// 	want: []byte("13" + "23" + "34"), // Encodes as '1' three times, '2' three times, '3' four times
		// },
		{
			name: "alternating characters",
			args: args{
				data: []byte("ababababab"),
			},
			want: []byte("ababababab"),
		},
		{
			name: "mixed",
			args: args{
				data: []byte("a2ba2b3abc"),
			},
			want: []byte("aabaabbbabc"),
		},
		// {
		// 	name: "whitespace characters",
		// 	args: args{
		// 		data: []byte("    \t\t\n\n\n"),
		// 	},
		// 	want: []byte(" 4\t2\n3"),
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RLDecode(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RLDecode() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}
