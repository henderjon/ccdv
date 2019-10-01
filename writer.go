// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ccdv

import (
	"bufio"
	"io"
	"strings"
)

// A Writer writes records using CSV encoding.
//
// As returned by NewWriter, a Writer writes records terminated by a
// newline and uses ',' as the field delimiter. The exported fields can be
// changed to customize the details before the first call to Write or WriteAll.
//
// Comma is the field delimiter.
//
// If UseCRLF is true, the Writer ends each output line with \r\n instead of \n.
//
// The writes of individual records are buffered.
// After all data has been written, the client should call the
// Flush method to guarantee all data has been forwarded to
// the underlying io.Writer.  Any errors that occurred should
// be checked by calling the Error method.
type Writer struct {
	w *bufio.Writer
}

// NewWriter returns a new Writer that writes to w.
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w: bufio.NewWriter(w),
	}
}

// Write writes a single CSV record to w along with any necessary quoting.
// A record is a slice of strings with each string being one field.
// Writes are buffered, so Flush must eventually be called to ensure
// that the record is written to the underlying io.Writer.
func (w *Writer) Write(record []string) error {
	var err error

	for n, field := range record {
		if n > 0 {
			if _, err := w.w.WriteRune(UnitSep); err != nil {
				return err
			}
		}

		// we don't allow the GroupSep, RecordSep, and UnitSep runes in strings
		if err = w.validateField(field); err != nil {
			return err
		}

		if _, err = w.w.WriteString(field); err != nil {
			return err
		}

		continue
	}

	_, err = w.w.WriteRune(RecordSep)

	return err
}

// Flush writes any buffered data to the underlying io.Writer.
// To check if an error occurred during the Flush, call Error.
func (w *Writer) Flush() {
	w.w.Flush()
}

// Error reports any error that has occurred during a previous Write or Flush.
func (w *Writer) Error() error {
	_, err := w.w.Write(nil)
	return err
}

// WriteAll writes multiple CSV records to w using Write and then calls Flush,
// returning any error from the Flush.
func (w *Writer) WriteAll(records [][]string) error {
	for _, record := range records {
		err := w.Write(record)
		if err != nil {
			return err
		}
	}
	return w.w.Flush()
}

func (w *Writer) validateField(field string) error {
	if field == "" {
		return nil
	}
	if strings.ContainsAny(field, "\035\036\037") {
		return errInvalidField
	}

	return nil
}
