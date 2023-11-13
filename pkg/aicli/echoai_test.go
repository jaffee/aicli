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
			msg, err := c.StreamResp(tst.msgs, buf)
			require.NoError(t, err)

			assert.Equal(t, tst.exp, buf.String())
			assert.Equal(t, tst.exp, msg.Content())
			assert.Equal(t, aicli.RoleAssistant, msg.Role())
		})
	}
}
