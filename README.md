# rc

The Runtime Conditions command line, built with [Cobra](https://github.com/spf13/cobra).

```sh
rc profile <language> [args...]
rc --version
```

`rc profile` runs an installed language profiler, passes the remaining arguments to it
unchanged, and exits with the profiler's status. It looks for the executable named by
`RC_<LANGUAGE>_PROFILER_BIN`, then `rc-profiler-<language>` on `PATH`. The `go`, `python`
and `typescript` profilers are also found as `go-rc-profiler`,
`runtimeconditions-python-profiler` and `runtimeconditions-typescript-profiler`.

```sh
npm install --save-dev @runtimeconditions/typescript-profiler
rc profile typescript generate --project . --name orders --workload-uri https://services.example.com/orders
```

A profiler for a new language needs no change here: put an executable named
`rc-profiler-<language>` on `PATH`.

## Build

```sh
go build -ldflags "-X main.version=0.1.0" -o rc ./cmd/rc
go test ./...
```

`rc bindings` still lives in
[`runtimeconditions/extensions`](https://github.com/runtimeconditions/extensions) and moves here once
its Go modules can be fetched from outside that repository.

MIT licensed.
