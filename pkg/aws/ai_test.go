package aws_test

import (
	"bytes"
	"testing"

	"github.com/jaffee/aicli/pkg/aicli"
	"github.com/jaffee/aicli/pkg/aws"
	"github.com/stretchr/testify/require"
)

func TestNewAI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ai, err := aws.NewAI()
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	resp, err := ai.GenerateStream(&aicli.GenerateRequest{
		Model:       aws.ModelLlama213BChatV1,
		Temperature: 0.7,
		Messages: []aicli.Message{
			aicli.SimpleMsg{
				RoleField:    aicli.RoleUser,
				ContentField: "hello, please respond with 'hello'",
			},
		}}, buf)

	require.NoError(t, err)
	require.Equal(t, "assistant", resp.Role())
	require.True(t, 4 < len(resp.Content()))
	require.True(t, 4 < len(buf.Bytes()))
}
