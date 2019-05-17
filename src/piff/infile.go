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
	"io"
	"os"

	"github.com/piot/brook-go/src/instream"
)

func TypeIDEqual(a [4]byte, b [4]byte) bool {
	return a == b
}

func TypeIDEqualString(a [4]byte, b string) bool {
	return a[0] == b[0] &&
		a[1] == b[1] &&
		a[2] == b[2] &&
		a[3] == b[3]
}

type InHeader struct {
	typeID      [4]byte
	octetLength int
}

func (i InHeader) TypeIDString() string {
	return string(i.typeID[0:])
}

func (i InHeader) OctetCount() int {
	return i.octetLength
}

func (i InHeader) String() string {
	return fmt.Sprintf("header %v %v", i.TypeIDString(), i.OctetCount())
}

type InSeekHeader struct {
	header InHeader
	tell   int64
}

func (i InSeekHeader) Header() InHeader {
	return i.header
}

func (i InSeekHeader) Tell() int64 {
	return i.tell
}

type InFile struct {
	inFile      *os.File
	header      InHeader
	isEOF       bool
	seekHeaders []InSeekHeader
}

func NewInFile(filename string) (*InFile, error) {
	newFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	c := &InFile{
		inFile: newFile,
	}
	headerErr := c.readHeader()
	return c, headerErr
}

func (c *InFile) AllHeaders() []InSeekHeader {
	return c.seekHeaders
}

func (c *InFile) scanAllChunks() error {
	beforePosition, tellErr := c.inFile.Seek(0, 1)
	if tellErr != nil {
		return tellErr
	}

	var seekHeaders []InSeekHeader
	for {
		tell, tellErr := c.inFile.Seek(0, 1)
		if tellErr != nil {
			return tellErr
		}

		header, headerErr := c.readHeaderInternal()
		if headerErr != nil {
			return headerErr
		}
		if c.isEOF {
			break
		}
		seekHeader := InSeekHeader{tell: tell, header: header}
		seekHeaders = append(seekHeaders, seekHeader)
		c.inFile.Seek(int64(header.octetLength), 1)
	}
	c.seekHeaders = seekHeaders
	c.inFile.Seek(beforePosition, 0)
	return nil
}

func (c *InFile) ChunkCount() int {
	return len(c.seekHeaders)
}

func (c *InFile) seekToChunk(index int) error {
	if index >= len(c.seekHeaders) {
		return fmt.Errorf("I don't have that index %d", index)
	}

	seekHeader := c.seekHeaders[index]
	_, seekErr := c.inFile.Seek(seekHeader.tell, 0)
	if seekErr != nil {
		return seekErr
	}
	return nil
}

func (c *InFile) seekToChunkAndReadHeader(index int) (InHeader, error) {
	seekErr := c.seekToChunk(index)
	if seekErr != nil {
		return InHeader{}, seekErr
	}
	return c.readHeaderInternal()
}

func (c *InFile) internalRead(requestedOctetCount int) ([]byte, error) {
	payload := make([]byte, requestedOctetCount)
	_, err := c.inFile.Read(payload)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func (c *InFile) FindChunk(index int) (InHeader, []byte, error) {
	header, headerErr := c.seekToChunkAndReadHeader(index)
	if headerErr != nil {
		return InHeader{}, nil, headerErr
	}
	payload, payloadErr := c.internalRead(header.octetLength)
	return header, payload, payloadErr
}

func (c *InFile) FindPartialChunk(index int, octetCount int) (InHeader, []byte, error) {
	header, headerErr := c.seekToChunkAndReadHeader(index)
	if headerErr != nil {
		return InHeader{}, nil, headerErr
	}
	payload, payloadErr := c.internalRead(octetCount)
	return header, payload, payloadErr
}

func NewInFileScanChunks(filename string) (*InFile, error) {
	inFile, err := NewInFile(filename)
	if err != nil {
		return nil, err
	}

	scanErr := inFile.scanAllChunks()
	if scanErr != nil {
		return nil, scanErr
	}

	return inFile, nil
}

func (c *InFile) readHeaderInternal() (InHeader, error) {
	header := make([]byte, 8)
	octetCount, err := c.inFile.Read(header)

	if err == io.EOF {
		c.isEOF = true
		return InHeader{}, nil
	}
	if err != nil {
		return InHeader{}, err
	}
	if octetCount != 8 {
		return InHeader{}, fmt.Errorf("piff: couldn't read whole header")
	}
	s := instream.New(header)

	typeID, readErr := s.ReadOctets(4)
	if readErr != nil {
		return InHeader{}, readErr
	}
	fourCC := [4]byte{
		typeID[0],
		typeID[1],
		typeID[2],
		typeID[3],
	}
	chunkOctetCount, countErr := s.ReadUint32()
	if countErr != nil {
		return InHeader{}, countErr
	}
	return InHeader{octetLength: int(chunkOctetCount), typeID: fourCC}, nil
}

func (c *InFile) readHeader() error {
	var err error
	c.header, err = c.readHeaderInternal()
	if err != nil {
		return err
	}
	return nil
}

func (c *InFile) internalReadChunk(requestedOctetCount int) (InHeader, []byte, error) {
	if c.isEOF {
		return InHeader{}, nil, io.EOF
	}
	if requestedOctetCount > c.header.octetLength {
		return InHeader{}, nil, fmt.Errorf("trying to read too much")
	}
	skipCount := c.header.octetLength - requestedOctetCount
	payload := make([]byte, requestedOctetCount)
	octetsRead, err := c.inFile.Read(payload)
	if err != nil {
		return InHeader{}, nil, err
	}
	if skipCount > 0 {
		c.inFile.Seek(int64(skipCount), 1)
	}
	savedHeader := c.header

	if octetsRead+skipCount != savedHeader.octetLength {
		return InHeader{}, nil, fmt.Errorf("couldnt read the whole payload")
	}
	headerErr := c.readHeader()
	return savedHeader, payload, headerErr

}

func (c *InFile) ReadChunk() (InHeader, []byte, error) {
	return c.internalReadChunk(c.header.OctetCount())
}

func (c *InFile) SkipChunk() (InHeader, error) {
	savedHeader := c.header
	c.inFile.Seek(int64(savedHeader.OctetCount()), 1)
	headerErr := c.readHeader()
	return savedHeader, headerErr
}

func (c *InFile) ReadPartChunk(requestedOctetCount int) (InHeader, []byte, error) {
	return c.internalReadChunk(requestedOctetCount)
}

func (c *InFile) Close() {
	c.inFile.Close()
}
