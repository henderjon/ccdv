// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ccdv

import (
	"bytes"
	"errors"
	"testing"
)

var writeTests = []struct {
	Input   [][]string
	Output  string
	Error   error
	UseCRLF bool
	Comma   rune
}{
	{Input: [][]string{{"abc"}}, Output: "abc\036"},
	{Input: [][]string{{`"abc"`}}, Output: `"abc"` + string(RecordSep)},
	{Input: [][]string{{`a"b`}}, Output: `a"b` + string(RecordSep)},
	{Input: [][]string{{`"a"b"`}}, Output: `"a"b"` + string(RecordSep)},
	{Input: [][]string{{" abc"}}, Output: ` abc` + string(RecordSep)},
	{Input: [][]string{{"abc,def"}}, Output: `abc,def` + string(RecordSep)},
	{Input: [][]string{{"abc", "def"}}, Output: "abc\037def\036"},
	{Input: [][]string{{"abc"}, {"def"}}, Output: "abc\036def\036"},
	{Input: [][]string{{"abc\ndef"}}, Output: "abc\ndef\036"},
	{Input: [][]string{{"abc\rdef"}}, Output: "abc\rdef\036"},
	{Input: [][]string{{""}}, Output: "\036"},
	{Input: [][]string{{"", ""}}, Output: "\037\036"},
	{Input: [][]string{{"", "", ""}}, Output: "\037\037\036"},
	{Input: [][]string{{"", "", "a"}}, Output: "\037\037a\036"},
	{Input: [][]string{{"", "a", ""}}, Output: "\037a\037\036"},
	{Input: [][]string{{"", "a", "a"}}, Output: "\037a\037a\036"},
	{Input: [][]string{{"a", "", ""}}, Output: "a\037\037\036"},
	{Input: [][]string{{"a", "", "a"}}, Output: "a\037\037a\036"},
	{Input: [][]string{{"a", "a", ""}}, Output: "a\037a\037\036"},
	{Input: [][]string{{"a", "a", "a"}}, Output: "a\037a\037a\036"},
	{Input: [][]string{{`\.`}}, Output: `\.` + string(RecordSep)},
	{Input: [][]string{{"x09\x41\xb4\x1c", "aktau"}}, Output: "x09\x41\xb4\x1c\037aktau\036"},
	{Input: [][]string{{",x09\x41\xb4\x1c", "aktau"}}, Output: ",x09\x41\xb4\x1c\037aktau\036"},
	{Input: [][]string{{"fo\036o"}}, Comma: '"', Error: errInvalidField},
}

func TestWrite(t *testing.T) {
	for n, tt := range writeTests {
		b := &bytes.Buffer{}
		f := NewWriter(b)

		err := f.WriteAll(tt.Input)
		if err != tt.Error {
			t.Errorf("Unexpected error:\ngot  %v\nwant %v", err, tt.Error)
		}
		out := b.String()
		if out != tt.Output {
			t.Errorf("#%d: out=%q want %q", n, out, tt.Output)
		}
	}
}

type errorWriter struct{}

func (e errorWriter) Write(b []byte) (int, error) {
	return 0, errors.New("Test")
}

func TestError(t *testing.T) {
	b := &bytes.Buffer{}
	f := NewWriter(b)
	f.Write([]string{"abc"})
	f.Flush()
	err := f.Error()

	if err != nil {
		t.Errorf("Unexpected error: %s\n", err)
	}

	f = NewWriter(errorWriter{})
	f.Write([]string{"abc"})
	f.Flush()
	err = f.Error()

	if err == nil {
		t.Error("Error should not be nil")
	}
}
