package kustomize

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	ignore "github.com/sabhiram/go-gitignore"
)

func ListKustomizeTarget(dir string) ([]string, error) {
	results := []string{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		switch d.Name() {
		case "kustomization.yaml":
		case "kustomization.yml":
		case "Kustomization":

		default:
			return nil
		}

		results = append(results, path)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk dir %s: %w", dir, err)
	}

	sort.Strings(results)

	return results, nil
}

func FilterByIgnore(ignoreFile string, targets []string) ([]string, error) {
	ignore, err := ignore.CompileIgnoreFile(ignoreFile)

	if err != nil {
		if err == os.ErrNotExist {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to open %s: %w", ignoreFile, err)
	}

	results := make([]string, 0, len(targets))
	for _, t := range targets {
		if ignore.MatchesPath(t) {
			continue
		}

		results = append(results, t)
	}

	return results, nil
}
