package main

import (
	"log"

	"github.com/jaffee/aicli/pkg/aicli"
	"github.com/jaffee/aicli/pkg/openai"
	"github.com/jaffee/commandeer"
)

func main() {
	cmd := aicli.NewCmd(nil)
	err := commandeer.LoadEnv(cmd, "", func(a interface{}) error { return nil })
	if err != nil {
		log.Fatal(err)
	}
	client := openai.NewClient(cmd.OpenAIAPIKey, cmd.OpenAIModel)
	cmd.SetAI(client)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
