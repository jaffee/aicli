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
	cmd.OpenAIAPIKey = "blah"

	done := make(chan struct{})
	var runErr error
	go func() {
		runErr = cmd.Run()
		close(done)
	}()
	require.NoError(t, runErr)
	// expect(t, stdout, []byte{0x20, 0x08, 0x1b, 0x5b, 0x36, 0x6e, 0x3e, 0x20})
	_, _ = stdinw.Write([]byte("blah\n"))
	require.NoError(t, runErr)
	expect(t, stdout, []byte("blah\n"))
	_, _ = stdinw.Write([]byte("bleh\n"))
	require.NoError(t, runErr)
	expect(t, stdout, []byte("bleh\n"))
	_, _ = stdinw.Write([]byte("\\messages\n"))
	expect(t, stdout, []byte("     user: blah\nassistant: blah\n     user: bleh\nassistant: bleh\n"))
	_, _ = stdinw.Write([]byte("\\reset\n"))
	require.NoError(t, runErr)
	_, _ = stdinw.Write([]byte("\\config\n"))
	expect(t, stderr, []byte("OpenAI_API_Key: length=4\nOpenAIModel: gpt-3.5-turbo\nTemperature: 0.700000\nVerbose: false\n"))
	_, _ = stdinw.Write([]byte("\\reset\n"))
	_, _ = stdinw.Write([]byte("\\file ./testdata/myfile\n"))
	expect(t, stdout, []byte("Here is a file named './testdata/myfile' that I'll refer to later, you can just say 'ok': \n```\nhaha\n```\n\n"))
	_, _ = stdinw.Write([]byte("\\reset\n"))
	_, _ = stdinw.Write([]byte("\\system I like chicken\n"))
	_, _ = stdinw.Write([]byte("\\messages\n"))
	expect(t, stdout, []byte("   system: I like chicken\n"))

	_ = stdinw.Close()

	select {
	case <-done:
	case <-time.After(time.Second * 2):
		t.Fatal("command should have ended after stdin was closed")
	}

	require.NoError(t, runErr)
}

func expect(t *testing.T, r io.Reader, exp []byte) {
	t.Helper()
	buffer := make([]byte, len(exp)*20)
	i := 0
	var n int
	var err error
	for {
		i++
		n, err = r.Read(buffer)
		if err != nil && err.Error() != "EOF" {
			require.NoError(t, err)
		}
		if n > 0 {
			break
		}
		if i > 100 {
			t.Fatal("spent too long waiting for output")
		}
		time.Sleep(time.Millisecond)
	}

	require.Equal(t, exp, buffer[:n])
}
