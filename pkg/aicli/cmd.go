package aicli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/wader/readline"
)

const (
	maxFileSize = 10000
)

type Cmd struct {
	OpenAIAPIKey string  `flag:"openai-api-key" help:"Your API key for OpenAI."`
	OpenAIModel  string  `flag:"openai-model" help:"Model name for OpenAI."`
	Temperature  float64 `help:"Passed to model, higher numbers tend to generate less probable responses."`
	Verbose      bool    `help:"Enables debug output."`

	messages []Message

	stdin  io.ReadCloser
	stdout io.Writer
	stderr io.Writer

	dotAICLIDir string
	historyPath string

	client AI
}

func NewCmd(client AI) *Cmd {
	return &Cmd{
		OpenAIAPIKey: "",
		OpenAIModel:  "gpt-3.5-turbo",
		Temperature:  0.7,

		messages: []Message{},

		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,

		client: client,
	}
}

func (cmd *Cmd) SetAI(ai AI) {
	cmd.client = ai
}

func (cmd *Cmd) Run() error {
	if err := cmd.checkConfig(); err != nil {
		return errors.Wrap(err, "checking config")
	}
	if err := cmd.setupConfigDir(); err != nil {
		return errors.Wrap(err, "setting up config dir")
	}
	if err := cmd.readConfigFile(); err != nil {
		return errors.Wrap(err, "reading config file")
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:       "> ",
		HistoryFile:  cmd.getHistoryFilePath(),
		HistoryLimit: 1000000,

		Stdin:  cmd.stdin,
		Stdout: cmd.stdout,
		Stderr: cmd.stderr,
	})
	if err != nil {
		return err
	}

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
		cmd.messages = append(cmd.messages, SimpleMsg{RoleField: "user", ContentField: line})
		if err := cmd.sendMessages(); err != nil {
			cmd.errOut(err, "")
		}
	}

	return nil
}

func (cmd *Cmd) sendMessages() error {
	msg, err := cmd.client.StreamResp(cmd.messages, cmd.stdout)
	if err != nil {
		return err
	}
	cmd.messages = append(cmd.messages, msg)
	cmd.out("")
	return nil
}

func (cmd *Cmd) handleMeta(line string) {
	parts := strings.SplitN(line, " ", 2)
	var err error
	switch parts[0] {
	case `\reset`:
		cmd.messages = []Message{}
	case `\messages`:
		cmd.printMessages()
	case `\config`:
		cmd.printConfig()
	case `\file`:
		if len(parts) < 2 {
			err = errors.New("need a file name for \\file command")
		} else {
			err = cmd.sendFile(parts[1])
		}
	case `\system`:
		msg := SimpleMsg{
			RoleField:    "system",
			ContentField: parts[1],
		}
		if len(cmd.messages) == 0 || cmd.messages[0].Role() != "system" {
			// prepend a system message
			cmd.messages = append([]Message{msg}, cmd.messages...)
		} else {
			// replace the existing system message
			cmd.messages[0] = msg
		}
	default:
		err = errors.Errorf("Unknown meta command '%s'", line)
	}
	if err != nil {
		cmd.err(err)
	}
}

func (cmd *Cmd) sendFile(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return errors.Wrapf(err, "opening file '%s'", file)
	}
	if stats, err := f.Stat(); err != nil {
		return errors.Wrapf(err, "couldn't stat '%s'", file)
	} else if stats.Size() > maxFileSize {
		return errors.Wrap(err, "file too large")
	}

	b := &strings.Builder{}
	b.WriteString(fmt.Sprintf("Here is a file named '%s' that I'll refer to later, you can just say 'ok': \n```\n", file))
	if _, err := io.Copy(b, f); err != nil {
		return errors.Wrapf(err, "reading file '%s'", file)
	}
	b.WriteString("\n```\n")
	cmd.messages = append(cmd.messages, SimpleMsg{RoleField: "user", ContentField: b.String()})
	msg, err := cmd.client.StreamResp(cmd.messages, cmd.stdout)
	if err != nil {
		return errors.Wrap(err, "sending file")
	}
	cmd.messages = append(cmd.messages, msg)
	cmd.out("")
	return nil
}

func (cmd *Cmd) printMessages() {
	for _, msg := range cmd.messages {
		cmd.out("%9s: %s", msg.Role(), msg.Content())
	}
}

// isMeta returns true if the line is a meta command
func isMeta(line string) bool {
	return line[0] == '\\'
}

// out writes output back to the user on stdout. For convenience, it adds a newline.
func (cmd *Cmd) out(format string, a ...any) {
	fmt.Fprintf(cmd.stdout, format+"\n", a...)
}

func (cmd *Cmd) err(err error) {
	fmt.Fprintf(cmd.stderr, err.Error()+"\n")
}

// errOut wraps the error and writes it to the user on stderr.
func (cmd *Cmd) errOut(err error, format string, a ...any) {
	fmt.Fprintf(cmd.stderr, "%s: %v", fmt.Sprintf(format, a...), err.Error())
}

// checkConfig ensures the command configuration is valid before proceeding.
func (cmd *Cmd) checkConfig() error {
	if cmd.OpenAIAPIKey == "" {
		return errors.New("Need an API key")
	}
	return nil
}

func (cmd *Cmd) printConfig() {
	fmt.Fprintf(cmd.stderr, "OpenAI_API_Key: length=%d\n", len(cmd.OpenAIAPIKey))
	fmt.Fprintf(cmd.stderr, "OpenAIModel: %s\n", cmd.OpenAIModel)
	fmt.Fprintf(cmd.stderr, "Temperature: %f\n", cmd.Temperature)
	fmt.Fprintf(cmd.stderr, "Verbose: %v\n", cmd.Verbose)
}

func (cmd *Cmd) setupConfigDir() error {
	if cmd.dotAICLIDir != "" {
		return nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".aicli")
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	cmd.dotAICLIDir = path
	return nil
}

func (cmd *Cmd) getHistoryFilePath() string {
	if cmd.historyPath != "" {
		return cmd.historyPath
	}
	return filepath.Join(cmd.dotAICLIDir, "history")
}

func (cmd *Cmd) readConfigFile() error {
	path := filepath.Join(cmd.dotAICLIDir, "config")
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "opening file")
	}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if len(line) == 0 {
			continue
		}
		if isMeta(line) {
			cmd.handleMeta(line)
			continue
		} else if line[0] == '#' {
			continue
		} else {
			return err
		}
	}
	return nil
}
