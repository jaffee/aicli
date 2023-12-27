package aws

import (
	"testing"

	"github.com/jaffee/aicli/pkg/aicli"
	"github.com/stretchr/testify/require"
)

func TestTitanPromptifyMessages(t *testing.T) {
	s, err := titanPromptifyMessages([]aicli.Message{
		aicli.SimpleMsg{
			RoleField:    aicli.RoleSystem,
			ContentField: "do stuff",
		},
		aicli.SimpleMsg{
			RoleField:    aicli.RoleUser,
			ContentField: "i am user",
		},
		aicli.SimpleMsg{
			RoleField:    aicli.RoleAssistant,
			ContentField: "i am assistant",
		},
		aicli.SimpleMsg{
			RoleField:    aicli.RoleUser,
			ContentField: "please the towel",
		},
	})
	require.NoError(t, err)

	exp := `do stuff

User: i am user

Bot: i am assistant

User: please the towel

Bot: `
	require.Equal(t, exp, s)
}
