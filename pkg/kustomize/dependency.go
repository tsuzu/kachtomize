package kustomize

import (
	"fmt"
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
	kustomizes []string
	numOfCPU   int

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
	kustomizePath string
	depenededBy   *sets.SyncSet[string]
	dependencies  []string
}

func NewDependencyAnalyzer(kustomizes []string, numOfCPU int) *DependencyAnalyzer {
	da := DependencyAnalyzer{
		kustomizes: kustomizes,
		numOfCPU:   numOfCPU,
		nodes:      map[string]*node{},
		ErrorChan:  make(chan error, 1),
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
			da.errCount.Add(1)
			da.ErrorChan <- err
		}
	}
}

func (da *DependencyAnalyzer) processKustomize(file string) error {
	b, err := os.ReadFile(file)

	if err != nil {
		return fmt.Errorf("failed to open %s: %w", file, err)
	}

	var k types.Kustomization

	if err := k.Unmarshal(b); err != nil {
		return fmt.Errorf("failed to unmarshal %s: %w", file, err)
	}

	// k.Resources is overwritten
	deps := append(k.Resources, k.Bases...)

	dir := filepath.Dir(file)
	selfNode := da.nodes[dir]
	for _, d := range deps {
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
