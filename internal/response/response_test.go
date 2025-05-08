package response

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteStatusLine(t *testing.T) {
	tests := []struct {
		name       string
		statusCode StatusCode
		expected   string
		expectErr  bool
	}{
		{
			name:       "OK status",
			statusCode: StatusCodeOk,
			expected:   "HTTP/1.1 200 OK\r\n",
			expectErr:  false,
		},
		{
			name:       "Bad Request status",
			statusCode: StatusCodeBadRequest,
			expected:   "HTTP/1.1 400 Bad Request\r\n",
			expectErr:  false,
		},
		{
			name:       "Internal Server Error status",
			statusCode: StatusCodeServerError,
			expected:   "HTTP/1.1 500 Internal Server Error\r\n",
			expectErr:  false,
		},
		{
			name:       "Unknown status code",
			statusCode: 999,
			expected:   "HTTP/1.1 999 \r\n",
			expectErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteStatusLine(&buf, tt.statusCode)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, buf.String())
			}
		})
	}
}
