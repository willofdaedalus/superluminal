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
			name: "with newline",
			args: args{
				data: []byte("\015"),
			},
			want: []token{
				{
					tokenType: newline,
					content:   []byte{'\n'},
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
				{
					tokenType: ansiColour,
					content:   []byte("[31m"),
				},
				{
					tokenType: normalText,
					content:   []byte("World"),
				},
				{
					tokenType: ansiColour,
					content:   []byte("[0m"),
				},
				{
					tokenType: normalText,
					content:   []byte("!"),
				},
			},
		},
		{
			name: "complex",
			args: args{
				data: []byte(`[?2004l[?1049h[1;40r[m[4l[?7h[?1h[?25l[39;49m[?1000h[39;49m`),
			},
			want: []token{
				{
					tokenType: ansiColour,
					content:   []byte("[m"),
				},
				{
					tokenType: ansiColour,
					content:   []byte("[39;49m"),
				},
				{
					tokenType: ansiColour,
					content:   []byte("[39;49m"),
				},
			},
		},
		{
			name: "usual",
			args: args{
				data: []byte(`]0;daedalus@theforge:~/projects/golang/personal/superluminal[?2004h[daedalus@theforge superluminal]$ ls --color=auto
[?2004lREADME.md  btop.output  go.mod  go.sum  hello  hello_log  [0m[01;34minternal[0m  log.output  log.output.bak  main.go  out.gif  [01;34mprotos[0m  [01;32mprotos.sh[0m  [01;34mscripts[0m  [01;32msuperluminal[0m  tape.vhs  [01;34mtmp[0m  tmux.output
]0;daedalus@theforge:~/projects/golang/personal/superluminal[?2004h[daedalus@theforge superluminal]$ exit
[?2004lexit
`),
			},
			want: []token{
				{
					tokenType: normalText,
					content:   []byte("[daedalus@theforge superluminal]$ ls --color=auto"),
				},
				{
					tokenType: newline,
					content:   []byte("\n"),
				},
				{
					tokenType: normalText,
					content:   []byte("README.md  btop.output  go.mod  go.sum  hello  hello_log  "),
				},
				{
					tokenType: ansiColour,
					content:   []byte("[0m"),
				},
				{
					tokenType: ansiColour,
					content:   []byte("[01;34m"),
				},
				{
					tokenType: normalText,
					content:   []byte("internal"),
				},
				{
					tokenType: ansiColour,
					content:   []byte("[0m"),
				},
				{
					tokenType: normalText,
					content:   []byte("  log.output  log.output.bak  main.go  out.gif  "),
				},
				{
					tokenType: ansiColour,
					content:   []byte("[01;34m"),
				},
				{
					tokenType: normalText,
					content:   []byte("protos"),
				},
				{
					tokenType: ansiColour,
					content:   []byte("[0m"),
				},
				{
					tokenType: normalText,
					content:   []byte("  "),
				},
				{
					tokenType: ansiColour,
					content:   []byte("[01;32m"),
				},
				{
					tokenType: normalText,
					content:   []byte("protos.sh"),
				},
				{
					tokenType: ansiColour,
					content:   []byte("[0m"),
				},
				{
					tokenType: normalText,
					content:   []byte("  "),
				},
				{
					tokenType: ansiColour,
					content:   []byte("[01;34m"),
				},
				{
					tokenType: normalText,
					content:   []byte("scripts"),
				},
				{
					tokenType: ansiColour,
					content:   []byte("[0m"),
				},
				{
					tokenType: normalText,
					content:   []byte("  "),
				},
				{
					tokenType: ansiColour,
					content:   []byte("[01;32m"),
				},
				{
					tokenType: normalText,
					content:   []byte("superluminal"),
				},
				{
					tokenType: ansiColour,
					content:   []byte("[0m"),
				},
				{
					tokenType: normalText,
					content:   []byte("  tape.vhs  "),
				},
				{
					tokenType: ansiColour,
					content:   []byte("[01;34m"),
				},
				{
					tokenType: normalText,
					content:   []byte("tmp"),
				},
				{
					tokenType: ansiColour,
					content:   []byte("[0m"),
				},
				{
					tokenType: normalText,
					content:   []byte("  tmux.output"),
				},
				{
					tokenType: newline,
					content:   []byte("\n"),
				},
				{
					tokenType: normalText,
					content:   []byte("[daedalus@theforge superluminal]$ exit"),
				},
				{
					tokenType: newline,
					content:   []byte("\n"),
				},
				{
					tokenType: normalText,
					content:   []byte("exit"),
				},
				{
					tokenType: newline,
					content:   []byte("\n"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scanner(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Scanner() = %v, want %v", got, tt.want)
				// for _, tk := range got {
				// 	t.Errorf("%s", string(tk.content))
				// }
			}
		})
	}
}
