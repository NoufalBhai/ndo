package recipe

import "testing"

func TestRecipeBind(t *testing.T) {
	tests := []struct {
		name    string
		recipe  Recipe
		args    []string
		want    func() string
		wantErr bool
	}{
		{
			name:   "exact param count",
			recipe: Recipe{Command: "code {{file}}", Params: []string{"file"}},
			args:   []string{"main.go"},
			want:   func() string { return "code " + QuoteArg("main.go") },
		},
		{
			name:    "missing required param is hard error",
			recipe:  Recipe{Command: "code {{file}}", Params: []string{"file"}},
			args:    []string{},
			wantErr: true,
		},
		{
			name:   "leftover args passed through raw, quoted",
			recipe: Recipe{Command: "go test ./...", Params: []string{}},
			args:   []string{"-run", "TestFoo"},
			want: func() string {
				return "go test ./... " + QuoteArg("-run") + " " + QuoteArg("TestFoo")
			},
		},
		{
			name:   "value with spaces gets quoted",
			recipe: Recipe{Command: "echo {{msg}}", Params: []string{"msg"}},
			args:   []string{"hello world"},
			want:   func() string { return "echo " + QuoteArg("hello world") },
		},
		{
			name:   "raw escape hatch skips quoting",
			recipe: Recipe{Command: "echo {{msg|raw}}", Params: []string{"msg"}},
			args:   []string{"hello world"},
			want:   func() string { return "echo hello world" },
		},
		{
			name:   "empty params recipe with no args",
			recipe: Recipe{Command: "go test ./...", Params: []string{}},
			args:   []string{},
			want:   func() string { return "go test ./..." },
		},
		{
			name:    "undeclared parameter referenced in command",
			recipe:  Recipe{Command: "echo {{oops}}", Params: []string{}},
			args:    []string{},
			wantErr: true,
		},
		{
			name:   "value with quote-significant character is escaped safely",
			recipe: Recipe{Command: "echo {{msg}}", Params: []string{"msg"}},
			args:   []string{`it's "quoted"`},
			want:   func() string { return "echo " + QuoteArg(`it's "quoted"`) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.recipe.Bind(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Bind() = %q, want error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Bind() unexpected error: %v", err)
			}
			if want := tt.want(); got != want {
				t.Fatalf("Bind() = %q, want %q", got, want)
			}
		})
	}
}
