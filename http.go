package runy

import (
	"errors"
	"net/http"
)

// IgnoreHTTPServerClosed returns nil on http.ErrServerClosed errors.
// All other values that are not http.ErrServerClosed errors or nil are returned unmodified.
func IgnoreHTTPServerClosed(err error) error {
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
