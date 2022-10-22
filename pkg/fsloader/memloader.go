package fsloader

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/sync/errgroup"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type Loader struct {
	fSys filesys.FileSystem
	lock sync.Mutex
}

func New(fSys filesys.FileSystem) *Loader {
	return &Loader{
		fSys: fSys,
	}
}

func (l *Loader) LoadAll(dirs []string, numOfCPU int) error {
	ch := make(chan string, 1)

	ctx, cancel := context.WithCancel(context.Background())
	var wg errgroup.Group
	for i := 0; i < numOfCPU; i++ {
		wg.Go(func() error {
			defer cancel()

			var dir string
			var ok bool
			for {
				select {
				case <-ctx.Done():
					return nil
				case dir, ok = <-ch:
					if !ok {
						return nil
					}
				}

				if err := l.Load(dir); err != nil {
					return err
				}
			}
		})
	}

	for _, d := range dirs {
		ch <- d
	}
	close(ch)

	if err := wg.Wait(); err != nil {
		return fmt.Errorf("something went wrong: %w", err)
	}

	return nil
}

func (l *Loader) Load(dir string) error {
	abs, err := filepath.Abs(dir)

	if err != nil {
		return fmt.Errorf("failed to get absolute path for %s: %w", dir, err)
	}

	err = filepath.WalkDir(abs, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			l.lock.Lock()
			l.fSys.MkdirAll(path)
			l.lock.Unlock()

			return nil
		}

		b, err := os.ReadFile(path)

		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		l.lock.Lock()
		l.fSys.WriteFile(path, b)
		l.lock.Unlock()

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk %s: %w", dir, err)
	}

	return nil
}
