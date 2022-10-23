# kachtomize
A cacheable version of `kustomize build`

## Usage

```console
$ kachtomize << EOF
overlays/dev/01
overlays/dev/02
overlays/prod/01
overlays/prod/02
EOF
```

Built artifacts are saved in `overlays/{dev,prod}/{01,02}/artifact.yaml`

## License
Under the MIT License
