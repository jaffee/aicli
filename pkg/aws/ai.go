package aws

import (
	"context"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/jaffee/aicli/pkg/aicli"
	"github.com/pkg/errors"
)

const (
	ModelLlama213BChatV1  = "meta.llama2-13b-chat-v1"
	ModelLlama270BChatV1  = "meta.llama2-70b-chat-v1"
	ModelTitanTextExpress = "amazon.titan-text-express-v1"
	ModelTitanEmbedText   = "amazon.titan-embed-text-v1"
)

// NewAI gets a new AI which uses the default AWS configuration (i.e. ~/.aws/config and standard AWS env vars).
func NewAI() (*AI, error) {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, errors.Wrap(err, "loading default aws config")
	}

	brrc := bedrockruntime.NewFromConfig(cfg)
	return &AI{
		client: brrc,
	}, nil
}

func NewAIFromConfig(cfg aws.Config) (*AI, error) {
	brrc := bedrockruntime.NewFromConfig(cfg)
	return &AI{
		client: brrc,
	}, nil
}

type AI struct {
	client *bedrockruntime.Client

	Output       *bedrock.ListFoundationModelsOutput
	CustomOutput *bedrock.ListCustomModelsOutput
}

func (ai *AI) GenerateStream(req *aicli.GenerateRequest, output io.Writer) (aicli.Message, error) {
	var body []byte
	var sub AWSSubModel
	switch req.Model {
	case ModelLlama213BChatV1, "":
		// TODO we'll eventually need different implementations for different
		// models, but I only care about llama2 at the moment
		sub = LlamaSubModel{}
	default:
		return nil, errors.Errorf("%s is not currently a supported model (try 'meta.llama2-13b-chat-v1')", req.Model)
	}
	body, err := sub.MakeBody(req)
	if err != nil {
		return nil, errors.Wrap(err, "making body")
	}

	accept := "application/json"
	streamOutput, err := ai.client.InvokeModelWithResponseStream(context.Background(), &bedrockruntime.InvokeModelWithResponseStreamInput{
		Body:        body,
		ModelId:     &req.Model,
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
			chunk, err := sub.HandleResponseChunk(eventT.Value.Bytes)
			if err != nil {
				return nil, errors.Wrap(err, "handling chunk")
			}

			_, _ = bldr.Write(chunk)
			_, err = output.Write(chunk)
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

type AWSSubModel interface {
	MakeBody(req *aicli.GenerateRequest) ([]byte, error)
	HandleResponseChunk(chunkBytes []byte) ([]byte, error)
}
