package hash

import (
	"reflect"
	"testing"
)

func TestHasher_Create(t *testing.T) {
	tests := []struct {
		name string
		salt string
		s    string
		want string
	}{
		{
			name: "Test empty salt",
			salt: "",
			s:    "password",
			want: "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8",
		},
		{
			name: "Test ' ' as salt",
			salt: " ",
			s:    "password",
			want: "d41ca5c2608fab7103b6209f00713655fb6cc53a5dbf73e30c4b8ebeaa082a9f",
		},
		{
			name: "Test with salt",
			salt: "salt",
			s:    "password",
			want: "7a37b85c8918eac19a9089c0fa5a2ab4dce3f90528dcdeec108b23ddf3607b99",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Hasher{
				salt: tt.salt,
			}
			if got := h.Create(tt.s); got != tt.want {
				t.Errorf("Create() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		salt string
	}
	tests := []struct {
		name string
		args args
		want *Hasher
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.salt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
