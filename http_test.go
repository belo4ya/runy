package runy

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIgnoreHTTPServerClosed(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Returns nil when error is nil",
			args:    args{err: nil},
			wantErr: false,
		},
		{
			name:    "Returns nil when error is http.ErrServerClosed",
			args:    args{err: http.ErrServerClosed},
			wantErr: false,
		},
		{
			name:    "Returns error when other error",
			args:    args{err: errors.New("some error")},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IgnoreHTTPServerClosed(tt.args.err)
			assert.Equalf(t, tt.wantErr, err != nil, "IgnoreHTTPServerClosed() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
