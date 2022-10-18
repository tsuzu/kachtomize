package pathutil

import (
	"fmt"
	"path/filepath"

	"golang.org/x/sync/errgroup"
)

// CleanFilepath evals a symlink and returns it as an absolute path
func CleanFilepath(target string) (string, error) {
	evaled, err := filepath.EvalSymlinks(target)

	if err != nil {
		return "", fmt.Errorf("failed to eval symlink for %s: %w", target, err)
	}

	absPath, err := filepath.Abs(evaled)

	if err != nil {
		return "", fmt.Errorf("failed to get abs path for %s evaled from %s: %w", evaled, target, err)
	}

	return absPath, nil
}

// CleanFilepaths evals symlinks and returns them as absolute paths
func CleanFilepaths(targets []string, numOfCPUs int) ([]string, error) {
	var wg errgroup.Group
	ch := make(chan int, 1)

	go func() {
		defer close(ch)

		for c := range targets {
			ch <- c
		}
	}()

	results := make([]string, len(targets))
	for i := 0; i < numOfCPUs; i++ {
		wg.Go(func() error {
			for idx := range ch {
				cleaned, err := CleanFilepath(targets[idx])

				if err != nil {
					return fmt.Errorf("failed to clean %s: %w", targets[idx], err)
				}

				results[idx] = cleaned
			}

			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}
