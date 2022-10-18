package kustomize

import (
	"reflect"
	"testing"
)

func TestListKustomizeTarget(t *testing.T) {
	type args struct {
		dir string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "testdata/dir",
			args: args{
				dir: "./testdata/dir",
			},
			want: []string{
				"testdata/dir/a/b/c/kustomization.yaml",
				"testdata/dir/a/b/c2/kustomization.yaml",
				"testdata/dir/a/b/kustomization.yml",
				"testdata/dir/kustomization.yaml",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ListKustomizeTarget(tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListKustomizeTarget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListKustomizeTarget() = %v, want %v", got, tt.want)
			}
		})
	}
}
