/*

MIT License

Copyright (c) 2019 Peter Bjorklund

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

*/

package piff

import (
	"io"
	"testing"
)

func TestReadWrite(t *testing.T) {
	const testString = "this is a string"
	const ibdFilename = "test.ibdf"
	const typeID = "cafe"
	f, outErr := NewOutFile(ibdFilename)
	if outErr != nil {
		t.Fatal(outErr)
	}
	writeErr := f.WriteChunkTypeIDString(typeID, []byte(testString))
	if writeErr != nil {
		t.Fatal(writeErr)
	}
	f.Close()

	i, _ := NewInFile(ibdFilename)

	header, payload, readErr := i.ReadChunk()

	if readErr != nil {
		t.Fatal(readErr)
	}
	if header.octetLength != len(testString) {
		t.Errorf("wrong octet length")
	}

	if header.typeID[1] != typeID[1] {
		t.Errorf("wrong typeid")
	}

	if string(payload) != testString {
		t.Errorf("wrong string")
	}
	_, _, nextReadErr := i.ReadChunk()
	if nextReadErr != io.EOF {
		t.Errorf("file should have ended")
	}
	i.Close()
}
