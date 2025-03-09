# runy - tiny run manager

[![tag](https://img.shields.io/github/tag/belo4ya/runy.svg)](https://github.com/belo4ya/runy/releases)
![go version](https://img.shields.io/badge/-%E2%89%A51.20-%23027d9c?logo=go&logoColor=white&labelColor=%23555)
[![go doc](https://godoc.org/github.com/belo4ya/runy?status.svg)](https://pkg.go.dev/github.com/belo4ya/runy)
[![go report](https://goreportcard.com/badge/github.com/belo4ya/runy)](https://goreportcard.com/report/github.com/belo4ya/runy)
[![codecov](https://codecov.io/gh/belo4ya/runy/graph/badge.svg?token=GQZRP94G21)](https://codecov.io/gh/belo4ya/runy)
[![license](https://img.shields.io/github/license/belo4ya/runy)](./LICENSE)

🎯 The goal of the project is to provide developers with the opportunity not to think about the graceful shutdown
and not to make mistakes in its implementation in their application.
Instead, focus on startup components such as HTTP and gRPC servers and other `Runnable`s.

## 🚀 Install

```sh
go get -u github.com/belo4ya/runy
```

**Compatibility:** Go ≥ 1.20

## 🧠 Core Concepts

Runnable, SugaredRunnable, Group...

## 💡 Usage

- GoDoc: https://pkg.go.dev/github.com/belo4ya/runy
- End-to-end usage examples: [examples/](examples)
- Common `Runnable`s (http, grpc servers, kafka consumers): [examples/runnable](examples/runnables)

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
	ctx := runy.SetupSignalHandler() // handle SIGINT and SIGTERM

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
		wait.UntilWithContext(ctx, func(ctx context.Context) {
			slog.Info("worker does useful things")
		}, 10*time.Second)
		return nil
	})

	slog.InfoContext(ctx, "starting app")
	if err := runy.Start(ctx); err != nil { // run app
		return fmt.Errorf("problem with running app: %w", err)
	}
	return nil
}
```

## 📚 Acknowledgments

The following projects had a particular impact on the design of runy.

- [kubernetes-sigs/controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) - set of go libraries for building Kubernetes controllers.
- [oklog/run](https://github.com/oklog/run) - universal mechanism to manage goroutine lifecycles.
- [go-kratos/kratos](https://github.com/go-kratos/kratos) - ultimate Go microservices framework for the cloud-native era.
- [sourcegraph/conc](https://github.com/sourcegraph/conc) - better structured concurrency for Go.
