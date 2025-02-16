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
		{
			name: "simple complex",
			args: args{
				data: []byte(`
[?2004h[daedalus@theforge superluminal]$ ls
[?2004lREADME.md  btop.output	go.mod	go.sum	hello  hello_log  internal  log.output	main.go  out.gif  output.txt  protos  protos.sh  scripts  superluminal	tape.vhs  tmp  tmux.output
[?2004h[daedalus@theforge superluminal]$ whoami
[?2004ldaedalus
[?2004h[daedalus@theforge superluminal]$ exit
[?2004lexit
`),
			},
			want: []byte(`
[daedalus@theforge superluminal]$ ls
README.md  btop.output	go.mod	go.sum	hello  hello_log  internal  log.output	main.go  out.gif  output.txt  protos  protos.sh  scripts  superluminal	tape.vhs  tmp  tmux.output
[daedalus@theforge superluminal]$ whoami
daedalus
[daedalus@theforge superluminal]$ exit
exit
`),
		},

		{
			name: "simple colour output test",
			args: args{
				data: []byte(`
]0;daedalus@theforge:~/projects/golang/personal/superluminal[?2004h[daedalus@theforge superluminal]$ ls --color=auto
[?2004lREADME.md  btop.output  go.mod  go.sum  hello  hello_log  [0m[01;34minternal[0m  log.output  log.output.bak  main.go  out.gif  [01;34mprotos[0m  [01;32mprotos.sh[0m  [01;34mscripts[0m  [01;32msuperluminal[0m  tape.vhs  [01;34mtmp[0m  tmux.output
]0;daedalus@theforge:~/projects/golang/personal/superluminal[?2004h[daedalus@theforge superluminal]$ exit
[?2004lexit
`),
			},
			want: []byte(`
`),
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
