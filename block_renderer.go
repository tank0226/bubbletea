package tea

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/muesli/reflow/ansi"
	te "github.com/muesli/termenv"
)

// Block is a rectangular, positionable container.
type Block struct {
	X       int
	Y       int
	Z       int
	Content string

	currentLines []string
}

func (b Block) lines() []string {
	return strings.Split(b.Content, "\n")
}

// byZIndex is used to sort Blocks by their Z value.
type byZIndex []Block

func (b byZIndex) Len() int           { return len(b) }
func (b byZIndex) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byZIndex) Less(i, j int) bool { return b[i].Z < b[j].Z }

// blockRenderer renders Blocks in arbitrary locations, allowing for more
// complex layouts.
type blockRenderer struct {
	out       io.Writer
	framerate time.Duration
	ticker    *time.Ticker
	mtx       *sync.Mutex
	done      chan struct{}

	buffer        []Block
	lastRender    []Block
	linesRendered int

	width  int
	height int
}

func newBlockRenderer(out io.Writer, mtx *sync.Mutex) *blockRenderer {
	return &blockRenderer{
		out:       out,
		mtx:       mtx,
		framerate: defaultFramerate,
	}
}

func (r *blockRenderer) start() {
	if r.ticker == nil {
		r.ticker = time.NewTicker(r.framerate)
	}
	r.done = make(chan struct{})
	go r.listen()
}

// stop permanently halts the renderer.
func (r *blockRenderer) stop() {
	r.flush()
	r.done <- struct{}{}
}

// listen waits for ticks on the ticker, or a signal to stop the renderer.
func (r *blockRenderer) listen() {
	for {
		select {
		case <-r.ticker.C:
			if r.ticker != nil {
				r.flush()
			}
		case <-r.done:
			r.mtx.Lock()
			r.ticker.Stop()
			r.ticker = nil
			r.mtx.Unlock()
			close(r.done)
			return
		}
	}
}

func (r *blockRenderer) write(b []Block) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.buffer = b
}

// flush renders the buffer.
func (r *blockRenderer) flush() {

	sort.Sort(byZIndex(r.buffer))
	out := newBlockBuffer()

	out.cursorUp(r.linesRendered)

	var linesRendered int

	// get total height of area we're going to render
	for _, block := range r.buffer {
		linesRendered = max(linesRendered, len(block.lines())+block.Y)
	}

	// make area we're going to render available
	if linesRendered > r.linesRendered {
		_, _ = out.WriteString(strings.Repeat("\r\n", linesRendered))
		out.cursorUp(linesRendered)
	}

	var x, y int

	// render blocks
	for _, block := range r.buffer {

		// move cursor back up to line 0
		out.cursorUp(y)

		// move cursor to xy coordinates of block
		out.cursorForward(block.X)
		out.cursorDown(block.Y)

		// render block content
		lines := block.lines()
		for _, line := range lines {
			out.cursorDown(1)
			_, _ = io.WriteString(out, line)
			width := ansi.PrintableRuneWidth(line)
			out.cursorBack(width)
			x = max(x, block.X+width)
		}

		y = block.Y + len(lines) - 1
		out.cursorBack(x)
	}

	//_, _ = out.WriteString("\r\n")

	_, _ = r.out.Write(out.Bytes())
	r.linesRendered = linesRendered // + 1
}

func (r *blockRenderer) handleMessages(msg Msg) {
	switch msg := msg.(type) {
	case WindowSizeMsg:
		r.width = msg.Width
		r.height = msg.Height
	}
}

// blockBuffer is used in blockRenderer.flush() to translate a Block into a
// bytestring.
type blockBuffer struct {
	*bytes.Buffer
}

func newBlockBuffer() blockBuffer {
	return blockBuffer{new(bytes.Buffer)}
}

func (b *blockBuffer) cursorDown(n int) {
	fmt.Fprintf(b, te.CSI+te.CursorDownSeq, n)
}

func (b *blockBuffer) cursorUp(n int) {
	fmt.Fprintf(b, te.CSI+te.CursorUpSeq, n)
}

func (b *blockBuffer) cursorBack(n int) {
	fmt.Fprintf(b, te.CSI+te.CursorBackSeq, n)
}

func (b *blockBuffer) cursorForward(n int) {
	fmt.Fprintf(b, te.CSI+te.CursorForwardSeq, n)
}
