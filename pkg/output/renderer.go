package output

import (
	"io"
	"os"

	"github.com/charmbracelet/glamour"
	"golang.org/x/term"
)

type Renderer struct {
	isTTY bool
	raw   bool
}

func NewRenderer(rawFlag bool) *Renderer {
	return &Renderer{isTTY: term.IsTerminal(int(os.Stdout.Fd())), raw: rawFlag}
}

func (r *Renderer) Render(markdownText string, w io.Writer) error {
	if r.raw || !r.isTTY {
		_, err := io.WriteString(w, markdownText)
		return err
	}
	renderer, err := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(100))
	if err != nil {
		_, err = io.WriteString(w, markdownText)
		return err
	}
	rendered, err := renderer.Render(markdownText)
	if err != nil {
		_, err = io.WriteString(w, markdownText)
		return err
	}
	_, err = io.WriteString(w, rendered)
	return err
}
