package fsutil

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type ReadOnlyFile struct {
	fileName string
	f        filesys.File
}

func (f *ReadOnlyFile) Read(p []byte) (n int, err error) {
	return f.f.Read(p)
}

func (f *ReadOnlyFile) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("readonly fs: %s", f.fileName)
}

func (f *ReadOnlyFile) Close() error {
	return f.f.Close()
}

func (f *ReadOnlyFile) Stat() (os.FileInfo, error) {
	return f.f.Stat()
}
