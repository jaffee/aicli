package aws

import (
	"encoding/json"
	"strings"
	"text/template"

	"github.com/jaffee/aicli/pkg/aicli"
	"github.com/pkg/errors"
)

type LlamaSubModel struct{}

func (m LlamaSubModel) MakeBody(req *aicli.GenerateRequest) ([]byte, error) {
	prompt, err := llamaPromptifyMessages(req.Messages)
	if err != nil {
		return nil, err
	}
	bod := llamaBody{
		Prompt:      prompt,
		Temperature: req.Temperature,
		TopP:        req.TopP,
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
	TopP        float64 `json:"top_p,omitempty"`
	MaxGenLen   int     `json:"max_gen_len,omitempty"`
}

func llamaPromptifyMessages(msgs []aicli.Message) (string, error) {
	temp, err := template.ParseFS(templateFS, "templates/llama.txt")
	if err != nil {
		return "", errors.Wrap(err, "parsing template")
	}

	sb := &strings.Builder{}
	if err := temp.Execute(sb, msgs); err != nil {
		return "", errors.Wrap(err, "executing template")
	}
	return sb.String(), nil

}

type llamaEvent struct {
	Generation           string  `json:"generation"`
	PromptTokenCount     int     `json:"prompt_token_count"`
	GenerationTokenCount int     `json:"generation_token_count"`
	StopReason           *string `json:"stop_reason"`
}
