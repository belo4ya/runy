# runy - tiny run manager

[![tag](https://img.shields.io/github/tag/belo4ya/runy.svg)](https://github.com/belo4ya/runy/releases)
![Go Version](https://img.shields.io/github/go-mod/go-version/belo4ya/runy?logo=go)
[![GoDoc](https://godoc.org/github.com/belo4ya/runy?status.svg)](https://pkg.go.dev/github.com/belo4ya/runy)
[![Go report](https://goreportcard.com/badge/github.com/belo4ya/runy)](https://goreportcard.com/report/github.com/belo4ya/runy)
[![Coverage](https://img.shields.io/codecov/c/github/belo4ya/runy)](https://codecov.io/gh/belo4ya/runy)
[![License](https://img.shields.io/github/license/belo4ya/runy)](./LICENSE)

## 🚀 Install

```sh
go get -u github.com/belo4ya/runy
```

`runy` also supports semver releases.

Note that `runy` only [supports](https://go.dev/doc/devel/release#policy) the two most recent minor versions of Go.

## 💡 Usage

You can import `runy` using:

```go
import (
    "github.com/belo4ya/runy"
)
```

Then use one of the helpers below:

```go
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := runy.SetupSignalHandler(context.Background()) // setup handler for SIGTERM and SIGINT

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		return fmt.Errorf("net.Listen: %w", err)
	}

	httpSrv := &http.Server{Addr: ":8080"}
	grpcSrv := grpc.NewServer()

	// register all Runnable app components
	runy.SAddF(func(ctx context.Context) error {
		return runy.IgnoreHTTPServerClosed(httpSrv.ListenAndServe())
	}, httpSrv.Shutdown)
	runy.SAddF(func(ctx context.Context) error {
		return grpcSrv.Serve(lis)
	}, func(ctx context.Context) error {
		grpcSrv.GracefulStop()
		return nil
	})
	runy.AddF(func(ctx context.Context) error {
		wait.UntilWithContext(ctx, func(ctx context.Context) { slog.Info("worker does useful things") }, 10*time.Second)
		return nil
	})

	slog.InfoContext(ctx, "starting app")
	if err := runy.Start(ctx); err != nil { // run app
		return fmt.Errorf("problem with running app: %w", err)
	}
	return nil
}
```

More examples in documentation.
