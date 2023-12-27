package aws

import (
	"testing"

	"github.com/jaffee/aicli/pkg/aicli"
	"github.com/stretchr/testify/require"
)

func TestLlamaPromptifyMessages(t *testing.T) {
	s, err := llamaPromptifyMessages([]aicli.Message{
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

	exp := `<s>[INST] <<SYS>>
do stuff
<</SYS>>
i am user [/INST] i am assistant </s><s>[INST] please the towel [/INST] `
	require.Equal(t, exp, s)
}
