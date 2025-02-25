package parser

import (
	"reflect"
	"testing"
)

func TestEncodeTokens(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// {
		// 	name: "simple test",
		// 	args: args{
		// 		data: []byte("abcdef"),
		// 	},
		// 	want: []byte("abcdef"),
		// },
		// {
		// 	name: "testing rle",
		// 	args: args{
		// 		data: []byte("aaaaabbcdddddaaakj"),
		// 	},
		// 	want: []byte("a5b2cd5a3kj"),
		// },
		{
			name: "simple ansi code",
			args: args{
				data: []byte("\x1B[37mhello world"),
			},
			want: []byte("\x1B070hel2o world"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeTokens(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EncodeTokens() = %v, want %v", got, tt.want)
			}
		})
	}
}
