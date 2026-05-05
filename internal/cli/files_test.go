package cli

import (
	"path/filepath"
	"testing"
)

func TestShouldSkipLocalPath(t *testing.T) {
	root := filepath.Clean("/tmp/site")

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "root is not skipped",
			path: root,
			want: false,
		},
		{
			name: "hidden file is skipped",
			path: filepath.Join(root, ".DS_Store"),
			want: true,
		},
		{
			name: "hidden nested file is skipped",
			path: filepath.Join(root, "app", ".env"),
			want: true,
		},
		{
			name: "hidden directory child is skipped",
			path: filepath.Join(root, ".git", "config"),
			want: true,
		},
		{
			name: "node modules child is skipped",
			path: filepath.Join(root, "node_modules", "pkg", "index.js"),
			want: true,
		},
		{
			name: "build output child is skipped",
			path: filepath.Join(root, "dist", "bundle.js"),
			want: true,
		},
		{
			name: "regular source file is not skipped",
			path: filepath.Join(root, "src", "index.ts"),
			want: false,
		},
		{
			name: "similar visible name is not skipped",
			path: filepath.Join(root, "src", "node_modules_backup", "index.ts"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldSkipLocalPath(root, tt.path)
			if got != tt.want {
				t.Fatalf("shouldSkipLocalPath(%q, %q) = %v, want %v", root, tt.path, got, tt.want)
			}
		})
	}
}

func TestIsUTF8FileBody(t *testing.T) {
	tests := []struct {
		name string
		body []byte
		want bool
	}{
		{
			name: "plain text is accepted",
			body: []byte("hello\n"),
			want: true,
		},
		{
			name: "utf8 text is accepted",
			body: []byte("你好\n"),
			want: true,
		},
		{
			name: "empty file is accepted",
			body: nil,
			want: true,
		},
		{
			name: "png-like binary is skipped",
			body: []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a},
			want: false,
		},
		{
			name: "invalid utf8 is skipped",
			body: []byte{0xff, 0xfe, 0xfd},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUTF8FileBody(tt.body)
			if got != tt.want {
				t.Fatalf("isUTF8FileBody(%v) = %v, want %v", tt.body, got, tt.want)
			}
		})
	}
}
