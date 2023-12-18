package aws

import (
	"encoding/json"

	"github.com/jaffee/aicli/pkg/aicli"
	"github.com/pkg/errors"
)

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
	TopP          float32  `json:"topP"`
	MaxTokenCount int      `json:"maxTokenCount"`
	StopSequences []string `json:"stopSequences"`
}

type titanEmbedResponse struct {
	Embedding           []float32 `json:"embedding"`
	InputTextTokenCount int       `json:"inputTextTokenCount"`
}

type TitanTextSubModel struct{}

func (m TitanTextSubModel) MakeBody(req *aicli.GenerateRequest) ([]byte, error) {
	return nil, errors.New("unimplemented")
}

func (m TitanTextSubModel) HandleResponseChunk(chunkBytes []byte) ([]byte, error) {
	return nil, errors.New("unimplemented")
}
