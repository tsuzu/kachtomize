package main

import (
	"flag"
	"log"
	"runtime"
	"sync"

	"github.com/tsuzu/kachtomize/pkg/kustomize"
)

var (
	dir              string
	kustomizeOptions []string
)

func init() {
	flag.Parse()

	dir = flag.Arg(0)

	if flag.NArg() >= 2 {
		if flag.Arg(1) != "--" {
			log.Fatal("Unknown flag ", flag.Arg(1))
		}

		kustomizeOptions = flag.Args()[2:]
	}
}

func main() {
	targets, err := kustomize.ListKustomizeTarget(dir)

	if err != nil {
		log.Fatal(err)
	}

	analyzer := kustomize.NewDependencyAnalyzer(targets, runtime.GOMAXPROCS(0))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for err := range analyzer.ErrorChan {
			if err != nil {
				log.Printf("analyzer failed: %+v", err)
			}
		}
	}()

	nodes, err := analyzer.Run()
	wg.Wait()

	if err != nil {
		log.Fatal(err)
	}

	e := kustomize.NewTopologicalExecutor(nodes, kustomizeOptions, runtime.GOMAXPROCS(0))

	wg.Add(1)
	go func() {
		defer wg.Done()
		for err := range e.ErrorChan {
			if err != nil {
				log.Printf("analyzer failed: %+v", err)
			}
		}
	}()

	err = e.Run()
	wg.Wait()

	if err != nil {
		log.Fatal(err)
	}
}
