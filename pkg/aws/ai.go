package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/jaffee/aicli/pkg/aicli"
	"github.com/pkg/errors"
)

const (
	ModelLlama213BChatV1 = "meta.llama2-13b-chat-v1"
)

// NewAI gets a new AI which uses the default AWS configuration (i.e. ~/.aws/config and standard AWS env vars).
func NewAI() (*AI, error) {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, errors.Wrap(err, "loading default aws config")
	}

	// brc := bedrock.NewFromConfig(cfg)
	// lfmOutput, err := brc.ListFoundationModels(context.Background(), &bedrock.ListFoundationModelsInput{})
	// if err != nil {
	// 	return nil, errors.Wrap(err, "listing models")
	// }

	// lcmOutput, err := brc.ListCustomModels(context.Background(), &bedrock.ListCustomModelsInput{})
	// if err != nil {
	// 	return nil, errors.Wrap(err, "listing custom models")
	// }

	brrc := bedrockruntime.NewFromConfig(cfg)
	return &AI{
		client: brrc,
		//Output:       lfmOutput,
		//CustomOutput: lcmOutput,
	}, nil
}

type AI struct {
	client *bedrockruntime.Client

	Output       *bedrock.ListFoundationModelsOutput
	CustomOutput *bedrock.ListCustomModelsOutput
}

func (ai *AI) GenerateStream(req *aicli.GenerateRequest, output io.Writer) (aicli.Message, error) {
	fmt.Printf("req: %+v\n", req)
	accept := "application/json"
	model := ModelLlama213BChatV1
	switch req.Model {
	case ModelLlama213BChatV1, "":
		// TODO we'll eventually need different implementations for different
		// models, but I only care about llama2 at the moment
	default:
		return nil, errors.Errorf("%s is not currently a supported model (try 'meta.llama2-13b-chat-v1')", req.Model)
	}
	bod := LlamaBody{
		Prompt:      promptifyMessages(req.Messages),
		Temperature: req.Temperature,
		TopP:        0.9,
		MaxGenLen:   100,
	}

	bs, err := json.Marshal(bod)
	if err != nil {
		return nil, errors.Wrap(err, "marshalling")
	}
	fmt.Printf("bod: %s\n", bs)

	streamOutput, err := ai.client.InvokeModelWithResponseStream(context.Background(), &bedrockruntime.InvokeModelWithResponseStreamInput{
		Body:        bs,
		ModelId:     &model,
		Accept:      &accept,
		ContentType: &accept,
	})
	if err != nil {
		return nil, errors.Wrap(err, "invoking model")
	}

	bldr := &strings.Builder{}

	echan := streamOutput.GetStream().Events()
	for event := range echan {
		switch eventT := event.(type) {
		case *types.ResponseStreamMemberChunk:
			chunk := &Event{}
			if err := json.Unmarshal(eventT.Value.Bytes, chunk); err != nil {
				return nil, errors.Wrap(err, "unmarshaling response")
			}

			bldr.WriteString(chunk.Generation)
			_, err := output.Write([]byte(chunk.Generation))
			if err != nil {
				return nil, errors.Wrap(err, "writing output")
			}
		default:
			return nil, errors.Errorf("unknown event type %+v", eventT)
		}
	}

	return aicli.SimpleMsg{ContentField: bldr.String(), RoleField: aicli.RoleAssistant}, nil
}

func (ai *AI) GetEmbedding(req *aicli.EmbeddingRequest) ([]aicli.Embedding, error) {
	return nil, errors.New("unimplemented")
}

type LlamaBody struct {
	Prompt      string  `json:"prompt"`
	Temperature float64 `json:"temperature"`
	TopP        float64 `json:"top_p"`
	MaxGenLen   int     `json:"max_gen_len"`
}

func promptifyMessages(msgs []aicli.Message) string {
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
	bldr.WriteString(" [/INST]\n")
	return bldr.String()
}

type Event struct {
	Generation           string  `json:"generation"`
	PromptTokenCount     int     `json:"prompt_token_count"`
	GenerationTokenCount int     `json:"generation_token_count"`
	StopReason           *string `json:"stop_reason"`
}
