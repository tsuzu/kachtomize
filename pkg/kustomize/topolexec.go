package kustomize

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/tsuzu/kachtomize/pkg/sets"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type TopoloigcalExecutor struct {
	nodes    map[string]*topoNode
	options  []string
	numOfCPU int

	leftCount atomic.Int32
	wg        sync.WaitGroup
	errCount  atomic.Int32
	ch        chan string
	aborted   chan struct{}

	ErrorChan chan error
}

type topoNode struct {
	kustomizePath string
	depenededBy   []string
	dependencies  *sets.SyncSet[string]
}

func NewTopologicalExecutor(nodes []Node, options []string, numOfCPU int) *TopoloigcalExecutor {
	te := &TopoloigcalExecutor{
		nodes:     map[string]*topoNode{},
		options:   options,
		numOfCPU:  numOfCPU,
		aborted:   make(chan struct{}),
		ErrorChan: make(chan error, 1),
	}

	noDepNodes := make([]string, 0, 128)
	for _, node := range nodes {
		te.nodes[node.AbsDirPath] = &topoNode{
			kustomizePath: node.KustomizePath,
			depenededBy:   node.DependedBy,
			dependencies:  sets.FromSlice(node.Dependencies),
		}

		if len(node.Dependencies) == 0 {
			noDepNodes = append(noDepNodes, node.AbsDirPath)
		}
	}

	te.ch = make(chan string, len(noDepNodes))
	for i := range noDepNodes {
		te.ch <- noDepNodes[i]
	}
	te.leftCount.Store(int32(len(nodes) - len(noDepNodes)))

	return te
}

func (te *TopoloigcalExecutor) Run() error {
	te.startWorker()

	te.wg.Wait()

	close(te.ErrorChan)

	if errCount := te.errCount.Load(); errCount != 0 {
		return fmt.Errorf("%d errors occurred", errCount)
	}

	return nil
}

func (te *TopoloigcalExecutor) abort() {
	defer func() {
		recover()
	}()

	close(te.aborted)
}

func (te *TopoloigcalExecutor) startWorker() {
	for i := 0; i < te.numOfCPU; i++ {
		te.wg.Add(1)

		go te.worker(te.ch)
	}
}

func (te *TopoloigcalExecutor) worker(ch <-chan string) {
	defer te.wg.Done()

	for {
		var c string
		select {
		case c = <-ch:
		case <-te.aborted:
			return
		}

		err := te.processKustomize(c)

		if err != nil {
			te.errCount.Add(1)
			te.ErrorChan <- err
			te.abort()
		}
	}
}

func (te *TopoloigcalExecutor) processKustomize(dir string) error {
	built, err := te.build(dir)

	if err != nil {
		return fmt.Errorf("kustomize build failed: %w", err)
	}

	err = te.replaceKustomize(dir, te.nodes[dir].kustomizePath, built)

	if err != nil {
		return fmt.Errorf("failed to replace kustomization.yaml in %s: %w", dir, err)
	}

	go func() {
		for _, pending := range te.nodes[dir].depenededBy {
			pendingNode := te.nodes[pending]

			if pendingNode.dependencies.Delete(dir) == 0 {
				select {
				case te.ch <- pending:
				case <-te.aborted:
					return
				}

				if te.leftCount.Add(-1) == 0 {
					close(te.ch)
				}
			}
		}
	}()

	return nil
}

func (te *TopoloigcalExecutor) build(dir string) (string, error) {
	tempDir, err := os.MkdirTemp("", "kachtomize-*")

	if err != nil {
		return "", fmt.Errorf("failed to create a temporary directory for %s: %w", dir, err)
	}

	err = os.WriteFile(filepath.Join(tempDir, "kustomization.yaml"), []byte(`resources: ["resource.yaml"]`), 0666)

	if err != nil {
		return "", fmt.Errorf("failed to save a temporary kustomization.yaml for %s: %w", dir, err)
	}

	fp, err := os.Create(filepath.Join(tempDir, "resource.yaml"))

	if err != nil {
		return "", fmt.Errorf("failed to create a temporary file for %s: %w", dir, err)
	}
	defer fp.Close()

	stderrBuffer := bytes.Buffer{}

	cmd := exec.Command("kustomize", append([]string{"build"}, te.options...)...)
	cmd.Dir = dir
	cmd.Stdout = fp
	cmd.Stderr = &stderrBuffer

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("processing %s failed(%v): %s", dir, err, stderrBuffer.String())
	}

	return fp.Name(), nil
}

func (te *TopoloigcalExecutor) replaceKustomize(dir, kustomizeFile, builtPath string) error {
	fp, err := os.Create(kustomizeFile)

	if err != nil {
		return fmt.Errorf("failed to open for writing %s: %w", kustomizeFile, err)
	}
	defer fp.Close()

	relPath, err := filepath.Rel(dir, builtPath)

	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}

	k := types.Kustomization{
		TypeMeta: types.TypeMeta{
			Kind:       "Kustomization",
			APIVersion: "kustomize.config.k8s.io/v1beta1",
		},
		Resources: []string{
			relPath,
		},
	}

	if err := yaml.NewEncoder(fp).Encode(k); err != nil {
		return fmt.Errorf("failed to save Kustomization in %s: %w", kustomizeFile, err)
	}

	return nil
}
