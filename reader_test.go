// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ccdv

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestRead(t *testing.T) {
	tests := []struct {
		Name   string
		Input  string
		Output [][]string
		Error  error

		// These fields are copied into the Reader
		Comma              rune
		Comment            rune
		UseFieldsPerRecord bool // false (default) means FieldsPerRecord is -1
		FieldsPerRecord    int
		LazyQuotes         bool
		TrimLeadingSpace   bool
		ReuseRecord        bool
	}{{
		Name:   "Simple",
		Input:  "a\037b\037c\036",
		Output: [][]string{{"a", "b", "c"}},
	}, {
		Name:   "CRLF",
		Input:  "a\037b\036c\037d\036",
		Output: [][]string{{"a", "b"}, {"c", "d"}},
	}, {
		Name:   "BareCR",
		Input:  "a\037b\rc\037d\036",
		Output: [][]string{{"a", "b\rc", "d"}},
	}, {
		Name:  "RFC4180test",
		Input: "#field1\037field2\037field3\036aaa\037bb\nb\037ccc\036a,a\037b\"bb\037ccc\036zzz\037yyy\037xxx\036",
		Output: [][]string{
			{"#field1", "field2", "field3"},
			{"aaa", "bb\nb", "ccc"},
			{"a,a", `b"bb`, "ccc"},
			{"zzz", "yyy", "xxx"},
		},
		UseFieldsPerRecord: true,
		FieldsPerRecord:    0,
	}, {
		Name:   "NoEOLTest",
		Input:  "a\037b\037c",
		Output: [][]string{{"a", "b", "c"}},
	}, {
		Name:   "MultiLine",
		Input:  "two\nline\037one line\037three\nline\nfield\036",
		Output: [][]string{{"two\nline", "one line", "three\nline\nfield"}},
	}, {
		Name:  "BlankLine",
		Input: "a\037b\037c\036\036d\037e\037f\036\036",
		Output: [][]string{
			{"a", "b", "c"},
			{"d", "e", "f"},
		},
	}, {
		Name:  "BlankLineFieldCount",
		Input: "a\037b\037c\036\036d\037e\037f\036\036",
		Output: [][]string{
			{"a", "b", "c"},
			{"d", "e", "f"},
		},
		UseFieldsPerRecord: true,
		FieldsPerRecord:    0,
	}, {
		Name:             "TrimSpace",
		Input:            " a\037  b\037   c\036",
		Output:           [][]string{{"a", "b", "c"}},
		TrimLeadingSpace: true,
	}, {
		Name:   "LeadingSpace",
		Input:  " a\037  b\037   c\036",
		Output: [][]string{{" a", "  b", "   c"}},
	}, {
		Name:    "Comment",
		Input:   "\0201\0372\0373\036a\037b\037c\036\020comment\036",
		Output:  [][]string{{"a", "b", "c"}},
		Comment: '#',
	}, {
		Name:   "NoComment",
		Input:  "#1\0372\0373\036a\037b\037c\036",
		Output: [][]string{{"#1", "2", "3"}, {"a", "b", "c"}},
	}, {
		Name:               "BadFieldCount",
		Input:              "a\037b\037c\036d\037e\036",
		Error:              &ParseError{StartLine: 2, Line: 2, Err: ErrFieldCount},
		UseFieldsPerRecord: true,
		FieldsPerRecord:    0,
	}, {
		Name:               "BadFieldCount1",
		Input:              "a\036b\036c\036",
		Error:              &ParseError{StartLine: 1, Line: 1, Err: ErrFieldCount},
		UseFieldsPerRecord: true,
		FieldsPerRecord:    2,
	}, {
		Name:   "FieldCount",
		Input:  "a\037b\037c\036d\037e\036",
		Output: [][]string{{"a", "b", "c"}, {"d", "e"}},
	}, {
		Name:  "CommaFieldTest",
		Input: "x\037y\037z\037w\036x\037y\037z\037\036x\037y\037\037\036x\037\037\037\036\037\037\037\036x\037y\037z\037w\036x\037y\037z\037\036x\037y\037\037\036x\037\037\037\036\037\037\037\036",
		Output: [][]string{
			{"x", "y", "z", "w"},
			{"x", "y", "z", ""},
			{"x", "y", "", ""},
			{"x", "", "", ""},
			{"", "", "", ""},
			{"x", "y", "z", "w"},
			{"x", "y", "z", ""},
			{"x", "y", "", ""},
			{"x", "", "", ""},
			{"", "", "", ""},
		},
	}, {
		Name:  "TrailingCommaIneffective1",
		Input: "a\037b\037\036c\037d\037e",
		Output: [][]string{
			{"a", "b", ""},
			{"c", "d", "e"},
		},
		TrimLeadingSpace: true,
	}, {
		Name:  "ReadAllReuseRecord",
		Input: "a\037b\036c\037d",
		Output: [][]string{
			{"a", "b"},
			{"c", "d"},
		},
		ReuseRecord: true,
	}, {
		Name:   "BinaryBlobField", // Issue 19410
		Input:  "x09\x41\xb4\x1c\037aktau",
		Output: [][]string{{"x09A\xb4\x1c", "aktau"}},
	}}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			r := NewReader(strings.NewReader(tt.Input))

			if tt.UseFieldsPerRecord {
				r.FieldsPerRecord = tt.FieldsPerRecord
			} else {
				r.FieldsPerRecord = -1
			}
			r.TrimLeadingSpace = tt.TrimLeadingSpace
			r.ReuseRecord = tt.ReuseRecord

			out, err := r.ReadAll()
			if !reflect.DeepEqual(err, tt.Error) {
				t.Errorf("ReadAll() error:\ngot  %v\nwant %v", err, tt.Error)
			} else if !reflect.DeepEqual(out, tt.Output) {
				t.Errorf("ReadAll() output:\ngot  %q\nwant %q", out, tt.Output)
			}
		})
	}
}

// nTimes is an io.Reader which yields the string s n times.
type nTimes struct {
	s   string
	n   int
	off int
}

func (r *nTimes) Read(p []byte) (n int, err error) {
	for {
		if r.n <= 0 || r.s == "" {
			return n, io.EOF
		}
		n0 := copy(p, r.s[r.off:])
		p = p[n0:]
		n += n0
		r.off += n0
		if r.off == len(r.s) {
			r.off = 0
			r.n--
		}
		if len(p) == 0 {
			return
		}
	}
}

// benchmarkRead measures reading the provided CSV rows data.
// initReader, if non-nil, modifies the Reader before it's used.
func benchmarkRead(b *testing.B, initReader func(*Reader), rows string) {
	b.ReportAllocs()
	r := NewReader(&nTimes{s: rows, n: b.N})
	if initReader != nil {
		initReader(r)
	}
	for {
		_, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			b.Fatal(err)
		}
	}
}

const benchmarkCSVData = "x\037y\037z\037w\036x\037y\037z\037\036x\037y\037\037\036x\037\037\037\036\037\037\037\036x\037y\037z\037w\036x\037y\037z\037\036x\037y\037\037\036x\037\037\037\036\037\037\037\036"

func BenchmarkRead(b *testing.B) {
	benchmarkRead(b, nil, benchmarkCSVData)
}

func BenchmarkReadWithFieldsPerRecord(b *testing.B) {
	benchmarkRead(b, func(r *Reader) { r.FieldsPerRecord = 4 }, benchmarkCSVData)
}

func BenchmarkReadWithoutFieldsPerRecord(b *testing.B) {
	benchmarkRead(b, func(r *Reader) { r.FieldsPerRecord = -1 }, benchmarkCSVData)
}

func BenchmarkReadLargeFields(b *testing.B) {
	benchmarkRead(b, nil, strings.Repeat("xxxxxxxxxxxxxxxx\037yyyyyyyyyyyyyyyy\037zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz\037wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww\037vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv\036xxxxxxxxxxxxxxxxxxxxxxxx\037yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy\037zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz\037wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww\037vvvv\036\037\037zzzz\037wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww\037vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv\036xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\037yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy\037zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz\037wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww\037vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv\036", 3))
}

func BenchmarkReadReuseRecord(b *testing.B) {
	benchmarkRead(b, func(r *Reader) { r.ReuseRecord = true }, benchmarkCSVData)
}

func BenchmarkReadReuseRecordWithFieldsPerRecord(b *testing.B) {
	benchmarkRead(b, func(r *Reader) { r.ReuseRecord = true; r.FieldsPerRecord = 4 }, benchmarkCSVData)
}

func BenchmarkReadReuseRecordWithoutFieldsPerRecord(b *testing.B) {
	benchmarkRead(b, func(r *Reader) { r.ReuseRecord = true; r.FieldsPerRecord = -1 }, benchmarkCSVData)
}

func BenchmarkReadReuseRecordLargeFields(b *testing.B) {
	benchmarkRead(b, func(r *Reader) { r.ReuseRecord = true }, strings.Repeat("xxxxxxxxxxxxxxxx\037yyyyyyyyyyyyyyyy\037zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz\037wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww\037vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv\036xxxxxxxxxxxxxxxxxxxxxxxx\037yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy\037zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz\037wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww\037vvvv\036\037\037zzzz\037wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww\037vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv\036xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\037yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy\037zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz\037wwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwwww\037vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv\036", 3))
}
