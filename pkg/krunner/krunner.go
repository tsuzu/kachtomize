package krunner

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type result struct {
	dir  string
	yaml []byte
}

type Runner struct {
	kustomizerInit func() *krusty.Kustomizer
	fSys           filesys.FileSystem
	callback       func(dir string, data []byte)

	callbackWg sync.WaitGroup
	requestCh  chan string
	resultCh   chan result
	errCounter atomic.Int32
}

func New(kustomizerInit func() *krusty.Kustomizer, fSys filesys.FileSystem, numOfCPU int) *Runner {
	r := &Runner{
		kustomizerInit: kustomizerInit,
		fSys:           fSys,
		requestCh:      make(chan string, 1),
		resultCh:       make(chan result, 1),
	}
	go r.startWorker(numOfCPU)

	return r
}

func (r *Runner) startWorker(numOfCPU int) {
	r.callbackWg.Add(1)
	go r.callCallbackWorker()

	var wg sync.WaitGroup

	wg.Add(numOfCPU)
	for i := 0; i < numOfCPU; i++ {
		go func() {
			defer wg.Done()
			r.worker()
		}()
	}

	wg.Wait()
	close(r.resultCh)
}

func (r *Runner) worker() {
	for dir := range r.requestCh {
		if err := r.runKustomize(dir); err != nil {
			r.errCounter.Add(1)
			log.Println(err)
		}
	}
}

func (r *Runner) runKustomize(dir string) error {
	kustomizer := r.kustomizerInit()

	resMap, err := kustomizer.Run(r.fSys, dir)

	if err != nil {
		return fmt.Errorf("kustomize for %s failed: %w", dir, err)
	}

	b, err := resMap.AsYaml()

	if err != nil {
		return fmt.Errorf("fetching YAML for %s failed: %w", dir, err)
	}

	r.resultCh <- result{
		dir:  dir,
		yaml: b,
	}

	return nil
}

func (r *Runner) callCallbackWorker() {
	defer r.callbackWg.Done()

	for res := range r.resultCh {
		r.callback(res.dir, res.yaml)
	}
}

func (r *Runner) RegisterCallback(fn func(dir string, data []byte)) {
	r.callback = fn
}

func (r *Runner) Enqueue(dir string) {
	r.requestCh <- dir
}

func (r *Runner) Wait() bool {
	close(r.requestCh)

	r.callbackWg.Wait()

	return r.errCounter.Load() == 0
}
