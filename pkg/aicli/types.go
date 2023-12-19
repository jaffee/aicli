package aicli

import "io"

type AI interface {
	// TODO add context to these
	GenerateStream(req *GenerateRequest, output io.Writer) (Message, error)
	GetEmbedding(req *EmbeddingRequest) ([]Embedding, error)
	// ListModels
}

type GenerateRequest struct {
	Model       string
	Temperature float64
	// MaxGenLen is the maximum number of tokens the model should generate. Only
	// supported by some models.
	MaxGenLen int
	Messages  []Message
}

type EmbeddingRequest struct {
	Inputs []string
	Model  string
}

type Embedding struct {
	Embedding []float32
}

const (
	RoleAssistant = "assistant"
	RoleUser      = "user"
	RoleSystem    = "system"
)

type Message interface {
	Role() string
	Content() string
}

type SimpleMsg struct {
	RoleField    string
	ContentField string
}

func (s SimpleMsg) Role() string    { return s.RoleField }
func (s SimpleMsg) Content() string { return s.ContentField }
