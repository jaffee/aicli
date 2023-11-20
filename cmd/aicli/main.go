package main

import (
	"log"

	"github.com/jaffee/aicli/pkg/aicli"
	"github.com/jaffee/aicli/pkg/ollama"
	"github.com/jaffee/aicli/pkg/openai"
	"github.com/jaffee/commandeer"
)

func main() {
	flags := NewFlags()
	err := commandeer.LoadEnv(flags, "", func(a interface{}) error { return nil })
	if err != nil {
		log.Fatal(err)
	}

	cmd := flags.Cmd
	client := openai.NewClient(flags.OpenAI)
	cmd.AddAI("openai", client)
	cmd.AddAI("echo", &aicli.Echo{})
	cmd.AddAI("ollama", ollama.NewClient(flags.Ollama))
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func NewFlags() *Flags {
	return &Flags{
		OpenAI: openai.NewConfig(),
		Ollama: ollama.NewConfig(),
		Cmd:    *aicli.NewCmd(),
	}
}

type Flags struct {
	OpenAI openai.Config `flag:"!embed"`
	Ollama ollama.Config `flag:"!embed"`
	Cmd    aicli.Cmd     `flag:"!embed"`
}
