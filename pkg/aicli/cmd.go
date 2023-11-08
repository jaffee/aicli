package aicli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/wader/readline"

	openai "github.com/sashabaranov/go-openai"
)

type Cmd struct {
	OpenAI_API_Key string  `flag:"openai-api-key" help:"Your API key for OpenAI."`
	OpenAIModel    string  `flag:"openai-model" help:"Model name for OpenAI."`
	Temperature    float64 `help:"Passed to model, higher numbers tend to generate less probable responses."`
	Verbose        bool    `help:"Enables debug output."`

	messages []openai.ChatCompletionMessage

	stdin  io.ReadCloser
	stdout io.Writer
	stderr io.Writer
}

func NewCmd() *Cmd {
	return &Cmd{
		OpenAI_API_Key: "",
		OpenAIModel:    "gpt-3.5-turbo",
		Temperature:    0.7,

		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

func (cmd *Cmd) Run() error {
	if err := cmd.checkConfig(); err != nil {
		return errors.Wrap(err, "checking config")
	}

	client := openai.NewClient(cmd.OpenAI_API_Key)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:       "> ",
		HistoryFile:  getHistoryFilePath(),
		HistoryLimit: 1000000,
		// DisableAutoSaveHistory: true,

		Stdin:  cmd.stdin,
		Stdout: cmd.stdout,
		Stderr: cmd.stderr,
	})
	if err != nil {
		return err
	}

	cmd.messages = []openai.ChatCompletionMessage{}

	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			continue
		} else if err == io.EOF {
			break
		} else if err != nil {
			return errors.Wrap(err, "reading line")
		} else if len(line) == 0 {
			continue
		}
		if isMeta(line) {
			cmd.handleMeta(line)
			continue
		}
		cmd.messages = append(cmd.messages,
			openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: line,
			})
		cmd.debugOut("sending: '%s' with %d total messages\n", line, len(cmd.messages))
		resp, err := client.CreateChatCompletion(context.Background(),
			openai.ChatCompletionRequest{
				Model:    cmd.OpenAIModel,
				Messages: cmd.messages,
				ResponseFormat: openai.ChatCompletionResponseFormat{
					Type: openai.ChatCompletionResponseFormatTypeText,
				},
			},
		)
		if err != nil {
			cmd.errOut(err, "making chat request")
			continue
		}
		cmd.debugOut("%+v\n", resp)
		cmd.messages = append(cmd.messages, resp.Choices[0].Message)
		cmd.out(resp.Choices[0].Message.Content)
	}

	return nil
}

func (cmd *Cmd) handleMeta(line string) {
	line = strings.TrimSpace(line)
	switch line {
	case `\reset`:
		cmd.messages = []openai.ChatCompletionMessage{}
	case `\messages`:
		cmd.printMessages()
	case `\config`:
		cmd.printConfig()
	default:
		cmd.errMsg("Unknown meta command '%s'", line)
	}
}

func (cmd *Cmd) printMessages() {
	for _, msg := range cmd.messages {
		cmd.out("%9s: %s", msg.Role, msg.Content)
	}
}

// isMeta returns true if the line is a meta command
func isMeta(line string) bool {
	return line[0] == '\\'
}

// out writes output back to the user.
func (cmd *Cmd) out(format string, a ...any) {
	fmt.Fprintf(cmd.stdout, format+"\n", a...)
}

// errOut wraps the error and writes it to the user on stderr.
func (cmd *Cmd) errOut(err error, format string, a ...any) {
	fmt.Fprintf(cmd.stderr, "%s: %v", fmt.Sprintf(format, a...), err.Error())
}

// errMsg writes the message to the user on stderr.
func (cmd *Cmd) errMsg(format string, a ...any) {
	fmt.Fprintf(cmd.stderr, format+"\n", a...)
}

// checkConfig ensures the command configuration is valid before proceeding.
func (cmd *Cmd) checkConfig() error {
	if cmd.OpenAI_API_Key == "" {
		return errors.New("Need an API key")
	}
	return nil
}

// debugOut writes to stderr if the verbose flag is set.
func (cmd *Cmd) debugOut(format string, a ...any) {
	if !cmd.Verbose {
		return
	}
	fmt.Fprintf(cmd.stderr, format, a...)
}

func (cmd *Cmd) printConfig() {
	fmt.Fprintf(cmd.stderr, "OpenAI_API_Key: %s...\n", cmd.OpenAI_API_Key[:2])
	fmt.Fprintf(cmd.stderr, "OpenAIModel: %s\n", cmd.OpenAIModel)
	fmt.Fprintf(cmd.stderr, "Temperature: %f\n", cmd.Temperature)
	fmt.Fprintf(cmd.stderr, "Verbose: %v\n", cmd.Verbose)
}

func getHistoryFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// if we can't get a home dir, we'll use the local directory
		return ".aicli_history"
	}
	return filepath.Join(home, ".aicli_history")
}
