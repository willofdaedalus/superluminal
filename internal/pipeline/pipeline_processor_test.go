package pipeline

import (
	"reflect"
	"testing"
)

func Test_skipAheadTo(t *testing.T) {
	type args struct {
		data []byte
		b    byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "byte present in data",
			args: args{data: []byte{1, 2, 3, 4, 5}, b: 3},
			want: []byte{4, 5},
		},
		{
			name: "byte not present in data",
			args: args{data: []byte{1, 2, 4, 5}, b: 3},
			want: []byte{1, 2, 4, 5},
		},
		{
			name: "byte is the first element",
			args: args{data: []byte{3, 4, 5}, b: 3},
			want: []byte{4, 5},
		},
		{
			name: "empty data slice",
			args: args{data: []byte{}, b: 3},
			want: []byte{},
		},
		{
			name: "multiple occurrences of byte",
			args: args{data: []byte{1, 3, 2, 3, 4}, b: 3},
			want: []byte{2, 3, 4},
		},
		{
			name: "single element equal to byte",
			args: args{data: []byte{3}, b: 3},
			want: []byte{},
		},
		{
			name: "single element not equal to byte",
			args: args{data: []byte{1}, b: 3},
			want: []byte{1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := skipAheadTo(tt.args.data, []byte{tt.args.b}); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("skipAheadTo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sanitizeRawCode(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// === OSC TEST CASES ===
		{
			name: "with osc",
			args: args{[]byte{OSC, 'a', 'b', 'c', 'd', '\a'}},
			want: []byte{},
		},
		{
			name: "no sequences",
			args: args{[]byte{'a', 'b', 'c', 'd', '\a'}},
			want: []byte{'a', 'b', 'c', 'd', '\a'},
		},
		{
			name: "multiple osc in a row",
			args: args{[]byte{OSC, 'a', OSC, OSC, OSC, '\a', 'a', 'b', 'c', OSC, 'c', 'd', '\a'}},
			want: []byte("abc"),
		},
		{
			name: "multiple osc",
			args: args{[]byte("\033]0;daedalus@theforge:~/projects/golang/personal/superluminal\007\033[?2004h[daedalus@theforge superluminal]$ ")},
			want: []byte("[?2004h[daedalus@theforge superluminal]$ "),
		},
		{
			name: "no ending osc",
			args: args{[]byte{OSC}},
			want: []byte{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeRawCode(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sanitizeRawCode() = %v, want %v w len %d", got, tt.want, len(got))
			}
		})
	}
}
