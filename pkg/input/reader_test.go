package input

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadContextFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "context.txt")
	if err := os.WriteFile(path, []byte("hello\nworld"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := ReadContext(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello\nworld" {
		t.Fatalf("ReadContext() = %q, want %q", got, "hello\nworld")
	}
}
