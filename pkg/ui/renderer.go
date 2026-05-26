package ui

import (
	"io"
	"time"

	"github.com/briandowns/spinner"
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
