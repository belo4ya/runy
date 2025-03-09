package runy

import (
	"errors"
	"net/http"
)

func IgnoreHTTPServerClosed(err error) error {
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
