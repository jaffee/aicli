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

	c := NewClient("http://localhost:11434", "llama2:nogpu", 0.4)

	buf := &bytes.Buffer{}

	resp, err := c.StreamResp([]aicli.Message{aicli.SimpleMsg{RoleField: "user", ContentField: "hello"}}, buf)
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
