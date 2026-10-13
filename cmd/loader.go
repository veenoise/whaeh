package cmd

import (
	"fmt"
	"io"
	"sync"
	"time"
)

type loader struct {
	stop chan struct{}
	done chan struct{}
	once sync.Once
	out  io.Writer
}

func startLoader(out io.Writer) *loader {
	l := &loader{
		stop: make(chan struct{}),
		done: make(chan struct{}),
		out:  out,
	}

	go func() {
		defer close(l.done)
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		frame := 0
		writeLoaderFrame := func() {
			_, _ = fmt.Fprintf(l.out, "\r\033[2K%s Thinking...", frames[frame])
			frame = (frame + 1) % len(frames)
		}
		writeLoaderFrame()

		for {
			select {
			case <-ticker.C:
				writeLoaderFrame()
			case <-l.stop:
				return
			}
		}
	}()

	return l
}

func (l *loader) stopLoader() {
	l.once.Do(func() {
		close(l.stop)
		<-l.done
		_, _ = fmt.Fprint(l.out, "\r\033[2K")
	})
}
