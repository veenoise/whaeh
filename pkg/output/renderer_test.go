package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestStreamWriter(t *testing.T) {
	var buf bytes.Buffer
	called := 0
	sw := NewStreamWriter(&buf, func() {
		called++
	})

	if _, err := sw.Write([]byte("chunk1")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if _, err := sw.Write([]byte("chunk2")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if called != 1 {
		t.Errorf("expected onFirst called 1 time, got %d", called)
	}
	if buf.String() != "chunk1chunk2" {
		t.Errorf("expected chunk1chunk2, got %s", buf.String())
	}
}

func TestRender(t *testing.T) {
	var buf bytes.Buffer
	input := "# Heading\n\nSome text with `inline code` and:\n\n```go\nfunc main() {}\n```"
	if err := Render(input, &buf); err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Heading") {
		t.Errorf("expected 'Heading' in output, got: %s", out)
	}
	if !strings.Contains(out, "func") {
		t.Errorf("expected 'func' in output, got: %s", out)
	}
}
