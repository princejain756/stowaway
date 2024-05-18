package link

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
)

func readlinkPath(target string) (string, error) {
	dest, err := os.Readlink(target)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(dest) {
		dest = filepath.Join(filepath.Dir(target), dest)
	}
	return dest, nil
}

func sameContents(a, b string) (bool, error) {
	ai, err := os.Stat(a)
	if err != nil {
		return false, err
	}
	bi, err := os.Stat(b)
	if err != nil {
		return false, err
	}
	if ai.IsDir() || bi.IsDir() || ai.Size() != bi.Size() {
		return false, nil
	}
	af, err := os.Open(a)
	if err != nil {
		return false, err
	}
	defer af.Close()
	bf, err := os.Open(b)
	if err != nil {
		return false, err
	}
	defer bf.Close()
	abuf := make([]byte, 32*1024)
	bbuf := make([]byte, 32*1024)
	for {
		an, aerr := io.ReadFull(af, abuf)
		bn, berr := io.ReadFull(bf, bbuf)
		if an != bn || !bytes.Equal(abuf[:an], bbuf[:bn]) {
			return false, nil
		}
		if aerr == io.EOF || aerr == io.ErrUnexpectedEOF {
			return berr == io.EOF || berr == io.ErrUnexpectedEOF, nil
		}
		if aerr != nil {
			return false, aerr
		}
		if berr != nil {
			return false, berr
		}
	}
}
