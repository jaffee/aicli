package main

import (
	"log"

	"github.com/jaffee/aicli/pkg/aicli"
	"github.com/jaffee/commandeer"
)

func main() {
	cmd := aicli.NewCmd()
	err := commandeer.LoadEnv(cmd, "", func(a interface{}) error { return nil })
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
