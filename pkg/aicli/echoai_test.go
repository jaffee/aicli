package aicli_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/jaffee/aicli/pkg/aicli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEcho(t *testing.T) {
	c := &aicli.Echo{}

	cases := []struct {
		msgs []aicli.Message
		exp  string
	}{
		{msgs: nil, exp: "0 msgs"},
		{
			msgs: []aicli.Message{
				aicli.SimpleMsg{
					ContentField: "hello",
					RoleField:    aicli.RoleUser,
				},
			},
			exp: "hello",
		},
	}

	for i, tst := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			buf := &bytes.Buffer{}
			gr := &aicli.GenerateRequest{
				Messages: tst.msgs,
			}
			msg, err := c.GenerateStream(gr, buf)
			require.NoError(t, err)

			assert.Equal(t, tst.exp, buf.String())
			assert.Equal(t, tst.exp, msg.Content())
			assert.Equal(t, aicli.RoleAssistant, msg.Role())
		})
	}
}

func TestEchoEmbeddings(t *testing.T) {
	c := &aicli.Echo{}
	cases := []struct {
		model  string
		exp    []float32
		expErr string
	}{
		{
			model: "basic",
			exp:   []float32{0.42},
		},
		{
			model: "numbered_2",
			exp:   []float32{0.42, 0.42},
		},
		{
			model:  "numbered_-11",
			expErr: "size must be greater than 0",
		},
	}

	for i, tst := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			er := &aicli.EmbeddingRequest{
				Model:  tst.model,
				Inputs: []string{"hello"},
			}
			emb, err := c.GetEmbedding(er)
			if tst.expErr != "" {
				require.Contains(t, err.Error(), tst.expErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tst.exp, emb[0].Embedding)
			}
		})
	}
}
