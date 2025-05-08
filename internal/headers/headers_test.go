package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaders(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.True(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Valid multiple headers
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nUser-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:138.0) Gecko/20100101 Firefox/138.0\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:138.0) Gecko/20100101 Firefox/138.0", headers["user-agent"])
	assert.Equal(t, 121, n)
	assert.True(t, done)

	// Test: Incomplete header line
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nUser-Agent: Mozilla/5.0")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Empty input
	headers = NewHeaders()
	data = []byte("")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Header with no value
	headers = NewHeaders()
	data = []byte("Host:\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Header with invalid characters in key
	headers = NewHeaders()
	data = []byte("Ho@st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Duplicate headers
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nSet-Person: lane-loves-go\r\nSet-Person: Bobs-your-uncle\r\nSet-Person: Good Course\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, "lane-loves-go, Bobs-your-uncle, Good Course", headers["set-person"])
	assert.Equal(t, 104, n)
	assert.True(t, done)
}
