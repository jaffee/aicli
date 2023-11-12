package aicli

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
)

var _ AI = &Echo{} // assert that Client satisfies AI interface

// Echo is an AI implementation for testing which repeats the user's last
// message back with some extra information.
type Echo struct{}

func (c *Echo) StreamResp(msgs []Message, output io.Writer) (Message, error) {
	var resp string
	if len(msgs) == 0 {
		resp = "0 msgs"
	} else {
		resp = fmt.Sprintf("msgs: %d, role: %s, content: %s", len(msgs), RoleAssistant, msgs[len(msgs)-1].Content())
	}

	_, err := output.Write([]byte(resp))
	if err != nil {
		return nil, errors.Wrap(err, "writing msg")
	}
	return SimpleMsg{RoleField: RoleAssistant, ContentField: resp}, nil
}
