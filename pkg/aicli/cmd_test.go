package aicli

import (
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCmd(t *testing.T) {
	cmd := NewCmd()
	cmd.AI = "echo"
	cmd.Color = false
	cmd.AddAI("echo", &Echo{})
	stdinr, stdinw := io.Pipe()
	stdout, stdoutw := io.Pipe()
	stderr, stderrw := io.Pipe()

	cmd.stdin = stdinr
	cmd.stdout = stdoutw
	cmd.stderr = stderrw
	cmd.dotAICLIDir = t.TempDir()

	done := make(chan struct{})
	var runErr error
	go func() {
		runErr = cmd.Run()
		if runErr != nil {
			fmt.Printf("error: %v\n", runErr)
		}
		close(done)
	}()

	// todo: need a way to fail quickly if anything hangs and report the line
	time.Sleep(time.Millisecond)
	require.NoError(t, runErr)
	write(t, stdinw, []byte("blah\n"))
	require.NoError(t, runErr)
	expect(t, stdout, []byte("blah\n"))
	write(t, stdinw, []byte("bleh\n"))
	require.NoError(t, runErr)
	expect(t, stdout, []byte("bleh\n"))
	write(t, stdinw, []byte("\\messages\n"))
	expect(t, stdout, []byte("     user: blah\nassistant: blah\n     user: bleh\nassistant: bleh\n"))
	write(t, stdinw, []byte("\\reset\n"))
	require.NoError(t, runErr)
	write(t, stdinw, []byte("\\config\n"))
	expect(t, stderr, []byte("AI: echo\nModel: gpt-3.5-turbo\nTemperature: 0.700000\nVerbose: false\nContextLimit: 10000\nColor: false\n"))
	write(t, stdinw, []byte("\\reset\n"))
	write(t, stdinw, []byte("\\file ./testdata/myfile\n"))
	write(t, stdinw, []byte("\\messages\n"))
	expect(t, stdout, []byte("     user: Here is a file named './testdata/myfile' that I'll refer to later:\n```\nhaha\n```\n\n"))

	write(t, stdinw, []byte("\\reset\n"))
	write(t, stdinw, []byte("\\system I like chicken\n"))
	write(t, stdinw, []byte("\\messages\n"))
	expect(t, stdout, []byte("   system: I like chicken\n"))

	_ = stdinw.Close()

	select {
	case <-done:
	case <-time.After(time.Second * 2):
		t.Fatal("command should have ended after stdin was closed")
	}

	require.NoError(t, runErr)
}

func write(t *testing.T, to io.Writer, p []byte) {
	t.Helper()
	done := make(chan struct{})
	var err error

	go func() {
		_, err = to.Write(p)
		close(done)
	}()

	select {
	case <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatalf("timeout writing '%s'", p)
		return
	}

}

func expect(t *testing.T, r io.Reader, exp []byte) {
	t.Helper()
	buffer := make([]byte, len(exp)*20)
	i := 0
	var tot int
	var n int
	var err error
	for {
		i++
		done := make(chan struct{})
		go func() {
			n, err = r.Read(buffer[tot:])
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("more than 1 second waiting for output")
			return
		}
		if err != nil && err.Error() != "EOF" {
			require.NoError(t, err)
		}
		tot += n
		if tot >= len(exp) {
			break
		}
		if i > 20 {
			t.Fatal("spent too long waiting for output")
		}
		time.Sleep(time.Millisecond)
	}

	require.Equal(t, exp, buffer[:tot])
}
