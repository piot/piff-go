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

import "fmt"

const FileFormatVersion = 0x01

type TypeID [4]byte

func (t TypeID) IsEqual(other [4]byte) bool {
	return t == other
}

func (t TypeID) IsEqualString(b string) bool {
	return t[0] == b[0] &&
		t[1] == b[1] &&
		t[2] == b[2] &&
		t[3] == b[3]
}

func NewTypeIDFromOctets(payload []byte) (TypeID, error) {
	if len(payload) != 4 {
		return TypeID{}, fmt.Errorf("typeid: payload must be exactly four octets.")
	}

	return TypeID{
		payload[0],
		payload[1],
		payload[2],
		payload[3],
	}, nil
}
