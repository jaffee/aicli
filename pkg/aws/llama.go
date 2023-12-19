package aws

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jaffee/aicli/pkg/aicli"
	"github.com/pkg/errors"
)

type LlamaSubModel struct{}

func (m LlamaSubModel) MakeBody(req *aicli.GenerateRequest) ([]byte, error) {
	bod := llamaBody{
		Prompt:      llamaPromptifyMessages(req.Messages),
		Temperature: req.Temperature,
		TopP:        0.9,
		MaxGenLen:   req.MaxGenLen,
	}

	bs, err := json.Marshal(bod)
	if err != nil {
		return nil, errors.Wrap(err, "marshalling")
	}

	return bs, nil
}

func (m LlamaSubModel) HandleResponseChunk(chunkBytes []byte) ([]byte, error) {
	chunk := llamaEvent{}
	if err := json.Unmarshal(chunkBytes, &chunk); err != nil {
		return nil, err
	}
	return []byte(chunk.Generation), nil
}

type llamaBody struct {
	Prompt      string  `json:"prompt"`
	Temperature float64 `json:"temperature"`
	TopP        float64 `json:"top_p"`
	MaxGenLen   int     `json:"max_gen_len"`
}

func llamaPromptifyMessages(msgs []aicli.Message) string {
	bldr := &strings.Builder{}
	bldr.WriteString("[INST] ")
	msgsStart := 0
	if msgs[0].Role() == aicli.RoleSystem {
		fmt.Fprintf(bldr, "<<SYS>>\n%s\n<</SYS>>\n", msgs[0].Content())
		msgsStart = 1
	}
	if len(msgs) == msgsStart {
		return bldr.String()
	}
	for _, msg := range msgs[msgsStart:] {
		fmt.Fprintf(bldr, "%s: %s\n", msg.Role(), msg.Content())
	}
	bldr.WriteString(" [/INST] ")
	return bldr.String()
}

type llamaEvent struct {
	Generation           string  `json:"generation"`
	PromptTokenCount     int     `json:"prompt_token_count"`
	GenerationTokenCount int     `json:"generation_token_count"`
	StopReason           *string `json:"stop_reason"`
}
