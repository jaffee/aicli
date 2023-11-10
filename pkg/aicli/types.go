package aicli

import "io"

type AI interface {
	StreamResp(msgs []Message, output io.Writer) (Message, error)
}

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
