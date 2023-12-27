package aws

import (
	"embed"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/jaffee/aicli/pkg/aicli"
	"github.com/pkg/errors"
)

//go:embed templates
var templateFS embed.FS

type TitanEmbedTextSubModel struct{}

func (m TitanEmbedTextSubModel) MakeBodyEmbed(req *aicli.EmbeddingRequest) ([]byte, error) {
	if len(req.Inputs) != 1 {
		return nil, errors.New("this model supports exactly 1 input for embedding")
	}
	bod := titanInvokeRequest{
		InputText: req.Inputs[0],
	}
	bs, err := json.Marshal(bod)
	return bs, err
}

func (m TitanEmbedTextSubModel) HandleResponseEmbed(body []byte) ([]float32, error) {
	resp := titanEmbedResponse{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, errors.Wrap(err, "unmarshaling")
	}
	return resp.Embedding, nil
}

type titanInvokeRequest struct {
	InputText            string                     `json:"inputText"`
	TextGenerationConfig *titanTextGenerationConfig `json:"textGenerationConfig,omitempty"`
}

type titanTextGenerationConfig struct {
	Temperature   float32  `json:"temperature"`
	TopP          float32  `json:"topP,omitempty"`
	MaxTokenCount int      `json:"maxTokenCount,omitempty"`
	StopSequences []string `json:"stopSequences,omitempty"`
}

type titanInvokeResponse struct {
	InputTextTokenCount int `json:"inputTextTokenCount"`
	Results             []titanInvokeResponseResult
}

type titanInvokeResponseResult struct {
	TokenCount int    `json:"tokenCount"`
	OutputText string `json:"outputText"`
}

type titanEmbedResponse struct {
	Embedding           []float32 `json:"embedding"`
	InputTextTokenCount int       `json:"inputTextTokenCount"`
	CompletionReason    string    `json:"completionReason"`
}

type titanInvokeResponseChunk struct {
	Index                     int    `json:"index"`
	InputTextTokenCount       int    `json:"inputTextTokenCount"`
	TotalOutputTextTokenCount int    `json:"totalOutputTextTokenCount"`
	OutputText                string `json:"outputText"`
	CompletionReason          string `json:"completionReason"`
}

const (
	completionReasonFinished = "FINISHED"
	completionReasonLength   = "LENGTH"
)

type TitanTextSubModel struct{}

func (m TitanTextSubModel) MakeBody(req *aicli.GenerateRequest) ([]byte, error) {
	body, err := titanPromptifyMessages(req.Messages)
	if err != nil {
		return nil, err
	}
	tr := titanInvokeRequest{
		InputText: body,
		TextGenerationConfig: &titanTextGenerationConfig{
			Temperature:   float32(req.Temperature),
			MaxTokenCount: req.MaxGenLen,
			TopP:          float32(req.TopP),
		},
	}
	return json.MarshalIndent(tr, "", "  ")
}

func (m TitanTextSubModel) HandleResponseChunk(chunkBytes []byte) ([]byte, error) {
	c := titanInvokeResponseChunk{}
	err := json.Unmarshal(chunkBytes, &c)
	return []byte(c.OutputText), err
}

func titanPromptifyMessages(msgs []aicli.Message) (string, error) {
	temp, err := template.ParseFS(templateFS, "templates/titan.txt")
	if err != nil {
		return "", errors.Wrap(err, "parsing template")
	}

	sb := &strings.Builder{}
	if err := temp.Execute(sb, msgs); err != nil {
		return "", errors.Wrap(err, "executing template")
	}
	return sb.String(), nil
}
