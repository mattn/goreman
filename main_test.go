package main

import "testing"

func Test_version(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "VersionTest",
			want: "0.3.15",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := version; got != tt.want {
				t.Errorf("version = %v, want %v", got, tt.want)
			}
		})
	}
}
