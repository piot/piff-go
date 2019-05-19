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
)

type InSeekHeader struct {
	header InHeader
}

func (i InSeekHeader) Header() InHeader {
	return i.header
}

func (i InSeekHeader) Tell() int64 {
	return i.header.tell
}

func (i InSeekHeader) String() string {
	return fmt.Sprintf("[inseekheader position:%v subheader:%v]", i.header.tell, i.header)
}

type InSeeker struct {
	inFile      *InStream
	seekHeaders []InSeekHeader
}

func NewInSeekerFile(filename string) (*InSeeker, error) {
	newFile, err := NewInStreamFile(filename)
	if err != nil {
		return nil, err
	}

	return newInSeekerStream(newFile)
}

func NewInSeeker(readSeeker io.ReadSeeker) (*InSeeker, error) {
	newFile, err := NewInStreamReadSeeker(readSeeker)
	if err != nil {
		return nil, err
	}

	return newInSeekerStream(newFile)
}

func newInSeekerStream(newFile *InStream) (*InSeeker, error) {
	c := &InSeeker{
		inFile: newFile,
	}
	scanErr := c.scanAllChunks()
	if scanErr != nil {
		return nil, scanErr
	}
	return c, nil
}

func (c *InSeeker) AllHeaders() []InSeekHeader {
	return c.seekHeaders
}

func (c *InSeeker) scanAllChunks() error {
	var seekHeaders []InSeekHeader
	for {
		header, headerErr := c.inFile.SkipChunk()
		if headerErr == io.EOF {
			break
		}
		if headerErr != nil {
			return headerErr
		}

		seekHeader := InSeekHeader{header: header}
		seekHeaders = append(seekHeaders, seekHeader)
	}
	c.seekHeaders = seekHeaders
	return nil
}

func (c *InSeeker) ChunkCount() int {
	return len(c.seekHeaders)
}

func (c *InSeeker) seekToChunk(index int) error {
	if index >= len(c.seekHeaders) {
		return fmt.Errorf("I don't have that index %d", index)
	}

	seekHeader := c.seekHeaders[index]
	_, seekErr := c.inFile.inStream.Seek(seekHeader.header.tell, 0)
	if seekErr != nil {
		return seekErr
	}
	return nil
}

func (c *InSeeker) seekToChunkAndReadHeader(index int) (InHeader, error) {
	seekErr := c.seekToChunk(index)
	if seekErr != nil {
		return InHeader{}, seekErr
	}
	return c.inFile.readHeaderInternal()
}

func (c *InSeeker) FindChunk(index int) (InHeader, []byte, error) {
	header, headerErr := c.seekToChunkAndReadHeader(index)
	if headerErr != nil {
		return InHeader{}, nil, headerErr
	}
	payload, payloadErr := c.inFile.internalRead(header.octetLength)
	return header, payload, payloadErr
}

func (c *InSeeker) FindPartialChunk(index int, octetCount int) (InHeader, []byte, error) {
	header, headerErr := c.seekToChunkAndReadHeader(index)
	if headerErr != nil {
		return InHeader{}, nil, headerErr
	}
	payload, payloadErr := c.inFile.internalRead(octetCount)
	return header, payload, payloadErr
}

func (c *InSeeker) Close() {
	c.inFile.Close()
}
