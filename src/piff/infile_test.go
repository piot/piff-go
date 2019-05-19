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
	"fmt"
	"testing"
)

func TestFetchChunks(t *testing.T) {
	const ibdFilename = "test2.ibdf"
	const typeID = "cafe"
	f, outErr := NewOutStream(ibdFilename)
	if outErr != nil {
		t.Fatal(outErr)
	}
	const chunksToWrite = 10

	for i := 0; i < chunksToWrite; i++ {
		testString := fmt.Sprintf("%02d:chunk", i)
		writeErr := f.WriteChunkTypeIDString(typeID, []byte(testString))
		if writeErr != nil {
			t.Fatal(writeErr)
		}
	}
	f.Close()

	i, err := NewInSeekerFile(ibdFilename)
	if err != nil {
		t.Error(err)
	}
	count := i.ChunkCount()
	if count != chunksToWrite {
		t.Errorf("chunk count is wrong %d", count)
	}

	partial, payload, findErr := i.FindPartialChunk(3, 3)
	if findErr != nil {
		t.Error(findErr)
	}
	payloadString := string(payload)
	if payloadString != "03:" {
		t.Errorf("illegal payload %s", payloadString)
	}
	if partial.OctetCount() != 8 {
		t.Errorf("illegal length:%v", partial)
	}
	i.Close()
}
