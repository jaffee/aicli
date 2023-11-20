package ollama

import (
	"bytes"
	"io"
	"testing"

	"github.com/jaffee/aicli/pkg/aicli"
	"github.com/stretchr/testify/require"
)

// must be running ollama
func TestOllamaClient(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	c := NewClient(NewConfig())

	buf := &bytes.Buffer{}

	gr := &aicli.GenerateRequest{
		Model:       "llama2:nogpu",
		Temperature: 0.7,
		Messages:    []aicli.Message{aicli.SimpleMsg{RoleField: "user", ContentField: "hello"}},
	}

	resp, err := c.StreamResp(gr, buf)
	if err != nil {
		t.Fatal(err)
	}

	bs := make([]byte, 200)
	n, err := buf.Read(bs)
	if err != nil && err == io.EOF {
		t.Fatal(err)
	}
	require.True(t, n > 5)
	require.Equal(t, string(bs[:n]), resp.Content())
	require.Equal(t, aicli.RoleAssistant, resp.Role())
}
