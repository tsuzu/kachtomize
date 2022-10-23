package main

import (
	"bufio"
	"flag"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/tsuzu/kachtomize/pkg/fsloader"
	"github.com/tsuzu/kachtomize/pkg/fsutil"
	"github.com/tsuzu/kachtomize/pkg/krunner"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

var (
	outputFileName string
	useInMemFS     bool
	loadDirs       []string
)

func init() {
	flag.StringVar(&outputFileName, "o", "artifact.yaml", "Output filename")
	flag.BoolVar(&useInMemFS, "inmemfs", false, "Load files on memory before kustomize build")

	flag.Parse()

	loadDirs = flag.Args()
}

func main() {
	fs := filesys.MakeFsOnDisk()

	if useInMemFS {
		fs = fsutil.MakeFsInMemory()
		loader := fsloader.New(fs)

		if err := loader.LoadAll(loadDirs, runtime.GOMAXPROCS(0)); err != nil {
			panic(err)
		}

		fs = fsutil.NewReadOnlyFS(fs)
	}

	kustomizerInit := func() *krusty.Kustomizer {
		opt := krusty.MakeDefaultOptions()
		c := types.EnabledPluginConfig(types.BploUseStaticallyLinked)
		c.FnpLoadingOptions.EnableExec = true
		opt.LoadRestrictions = types.LoadRestrictionsNone
		opt.PluginConfig = c

		return krusty.MakeKustomizer(opt)
	}

	krunner := krunner.New(kustomizerInit, fs, runtime.GOMAXPROCS(0))
	krunner.RegisterCallback(func(dir string, data []byte) {
		fileName := filepath.Join(dir, outputFileName)

		if err := os.MkdirAll(filepath.Dir(fileName), 0777); err != nil {
			panic(err)
		}

		if err := os.WriteFile(fileName, data, 0777); err != nil {
			panic(err)
		}
	})

	wd, err := os.Getwd()

	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		// ファイルパスがそんな長いわけないのでisPrefixは無視します
		line, _, err := reader.ReadLine()

		if err != nil {
			if err == io.EOF {
				break
			}

			panic(err)
		}

		if len(line) == 0 {
			continue
		}

		krunner.Enqueue(filepath.Join(wd, string(line)))
	}

	if !krunner.Wait() {
		os.Exit(1)
	}
}
