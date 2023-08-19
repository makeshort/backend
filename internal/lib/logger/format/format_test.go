package format

import "testing"

func Test_CompleteStringToLength(t *testing.T) {
	type args struct {
		s      string
		length int
		char   rune
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Correct string completion",
			args: args{
				s:      "string",
				length: 10,
				char:   ' ',
			},
			want: "string    ",
		},
		{
			name: "Length less that s.length",
			args: args{
				s:      "string",
				length: 3,
				char:   ' ',
			},
			want: "str",
		},
		{
			name: "Length equals s.length",
			args: args{
				s:      "string",
				length: 6,
				char:   ' ',
			},
			want: "string",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CompleteStringToLength(tt.args.s, tt.args.length, tt.args.char); got != tt.want {
				t.Errorf("completeStringToLength() = %v, want %v", got, tt.want)
			}
		})
	}
}
