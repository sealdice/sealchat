package api

import (
	"io"
	"mime/multipart"
	"sync"

	"github.com/samber/lo"
	"github.com/spf13/afero"
	"golang.org/x/crypto/blake2s"
)

var copyBufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 4096)
	},
}

func copyZeroAlloc(w io.Writer, r io.Reader) (int64, error) {
	vbuf := copyBufPool.Get()
	buf := vbuf.([]byte)
	n, err := io.CopyBuffer(w, r, buf)
	copyBufPool.Put(vbuf)
	return n, err
}

func SaveMultipartFile(fh *multipart.FileHeader, fOut afero.File, limit int64) (hashOut []byte, err error) {
	var (
		f multipart.File
	)
	f, err = fh.Open()
	if err != nil {
		return
	}

	defer func() {
		e := f.Close()
		if err == nil {
			err = e
		}
	}()

	limitReader := io.LimitReader(f, limit)
	hash := lo.Must(blake2s.New256(nil))
	teeReader := io.TeeReader(limitReader, hash)

	_, err = copyZeroAlloc(fOut, teeReader)
	hashOut = hash.Sum(nil)
	return
}
