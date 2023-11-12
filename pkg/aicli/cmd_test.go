package aicli

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCmd(t *testing.T) {
	cmd := NewCmd(&Echo{})
	stdinr, stdinw := io.Pipe()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd.stdin = stdinr
	cmd.stdout = stdout
	cmd.stderr = stderr
	cmd.historyPath = t.TempDir() + "/.aicli_history"
	cmd.OpenAI_API_Key = "blah"

	done := make(chan struct{})
	var runErr error
	go func() {
		runErr = cmd.Run()
		close(done)
	}()
	time.Sleep(time.Millisecond * 10)
	require.NoError(t, runErr)
	expect(t, stdout, []byte{0x20, 0x08, 0x1b, 0x5b, 0x36, 0x6e, 0x3e, 0x20})
	stdinw.Close()

	select {
	case <-done:
	case <-time.After(time.Second * 2):
		t.Fatal("command should have ended after stdin was closed")
	}

	require.NoError(t, runErr)
}

func expect(t *testing.T, r io.Reader, exp []byte) {
	t.Helper()
	buffer := make([]byte, len(exp))

	n, err := r.Read(buffer)
	if err != nil && err.Error() != "EOF" {
		require.NoError(t, err)
	}

	require.Equal(t, exp, buffer[:n])
}
