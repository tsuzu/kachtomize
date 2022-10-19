package kustomize

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/tsuzu/kachtomize/pkg/pathutil"
	"github.com/tsuzu/kachtomize/pkg/sets"
	"sigs.k8s.io/kustomize/api/types"
)

type DependencyAnalyzer struct {
	kustomizes   []string
	ignoreErrors bool
	numOfCPU     int

	wg       sync.WaitGroup
	nodes    map[string]*node
	errCount atomic.Int32

	ErrorChan chan error
}

type Node struct {
	AbsDirPath    string
	KustomizePath string
	DependedBy    []string
	Dependencies  []string
}

type node struct {
	invalid       bool
	kustomizePath string
	depenededBy   *sets.SyncSet[string]
	dependencies  []string
}

func NewDependencyAnalyzer(kustomizes []string, ignoreErrors bool, numOfCPU int) *DependencyAnalyzer {
	da := DependencyAnalyzer{
		kustomizes:   kustomizes,
		ignoreErrors: ignoreErrors,
		numOfCPU:     numOfCPU,
		nodes:        map[string]*node{},
		ErrorChan:    make(chan error, 1),
	}

	return &da
}

func (da *DependencyAnalyzer) Run() ([]Node, error) {
	var err error

	da.kustomizes, err = pathutil.CleanFilepaths(da.kustomizes, da.numOfCPU)

	if err != nil {
		return nil, fmt.Errorf("faield to clean filepaths: %w", err)
	}

	for _, k := range da.kustomizes {
		dir := filepath.Dir(k)

		da.nodes[dir] = &node{
			kustomizePath: k,
			depenededBy:   sets.NewSyncSet[string](),
		}
	}

	da.startWorker()

	da.wg.Wait()
	close(da.ErrorChan)

	if errCount := da.errCount.Load(); errCount != 0 {
		return nil, fmt.Errorf("%d errors occurred", errCount)
	}

	return da.generateNodes(), nil
}

func (da *DependencyAnalyzer) startWorker() {
	ch := make(chan string, 1)

	go func() {
		defer close(ch)

		for _, k := range da.kustomizes {
			ch <- k
		}
	}()

	for i := 0; i < da.numOfCPU; i++ {
		da.wg.Add(1)

		go da.worker(ch)
	}
}

func (da *DependencyAnalyzer) worker(ch <-chan string) {
	defer da.wg.Done()

	for c := range ch {
		err := da.processKustomize(c)

		if err != nil {
			da.nodes[filepath.Dir(c)].invalid = true

			if !da.ignoreErrors {
				da.errCount.Add(1)
			}

			da.ErrorChan <- err
		}
	}
}

func hasRemoteFileScheme(path string) bool {
	u, err := url.Parse(path)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https")
}

func (da *DependencyAnalyzer) processKustomize(file string) error {
	content, err := os.ReadFile(file)

	if err != nil {
		return fmt.Errorf("failed to open %s: %w", file, err)
	}

	content, err = types.FixKustomizationPreUnmarshalling(content)

	if err != nil {
		return fmt.Errorf("failed to fix kustomization.yaml %s: %w", file, err)
	}

	var k types.Kustomization

	if err := k.Unmarshal(content); err != nil {
		return fmt.Errorf("failed to unmarshal %s: %w", file, err)
	}
	k.FixKustomizationPostUnmarshalling()

	dir := filepath.Dir(file)
	selfNode := da.nodes[dir]

	if k.Kind != "Kustomization" {
		selfNode.invalid = true

		return nil
	}

	for _, d := range k.Resources {
		if hasRemoteFileScheme(d) {
			continue
		}

		d = filepath.Join(dir, d)

		cleaned, err := pathutil.CleanFilepath(d)

		if err != nil {
			return fmt.Errorf("failed to clean %s: %w", d, err)
		}

		depNode, ok := da.nodes[cleaned]

		if !ok {
			continue
		}

		depNode.depenededBy.Add(dir)
		selfNode.dependencies = append(selfNode.dependencies, cleaned)
	}

	return nil
}

func (da *DependencyAnalyzer) generateNodes() []Node {
	nodes := make([]Node, 0, len(da.nodes))

	for key, value := range da.nodes {
		if value.invalid {
			continue
		}

		nodes = append(nodes, Node{
			AbsDirPath:    key,
			KustomizePath: value.kustomizePath,
			DependedBy:    value.depenededBy.All(),
			Dependencies:  value.dependencies,
		})
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].AbsDirPath < nodes[j].AbsDirPath
	})

	return nodes
}
