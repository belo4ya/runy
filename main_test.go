package runy

import (
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// signal_test.go
		goleak.IgnoreAnyFunction("github.com/belo4ya/runy.(*Task).Run"),
		goleak.IgnoreAnyFunction("github.com/belo4ya/runy.sendSignal"),
		goleak.IgnoreAnyFunction("github.com/belo4ya/runy.SetupSignalHandler.func1"),
		goleak.IgnoreAnyFunction("github.com/belo4ya/runy.TestSetupSignalHandler.func1"),
	)
}
