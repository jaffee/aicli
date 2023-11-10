package openai

import (
	"context"
	"io"
	"strings"

	"github.com/jaffee/aicli/pkg/aicli"
	"github.com/pkg/errors"
	openai "github.com/sashabaranov/go-openai"
)

type Client struct {
	model string

	subclient *openai.Client
}

func NewClient(apiKey, model string) *Client {
	return &Client{
		model: model,

		subclient: openai.NewClient(apiKey),
	}
}

func (c *Client) SetModel(model string) {
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

func (c *Client) StreamResp(msgs []aicli.Message, output io.Writer) (resp aicli.Message, err error) {
	stream, err := c.subclient.CreateChatCompletionStream(context.Background(),
		openai.ChatCompletionRequest{
			Model:    c.model,
			Messages: toOpenAIMessages(msgs),
			ResponseFormat: openai.ChatCompletionResponseFormat{
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
