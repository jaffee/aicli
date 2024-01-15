package aicli

import (
	"io"
	"strconv"
	"strings"

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
	// parse req.Model to see what size vector to return. Format is name_num.
	split := strings.Split(req.Model, "_")
	var vectorSize int
	switch len(split) {
	case 0:
		return nil, errors.New("empty model name")
	case 1:
		vectorSize = 1
	default:
		num, err := strconv.Atoi(split[1])
		if err != nil {
			return nil, errors.Wrap(err, "parsing model vector size")
		}
		vectorSize = num
	}
	if vectorSize <= 0 {
		return nil, errors.Errorf("vector size must be greater than 0, got %d", vectorSize)
	}
	ret := make([]Embedding, len(req.Inputs))
	for i := range ret {
		ret[i].Embedding = make([]float32, vectorSize)
		for j := range ret[i].Embedding {
			ret[i].Embedding[j] = 0.42
		}
	}
	return ret, nil
}
