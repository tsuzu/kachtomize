package main

import (
	"flag"
	"log"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/tsuzu/kachtomize/pkg/kustomize"
)

var (
	ignoreErrors     bool
	dir              string
	kustomizeOptions []string
)

func init() {
	flag.BoolVar(&ignoreErrors, "ignore-errors", false, "Ignore corrupting kustomization.yaml while analyzing")

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

	targets, err = kustomize.FilterByIgnore(filepath.Join(dir, ".kacheignore"), targets)

	if err != nil {
		log.Fatal(err)
	}

	analyzer := kustomize.NewDependencyAnalyzer(targets, ignoreErrors, runtime.GOMAXPROCS(0))

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
				log.Printf("executor failed: %+v", err)
			}
		}
	}()

	err = e.Run()
	wg.Wait()

	if err != nil {
		log.Fatal(err)
	}
}
