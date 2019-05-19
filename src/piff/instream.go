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
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/piot/brook-go/src/instream"
)

type InStream struct {
	inStream      io.ReadSeeker
	pendingHeader InHeader
	isEOF         bool
	seekHeaders   []InSeekHeader
}

func NewInStreamFile(filename string) (*InStream, error) {
	newFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return NewInStreamReadSeeker(newFile)
}

func verifyFileHeader(reader io.Reader) error {
	expectedHeader := fileFormatHeaderWithVersion(FileFormatVersion)
	fileHeaderPayload := make([]byte, len(expectedHeader))
	_, readErr := reader.Read(fileHeaderPayload)
	if readErr != nil {
		return fmt.Errorf("couldnt read file pendingHeader %v", readErr)
	}
	if !bytes.Equal(fileHeaderPayload, expectedHeader) {
		return fmt.Errorf("unexpected file format pendingHeader")
	}

	return nil
}

func NewInStreamReadSeeker(inStream io.ReadSeeker) (*InStream, error) {
	c := &InStream{
		inStream: inStream,
	}
	fileHeaderErr := verifyFileHeader(inStream)
	if fileHeaderErr != nil {
		return nil, fileHeaderErr
	}
	headerErr := c.readHeader()
	return c, headerErr
}

func (c *InStream) internalRead(requestedOctetCount int) ([]byte, error) {
	payload := make([]byte, requestedOctetCount)
	_, err := c.inStream.Read(payload)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func (c *InStream) readHeaderInternal() (InHeader, error) {
	pendingHeader := make([]byte, 8)
	tell, tellErr := c.inStream.Seek(0, 1)
	if tellErr != nil {
		return InHeader{}, tellErr
	}
	octetCount, err := c.inStream.Read(pendingHeader)
	if err == io.EOF {
		return InHeader{}, io.EOF
	}
	if err != nil {
		return InHeader{}, err
	}
	if octetCount != 8 {
		return InHeader{}, fmt.Errorf("piff: couldn't read whole pendingHeader")
	}
	s := instream.New(pendingHeader)

	typeID, readErr := s.ReadOctets(4)
	if readErr != nil {
		return InHeader{}, readErr
	}
	fourCC, fourCCErr := NewTypeIDFromOctets(typeID)
	if fourCCErr != nil {
		return InHeader{}, fourCCErr
	}
	chunkOctetCount, countErr := s.ReadUint32()
	if countErr != nil {
		return InHeader{}, countErr
	}
	return InHeader{octetLength: int(chunkOctetCount), typeID: fourCC, tell: tell}, nil
}

func (c *InStream) readHeader() error {
	var err error
	c.pendingHeader, err = c.readHeaderInternal()
	if err == io.EOF {
		c.isEOF = true
		err = nil
	}

	if err != nil {
		return err
	}
	return nil
}

func (c *InStream) internalReadChunk(requestedOctetCount int) (InHeader, []byte, error) {
	if c.isEOF {
		return InHeader{}, nil, io.EOF
	}
	if requestedOctetCount > c.pendingHeader.octetLength {
		return InHeader{}, nil, fmt.Errorf("trying to read too much")
	}
	skipCount := c.pendingHeader.octetLength - requestedOctetCount
	payload := make([]byte, requestedOctetCount)
	octetsRead, err := c.inStream.Read(payload)
	if err != nil {
		return InHeader{}, nil, err
	}
	if skipCount > 0 {
		c.inStream.Seek(int64(skipCount), 1)
	}
	savedHeader := c.pendingHeader

	if octetsRead+skipCount != savedHeader.octetLength {
		return InHeader{}, nil, fmt.Errorf("couldnt read the whole payload read:%v, skip:%v expected:%v", octetsRead, skipCount, savedHeader.octetLength)
	}
	headerErr := c.readHeader()
	return savedHeader, payload, headerErr
}

func (c *InStream) ReadChunk() (InHeader, []byte, error) {
	return c.internalReadChunk(c.pendingHeader.OctetCount())
}

func (c *InStream) ReadPartChunk(requestedOctetCount int) (InHeader, []byte, error) {
	return c.internalReadChunk(requestedOctetCount)
}

func (c *InStream) SkipChunk() (InHeader, error) {
	if c.isEOF {
		return InHeader{}, io.EOF
	}
	savedHeader := c.pendingHeader
	c.inStream.Seek(int64(savedHeader.OctetCount()), 1)
	headerErr := c.readHeader()
	return savedHeader, headerErr
}

func (c *InStream) Close() {
	//c.inStream.Close()
}
