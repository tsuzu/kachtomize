package kustomize

import (
	"reflect"
	"runtime"
	"testing"
)

func TestDependencyAnalyzer_Run(t *testing.T) {
	type fields struct {
		kustomizes []string
	}
	tests := []struct {
		name    string
		fields  fields
		want    []Node
		wantErr bool
	}{
		{
			name: "testdata/dir",
			fields: fields{
				kustomizes: []string{
					"testdata/dir/a/b/c/kustomization.yaml",
					"testdata/dir/a/b/c2/kustomization.yaml",
					"testdata/dir/a/b/kustomization.yml",
					"testdata/dir/kustomization.yaml",
				},
			},
			want: []Node{
				{
					AbsDirPath:    "/Users/tsuzu/workspace/hobby/kachtomize/pkg/kustomize/testdata/dir",
					KustomizePath: "/Users/tsuzu/workspace/hobby/kachtomize/pkg/kustomize/testdata/dir/kustomization.yaml",
					DependedBy:    []string{},
					Dependencies: []string{
						"/Users/tsuzu/workspace/hobby/kachtomize/pkg/kustomize/testdata/dir/a/b",
					},
				},
				{
					AbsDirPath:    "/Users/tsuzu/workspace/hobby/kachtomize/pkg/kustomize/testdata/dir/a/b",
					KustomizePath: "/Users/tsuzu/workspace/hobby/kachtomize/pkg/kustomize/testdata/dir/a/b/kustomization.yml",
					DependedBy: []string{
						"/Users/tsuzu/workspace/hobby/kachtomize/pkg/kustomize/testdata/dir",
					},
					Dependencies: []string{
						"/Users/tsuzu/workspace/hobby/kachtomize/pkg/kustomize/testdata/dir/a/b/c",
						"/Users/tsuzu/workspace/hobby/kachtomize/pkg/kustomize/testdata/dir/a/b/c2",
					},
				},
				{
					AbsDirPath:    "/Users/tsuzu/workspace/hobby/kachtomize/pkg/kustomize/testdata/dir/a/b/c",
					KustomizePath: "/Users/tsuzu/workspace/hobby/kachtomize/pkg/kustomize/testdata/dir/a/b/c/kustomization.yaml",
					DependedBy: []string{
						"/Users/tsuzu/workspace/hobby/kachtomize/pkg/kustomize/testdata/dir/a/b",
					},
					Dependencies: nil,
				},
				{
					AbsDirPath:    "/Users/tsuzu/workspace/hobby/kachtomize/pkg/kustomize/testdata/dir/a/b/c2",
					KustomizePath: "/Users/tsuzu/workspace/hobby/kachtomize/pkg/kustomize/testdata/dir/a/b/c2/kustomization.yaml",
					DependedBy: []string{
						"/Users/tsuzu/workspace/hobby/kachtomize/pkg/kustomize/testdata/dir/a/b",
					},
					Dependencies: nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			da := *NewDependencyAnalyzer(tt.fields.kustomizes, false, runtime.GOMAXPROCS(0))
			got, err := da.Run()
			if (err != nil) != tt.wantErr {
				t.Errorf("DependencyAnalyzer.Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DependencyAnalyzer.Run() = %v, want %v", got, tt.want)
			}
		})
	}
}
