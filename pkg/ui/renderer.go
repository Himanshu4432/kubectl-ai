package ui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

type Renderer struct {
	spinner *spinner.Spinner
	writer  io.Writer
}

func NewRenderer(w io.Writer) *Renderer {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(w))
	s.Suffix = " Fetching cluster state and diagnosing workload..."
	s.Color("cyan")
	return &Renderer{
		spinner: s,
		writer:  w,
	}
}

func (r *Renderer) StartSpinner() {
	r.spinner.Start()
}

func (r *Renderer) StopSpinner() {
	r.spinner.Stop()
}

// StreamPrinter handles streaming markdown tokens on the fly with custom ANSI highlighting
type StreamPrinter struct {
	writer      io.Writer
	inCodeBlock bool
	currentLine strings.Builder
}

func NewStreamPrinter(w io.Writer) *StreamPrinter {
	return &StreamPrinter{writer: w}
}

func (sp *StreamPrinter) Print(token string) {
	for _, char := range token {
		sp.currentLine.WriteRune(char)
		if char == '\n' {
			sp.flushLine()
		}
	}

	if strings.Contains(token, "```") {
		sp.inCodeBlock = !sp.inCodeBlock
		if sp.inCodeBlock {
			color.New(color.FgCyan).Print("\n")
		} else {
			color.New(color.Reset).Print("\n")
		}
		return
	}

	if sp.inCodeBlock {
		color.New(color.FgCyan).Print(token)
	} else {
		// Highlight markdown structures in real-time
		if strings.HasPrefix(token, "##") {
			color.New(color.FgMagenta, color.Bold).Print(token)
		} else if strings.HasPrefix(token, "🚨") || strings.HasPrefix(token, "📊") || strings.HasPrefix(token, "🛠️") {
			color.New(color.Bold).Print(token)
		} else {
			fmt.Fprint(sp.writer, token)
		}
	}
}

func (sp *StreamPrinter) flushLine() {
	sp.currentLine.Reset()
}
