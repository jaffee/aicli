package aws_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/jaffee/aicli/pkg/aicli"
	"github.com/jaffee/aicli/pkg/aws"
	"github.com/stretchr/testify/require"
)

func TestNewAI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	err := os.Setenv("AWS_REGION", "us-east-1")
	require.NoError(t, err)

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

	embs, err := ai.GetEmbedding(&aicli.EmbeddingRequest{
		Inputs: []string{"stuff mang"},
		Model:  aws.ModelTitanEmbedText,
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(embs))
	require.Equal(t, 1536, len(embs[0].Embedding))

	buf = &bytes.Buffer{}
	resp, err = ai.GenerateStream(&aicli.GenerateRequest{
		Model:       aws.ModelTitanTextExpress,
		Temperature: 0.7,
		TopP:        0.5,
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
