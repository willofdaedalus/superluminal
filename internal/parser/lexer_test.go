package parser

import (
	"reflect"
	"testing"
)

func TestScanner(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want []token
	}{
		{
			name: "simple",
			args: args{
				data: []byte("helloworld"),
			},
			want: []token{
				{
					tokenType: normalText,
					content:   []byte("helloworld"),
				},
			},
		},
		{
			name: "with esc",
			args: args{
				data: []byte("hello\x1b[31mWorld\x1b[0m!"),
			},
			want: []token{
				{
					tokenType: normalText,
					content:   []byte("hello"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scanner(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Scanner() = %v, want %v", got, tt.want)
			}
		})
	}
}
