package headers

import "testing"

func Test_transformHeaderBySpec(t *testing.T) {
	type args struct {
		header string
		spec   *Spec
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"normal", args{header: "Hello world\nThis is my header", spec: &Spec{Comment: "#"}}, "# Hello world\n# This is my header"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := transformHeaderBySpec(tt.args.header, tt.args.spec); got != tt.want {
				t.Errorf("transformHeaderBySpec() = %v, want %v", got, tt.want)
			}
		})
	}
}
