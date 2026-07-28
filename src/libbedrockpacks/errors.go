package libbedrockpacks

import (
	"errors"
	"fmt"
	"runtime"
)

// ErrNotImplemented serves as our foundational sentinel error
var ErrNotImplemented = errors.New("not implemented")

// NotImplementedErr dynamically grabs the name of whichever function invokes it
func NotImplementedErr() error {
	// Get program counters of function invocations on the calling goroutine's stack.
	// Note: using 1 as argument to skip 'this function' and look at its immediately previous caller
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return ErrNotImplemented
	}

	// Extracts the fully qualified function name (e.g., "main.Run")
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return ErrNotImplemented
	}

	return fmt.Errorf("function %s: %w", fn.Name(), ErrNotImplemented)
}
