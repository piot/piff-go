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

package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/piot/piff-go/src/piff"

	"github.com/fatih/color"

	"encoding/base64"

	"github.com/piot/log-go/src/clog"
)

func options() string {
	//var piffFile string
	//	flag.StringVar(&piffFile, "filename", "", "file to view")
	flag.Parse()
	count := flag.NArg()
	if count < 1 {
		return ""
	}
	ibdfFilename := flag.Arg(0)
	return ibdfFilename
}

func openReadSeeker(filename string) (io.ReadSeeker, error) {
	var seekerToUse io.ReadSeeker
	if filename == "" {
		seekerToUse = os.Stdin
	} else {
		newFile, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		seekerToUse = newFile
	}

	return seekerToUse, nil
}

func run(filename string, log *clog.Log) error {
	seekerToUse, seekerErr := openReadSeeker(filename)
	if seekerErr != nil {
		return seekerErr
	}

	inFile, err := piff.NewInStreamReadSeeker(seekerToUse)
	if err != nil {
		return err
	}

	for {
		header, payload, readErr := inFile.ReadChunk()
		if readErr == io.EOF {
			break
		}
		fmt.Printf("-- %v: octetCount:%v index:%v\n", header.TypeIDString(), header.OctetCount(), header.ChunkIndex())
		color.Cyan("%v\n", hex.Dump(payload))
		base64String := base64.StdEncoding.EncodeToString(payload)
		color.Blue("%v\n", base64String)
	}

	return nil
}

func main() {
	log := clog.DefaultLog()
	log.Info("Piff viewer")
	filename := options()
	err := run(filename, log)
	if err != nil {
		log.Err(err)
		os.Exit(1)
	}
	log.Info("Done!")
}
