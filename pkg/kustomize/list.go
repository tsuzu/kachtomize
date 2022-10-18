package kustomize

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
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
