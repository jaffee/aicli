package aicli

import "io"

type AI interface {
	StreamResp(req *GenerateRequest, output io.Writer) (Message, error)
}

type GenerateRequest struct {
	Model       string
	Temperature float64
	Messages    []Message
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
