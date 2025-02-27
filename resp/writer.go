package resp

import "io"

type IWriter interface {
	Write(v Value) error
}

type writer struct {
	wr io.Writer
}

func NewWriter(w io.Writer) *writer {
	return &writer{wr: w}
}

func (w *writer) Write(v Value) error {
	var bytes = v.Marshal()

	_, err := w.wr.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}
