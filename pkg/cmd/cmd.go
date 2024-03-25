package cmd

import (
	"io"
)

type Command interface {
	Run(io.Reader, io.Writer, ...string) error
}
