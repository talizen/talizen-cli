package cli

import "testing"

func TestReleaseTag(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
		wantErr bool
	}{
		{
			name:    "plain semver",
			version: "0.1.0",
			want:    "v0.1.0",
		},
		{
			name:    "already prefixed",
			version: "v0.1.0",
			want:    "v0.1.0",
		},
		{
			name:    "trim spaces",
			version: " 0.1.0 ",
			want:    "v0.1.0",
		},
		{
			name:    "dev cannot publish",
			version: "dev",
			wantErr: true,
		},
		{
			name:    "empty cannot publish",
			version: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := releaseTag(tt.version)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("releaseTag(%q) = %q, want %q", tt.version, got, tt.want)
			}
		})
	}
}
