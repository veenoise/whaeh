package output

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type StreamWriter struct {
	w       io.Writer
	onFirst func()
	once    sync.Once
}

func NewStreamWriter(w io.Writer, onFirst func()) *StreamWriter {
	return &StreamWriter{w: w, onFirst: onFirst}
}

func (s *StreamWriter) Write(p []byte) (int, error) {
	if s.onFirst != nil {
		s.once.Do(s.onFirst)
	}
	return s.w.Write(p)
}

func Render(markdownText string, w io.Writer) error {
	src := []byte(markdownText)
	doc := goldmark.New().Parser().Parse(text.NewReader(src))

	return ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		switch n.Kind() {
		case ast.KindHeading:
			h := n.(*ast.Heading)
			if entering {
				color := "\033[1;36m"
				if h.Level == 1 {
					color = "\033[1;35m"
				}
				fmt.Fprintf(w, "\n%s", color)
			} else {
				fmt.Fprint(w, "\033[0m\n\n")
			}
		case ast.KindParagraph:
			if !entering {
				fmt.Fprint(w, "\n\n")
			}
		case ast.KindCodeSpan:
			if entering {
				fmt.Fprint(w, "\033[33m`")
			} else {
				fmt.Fprint(w, "`\033[0m")
			}
		case ast.KindEmphasis:
			e := n.(*ast.Emphasis)
			if entering {
				if e.Level > 1 {
					fmt.Fprint(w, "\033[1m")
				} else {
					fmt.Fprint(w, "\033[3m")
				}
			} else {
				fmt.Fprint(w, "\033[0m")
			}
		case ast.KindFencedCodeBlock:
			if entering {
				fc := n.(*ast.FencedCodeBlock)
				lang := strings.TrimSpace(string(fc.Language(src)))
				code := nodeBytes(fc.Lines(), src)
				if lang != "" {
					fmt.Fprintf(w, "\033[2m─── %s ───\033[0m\n", lang)
				} else {
					fmt.Fprint(w, "\033[2m────────────\033[0m\n")
				}
				if err := highlight(w, string(code), lang); err != nil {
					w.Write(code)
				}
				fmt.Fprint(w, "\033[2m────────────\033[0m\n\n")
			}
			return ast.WalkSkipChildren, nil
		case ast.KindCodeBlock:
			if entering {
				cb := n.(*ast.CodeBlock)
				code := nodeBytes(cb.Lines(), src)
				if err := highlight(w, string(code), ""); err != nil {
					w.Write(code)
				}
				fmt.Fprintln(w)
			}
			return ast.WalkSkipChildren, nil
		case ast.KindBlockquote:
			if entering {
				fmt.Fprint(w, "\033[2;3m│ ")
			} else {
				fmt.Fprint(w, "\033[0m\n")
			}
		case ast.KindListItem:
			if entering {
				fmt.Fprint(w, "  • ")
			} else {
				fmt.Fprintln(w)
			}
		case ast.KindThematicBreak:
			if entering {
				fmt.Fprint(w, "\033[2m────────────────────────\033[0m\n\n")
			}
		case ast.KindText:
			if entering {
				t := n.(*ast.Text)
				w.Write(t.Segment.Value(src))
				if t.HardLineBreak() || t.SoftLineBreak() {
					fmt.Fprintln(w)
				}
			}
		case ast.KindString:
			if entering {
				s := n.(*ast.String)
				w.Write(s.Value)
			}
		}
		return ast.WalkContinue, nil
	})
}

func nodeBytes(lines *text.Segments, src []byte) []byte {
	var buf bytes.Buffer
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		buf.Write(seg.Value(src))
	}
	return buf.Bytes()
}

func highlight(w io.Writer, code, lang string) error {
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Analyse(code)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}
	iter, err := lexer.Tokenise(nil, code)
	if err != nil {
		return err
	}
	return formatter.Format(w, style, iter)
}
