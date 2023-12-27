package openai

import (
	"context"
	"io"
	"strings"

	"github.com/jaffee/aicli/pkg/aicli"
	"github.com/pkg/errors"
	openai "github.com/sashabaranov/go-openai"
)

type Config struct {
	APIKey string `flag:"openai-api-key" help:"Your API key for OpenAI."`
}

func NewConfig() Config {
	return Config{}
}

var _ aicli.AI = &Client{} // assert that Client satisfies AI interface

type Client struct {
	subclient *openai.Client
}

func NewClient(conf Config) *Client {
	return &Client{
		subclient: openai.NewClient(conf.APIKey),
	}
}

func toOpenAIMessages(msgs []aicli.Message) []openai.ChatCompletionMessage {
	ret := make([]openai.ChatCompletionMessage, len(msgs))
	for i, msg := range msgs {
		ret[i] = openai.ChatCompletionMessage{
			Role:    msg.Role(),
			Content: msg.Content(),
		}
	}
	return ret
}

func (c *Client) GenerateStream(req *aicli.GenerateRequest, output io.Writer) (resp aicli.Message, err error) {
	stream, err := c.subclient.CreateChatCompletionStream(context.Background(),
		openai.ChatCompletionRequest{
			Model:       req.Model,
			Temperature: float32(req.Temperature),
			Messages:    toOpenAIMessages(req.Messages),
			MaxTokens:   req.MaxGenLen,
			TopP:        float32(req.TopP),
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeText,
			},
			Stream: true,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "making chat request")
	}

	defer stream.Close()
	totalResp := strings.Builder{}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, errors.Wrap(err, "recv")
		}
		if len(resp.Choices) == 0 {
			return nil, errors.New("no response choices in stream chunk")
		}
		chunk := resp.Choices[0].Delta.Content
		if _, err := io.WriteString(output, chunk); err != nil {
			return nil, errors.Wrap(err, "writing chunk")
		}
		_, _ = totalResp.WriteString(chunk)
	}

	msg := aicli.SimpleMsg{
		RoleField:    openai.ChatMessageRoleAssistant,
		ContentField: totalResp.String(),
	}
	return msg, nil
}

func (c *Client) GetEmbedding(req *aicli.EmbeddingRequest) ([]aicli.Embedding, error) {
	resp, err := c.subclient.CreateEmbeddings(context.Background(), openai.EmbeddingRequestStrings{
		Input: req.Inputs,
		Model: openai.EmbeddingModel(req.Model),
	})
	if err != nil {
		return nil, errors.Wrap(err, "get embeddings")
	}
	ret := make([]aicli.Embedding, len(resp.Data))
	for i, emb := range resp.Data {
		ret[i].Embedding = emb.Embedding
	}
	return ret, nil
}
