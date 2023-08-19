package alias

import "testing"

func TestGenerate(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{
			name: "Check alias length",
			size: 6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Generate(tt.size); len(got) != tt.size {
				t.Errorf("Generate() = %v, want %v", len(got), tt.size)
			}
		})
	}
}
