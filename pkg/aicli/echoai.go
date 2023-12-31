package aicli

import (
	"io"

	"github.com/pkg/errors"
)

var _ AI = &Echo{} // assert that Client satisfies AI interface

// Echo is an AI implementation for testing which repeats the user's last
// message back.
type Echo struct{}

func (c *Echo) GenerateStream(req *GenerateRequest, output io.Writer) (Message, error) {
	var resp string
	if len(req.Messages) == 0 {
		resp = "0 msgs"
	} else {
		resp = req.Messages[len(req.Messages)-1].Content()
	}

	_, err := output.Write([]byte(resp))
	if err != nil {
		return nil, errors.Wrap(err, "writing msg")
	}
	return SimpleMsg{RoleField: RoleAssistant, ContentField: resp}, nil
}

func (c *Echo) GetEmbedding(req *EmbeddingRequest) ([]Embedding, error) {
	if len(req.Inputs) == 0 {
		return nil, errors.New("emtpy list of inputs")
	}
	ret := make([]Embedding, len(req.Inputs))
	for i := range ret {
		ret[i].Embedding = []float32{0.42}
	}
	return ret, nil

}
