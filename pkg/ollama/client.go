package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jaffee/aicli/pkg/aicli"
	"github.com/pkg/errors"
)

var _ aicli.AI = &Client{} // assert that Client satisfies AI interface

type Config struct {
	Host string `flag:"ollama-host" help:"Endpoint to hit for Ollama API."`
}

func NewConfig() Config {
	return Config{
		Host: "http://localhost:11434",
	}
}

type Client struct {
	host string
}

func NewClient(conf Config) *Client {
	return &Client{
		host: conf.Host,
	}
}

func (c *Client) GenerateStream(req *aicli.GenerateRequest, output io.Writer) (resp aicli.Message, err error) {
	if len(req.Messages) == 0 {
		return nil, errors.New("need a message")
	}
	reqStruct := GenerateRequest{
		Model:   req.Model,
		Prompt:  req.Messages[len(req.Messages)-1].Content(), // TODO: obviously we need to fix this
		Options: GenerateOptions{Temperature: req.Temperature},
		Stream:  true,
	}
	if req.Messages[0].Role() == aicli.RoleSystem {
		reqStruct.System = req.Messages[0].Content()
	}
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	if err := enc.Encode(reqStruct); err != nil {
		return nil, errors.Wrap(err, "encoding request")
	}
	httpResp, err := http.Post(fmt.Sprintf("%s/api/generate", c.host), "application/json", buf)
	if err != nil {
		return nil, errors.Wrap(err, "making POST")
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode > 299 {
		bod, _ := io.ReadAll(httpResp.Body)
		return nil, errors.Errorf("bad status %v, bod: '%s'", httpResp.Status, bod)
	}

	respStruct := GenerateResponse{}
	dec := json.NewDecoder(httpResp.Body)
	respMsgBldr := &strings.Builder{}
	for !respStruct.Done {
		if err := dec.Decode(&respStruct); err != nil {
			return nil, errors.Wrap(err, "unmarshaling body")
		}
		respMsgBldr.WriteString(respStruct.Response)
		if _, err := output.Write([]byte(respStruct.Response)); err != nil {
			return nil, errors.Wrap(err, "writing output")
		}
	}

	return aicli.SimpleMsg{
		RoleField:    aicli.RoleAssistant,
		ContentField: respMsgBldr.String(),
	}, nil
}

type GenerateRequest struct {
	Model    string          `json:"model"`
	Prompt   string          `json:"prompt"`
	Format   string          `json:"format,omitempty"`   // the format to return a response in. Currently the only accepted value is json
	Options  GenerateOptions `json:"options,omitempty"`  // additional model parameters listed in the documentation for the Modelfile such as temperature
	System   string          `json:"system,omitempty"`   // system prompt to (overrides what is defined in the Modelfile)
	Template string          `json:"template,omitempty"` // the full prompt or prompt template (overrides what is defined in the Modelfile)
	Context  string          `json:"context,omitempty"`  // the context parameter returned from a previous request to /generate, this can be used to keep a short conversational memory
	Stream   bool            `json:"stream"`             // if false the response will be returned as a single response object, rather than a stream of objects
	Raw      string          `json:"raw,omitempty"`      // if true no formatting will be applied to the prompt and no context will be returned. You may choose to use the raw parameter if you are specifying a full templated prompt in your request to the API, and are managing history yourself.
}

type GenerateOptions struct {
	Temperature float64 `json:"temperature"`

	// like a billion others, see https://github.com/jmorganca/ollama/blob/main/docs/api.md
}

type GenerateResponse struct {
	Model              string `json:"model"`
	CreatedAt          string `json:"created_at"`
	Response           string `json:"response"`
	Context            []int  `json:"context"`
	Done               bool   `json:"done"`
	TotalDuration      time.Duration
	LoadDuration       time.Duration `json:"load_duration"`
	SampleCount        int           `json:"sample_count"`
	SampleDuration     time.Duration `json:"sample_duration"`
	PromptEvalCount    int           `json:"prompt_eval_count"`
	PromptEvalDuration time.Duration `json:"prompt_eval_duration"`
	EvalCount          int           `json:"eval_count"`
	EvalDuration       time.Duration `json:"eval_duration"`
}
