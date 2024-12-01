package aicli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/wader/readline"
)

const (
	maxFileSize = 10000
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorCyan   = "\033[36m" // for user input
	colorYellow = "\033[33m" // for AI responses
	colorRed    = "\033[31m" // for errors
	colorBlue   = "\033[34m" // for system messages
)

type Cmd struct {
	AI           string  `help:"Name of service"`
	Model        string  `help:"Name of model to talk to. Most services have multiple options."`
	Temperature  float64 `help:"Passed to model, higher numbers tend to generate less probable responses."`
	Verbose      bool    `help:"Enables debug output."`
	ContextLimit int     `help:"Maximum number of bytes of context to keep. Earlier parts of the conversation are discarded."`
	EnableAWS    bool    `help:"Enable AWS Bedrock as an AI option. Disabled by default because it slows startup time."`
	Color        bool    `help:"Enable colored output"`

	rl *readline.Instance

	messages []Message
	totalLen int

	stdin  io.ReadCloser
	stdout io.Writer
	stderr io.Writer

	dotAICLIDir string
	historyPath string

	ais map[string]AI
}

func NewCmd() *Cmd {
	return &Cmd{
		AI:           "openai",
		Model:        "gpt-3.5-turbo",
		Temperature:  0.7,
		ContextLimit: 10000, // 10,000 bytes ~2000 tokens
		Color:        true,

		messages: []Message{},

		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,

		ais: make(map[string]AI),
	}
}

func (cmd *Cmd) AddAI(name string, ai AI) {
	cmd.ais[name] = ai
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

	var err error
	cmd.rl, err = readline.NewEx(cmd.getReadlineConfig())
	if err != nil {
		return err
	}

	for {
		line, err := cmd.rl.Readline()
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
			line = cmd.handleMeta(line)
			if line == "" {
				continue
			}
		}
		cmd.appendMessage(SimpleMsg{RoleField: "user", ContentField: line})
		if err := cmd.sendMessages(); err != nil {
			cmd.errOut(err, "")
		}
	}

	return nil
}

func (cmd *Cmd) messagesWithinLimit() []Message {
	total := 0
	for i := len(cmd.messages) - 1; i >= 0; i-- {
		total += len(cmd.messages[i].Content())
		if total > cmd.ContextLimit {
			return cmd.messages[i:]
		}
	}
	return cmd.messages
}

func (cmd *Cmd) sendMessages() error {
	req := &GenerateRequest{
		Model:       cmd.Model,
		Temperature: cmd.Temperature,
		Messages:    cmd.messagesWithinLimit(),
	}
	msg, err := cmd.client().GenerateStream(req, &colorWriter{w: cmd.stdout, color: cmd.getColor(colorYellow)})
	if err != nil {
		return err
	}
	cmd.messages = append(cmd.messages, msg)
	cmd.out("")
	return nil
}

func (cmd *Cmd) hasSystemMessage() bool {
	return len(cmd.messages) > 0 && cmd.messages[0].Role() == RoleSystem
}

func (cmd *Cmd) handleMeta(line string) string {
	parts := strings.SplitN(line, " ", 2)
	var err error
	switch parts[0] {
	case `\reset`:
		if cmd.hasSystemMessage() {
			cmd.messages = cmd.messages[:1]
		} else {
			cmd.messages = cmd.messages[:0]
		}
		cmd.totalLen = 0
	case `\reset-system`:
		if cmd.hasSystemMessage() {
			cmd.messages = cmd.messages[1:]
		}
	case `\messages`:
		cmd.printMessages()
	case `\config`:
		cmd.printConfig()
	case `\set`:
		if len(parts) != 2 {
			err = errors.New("usage: \\set <param> <value>")
			break
		}
		pv := strings.SplitN(parts[1], " ", 2)
		if len(pv) != 2 {
			err = errors.New("usage: \\set <param> <value>")
			break
		}
		param, val := pv[0], pv[1]
		err = cmd.Set(param, val)
	case `\file`:
		if len(parts) < 2 {
			err = errors.New("need a file name for \\file command")
		} else {
			err = cmd.addFile(parts[1])
		}
	case `\system`:
		if len(parts) == 1 {
			if cmd.hasSystemMessage() {
				cmd.out(cmd.messages[0].Content() + "\n")
			}
			break
		}
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
	case `\context`:
		if len(parts) != 2 {
			err = errors.New("context requires a single integer argument")
			break
		}
		var limit int
		limit, err = strconv.Atoi(parts[1])
		if err != nil {
			err = errors.Wrap(err, "argument must be a number")
			break
		}
		if limit <= 0 {
			err = errors.New("limit must be positive")
			break
		}
		cmd.ContextLimit = limit
	case `\<<`:
		until := "EOF"
		if len(parts) == 2 {
			until = parts[1]
		} else if len(parts) > 2 {
			err = errors.Errorf("1 or 0 arguments only to multiline command")
			break
		}
		var newMsg string
		newMsg, err = cmd.readMulti(until)
		if err != nil {
			err = errors.Wrap(err, "reading multi")
			break
		} else {
			return newMsg
		}
	case `\quit`, `\exit`, `\q`:
		os.Exit(0)
	case `\help`, `\?`:
		cmd.sysOut("Available commands:")
		cmd.sysOut("  \\help, \\?        Show this help message")
		cmd.sysOut("  \\reset           Reset conversation (keeps system message)")
		cmd.sysOut("  \\reset-system    Remove system message")
		cmd.sysOut("  \\messages        Show all messages in the conversation")
		cmd.sysOut("  \\config          Show current configuration")
		cmd.sysOut("  \\set             Set a configuration parameter (usage: \\set <param> <value>)")
		cmd.sysOut("  \\file            Add file contents to conversation")
		cmd.sysOut("  \\system          Set or show system message")
		cmd.sysOut("  \\context         Set context limit in bytes")
		cmd.sysOut("  \\<<              Start multiline input (optional custom terminator)")
		cmd.sysOut("  \\quit, \\exit, \\q Exit the program")
	default:
		err = errors.Errorf("Unknown meta command '%s'", line)
	}
	if err != nil {
		cmd.err(err)
	}
	return ""
}

func (cmd *Cmd) readMulti(until string) (string, error) {
	bldr := &strings.Builder{}
	cmd.rl.SetPrompt("")
	defer cmd.rl.SetPrompt("> ")
	for {
		line, err := cmd.rl.Readline()
		if err == readline.ErrInterrupt {
			return "", nil
		} else if err != nil {
			return "", errors.Wrap(err, "reading line")
		}
		if line == until {
			return bldr.String(), nil
		}
		if _, err := bldr.WriteString(line + "\n"); err != nil {
			return "", errors.Wrap(err, "building string")
		}
	}

}

func (cmd *Cmd) Set(param, value string) error {
	switch param {
	case "ai":
		if _, ok := cmd.ais[value]; !ok {
			return errors.Errorf("unknown ai '%s'", value)
		}
		cmd.AI = value
	case "model":
		cmd.Model = value
	case "temp", "temperature":
		temp, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return errors.Wrapf(err, "parsing '%s' to float", value)
		}
		cmd.Temperature = temp
	case "verbose":
		switch strings.ToLower(value) {
		case "1", "true", "yes":
			cmd.Verbose = true
		case "0", "false", "no":
			cmd.Verbose = false
		default:
			return errors.Errorf("could not parse '%s' as bool", value)
		}
	case "context":
		lim, err := strconv.Atoi(value)
		if err != nil {
			return errors.Wrapf(err, "parsing '%s' to int", value)
		}
		cmd.ContextLimit = lim
	case "color":
		switch strings.ToLower(value) {
		case "1", "true", "yes", "on":
			cmd.Color = true
		case "0", "false", "no", "off":
			cmd.Color = false
		default:
			return errors.Errorf("could not parse '%s' as bool", value)
		}
		// Update readline config with new color setting
		var err error
		cmd.rl, err = readline.NewEx(cmd.getReadlineConfig())
		if err != nil {
			return errors.Wrap(err, "updating readline config")
		}
	default:
		return errors.Errorf("unknown parameter '%s'", param)
	}
	return nil
}

func (cmd *Cmd) addFile(file string) error {
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
	b.WriteString(fmt.Sprintf("Here is a file named '%s' that I'll refer to later:\n```\n", file))
	if _, err := io.Copy(b, f); err != nil {
		return errors.Wrapf(err, "reading file '%s'", file)
	}
	b.WriteString("\n```\n")
	cmd.appendMessage(SimpleMsg{RoleField: "user", ContentField: b.String()})
	return nil
}

func (cmd *Cmd) client() AI {
	return cmd.ais[cmd.AI]
}

func (cmd *Cmd) appendMessage(msg Message) {
	cmd.messages = append(cmd.messages, msg)
	cmd.totalLen += len(msg.Content())
}

func (cmd *Cmd) printMessages() {
	for _, msg := range cmd.messages {
		if msg.Role() == "user" {
			cmd.out(cmd.getColor(colorCyan)+"%9s: %s"+cmd.getColor(colorReset), msg.Role(), msg.Content())
		} else {
			cmd.out(cmd.getColor(colorYellow)+"%9s: %s"+cmd.getColor(colorReset), msg.Role(), msg.Content())
		}
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

func (cmd *Cmd) sysOut(format string, a ...any) {
	fmt.Fprintf(cmd.stdout, cmd.getColor(colorBlue)+format+cmd.getColor(colorReset)+"\n", a...)
}

func (cmd *Cmd) err(err error) {
	fmt.Fprintf(cmd.stderr, cmd.getColor(colorRed)+"%s"+cmd.getColor(colorReset)+"\n", err.Error())
}

// errOut wraps the error and writes it to the user on stderr.
func (cmd *Cmd) errOut(err error, format string, a ...any) {
	fmt.Fprintf(cmd.stderr, cmd.getColor(colorRed)+"%s: %v"+cmd.getColor(colorReset)+"\n", fmt.Sprintf(format, a...), err.Error())
}

// checkConfig ensures the command configuration is valid before proceeding.
func (cmd *Cmd) checkConfig() error {
	if cmd.client() == nil {
		return errors.Errorf("have no AI named '%s' configured", cmd.AI)
	}
	return nil
}

func (cmd *Cmd) printConfig() {
	fmt.Fprintf(cmd.stderr, "AI: %s\n", cmd.AI)
	fmt.Fprintf(cmd.stderr, "Model: %s\n", cmd.Model)
	fmt.Fprintf(cmd.stderr, "Temperature: %f\n", cmd.Temperature)
	fmt.Fprintf(cmd.stderr, "Verbose: %v\n", cmd.Verbose)
	fmt.Fprintf(cmd.stderr, "ContextLimit: %d\n", cmd.ContextLimit)
	fmt.Fprintf(cmd.stderr, "Color: %v\n", cmd.Color)
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

func (cmd *Cmd) getReadlineConfig() *readline.Config {
	return &readline.Config{
		Prompt:       "> ",
		HistoryFile:  cmd.getHistoryFilePath(),
		HistoryLimit: 1000000,
		Stdin:        cmd.stdin,
		Stdout:       &colorWriter{w: cmd.stdout, color: cmd.getColor(colorCyan)},
		Stderr:       cmd.stderr,
	}
}

// colorWriter wraps an io.Writer and adds color codes
type colorWriter struct {
	w     io.Writer
	color string
}

func (cw *colorWriter) Write(p []byte) (n int, err error) {
	if cw.color == "" {
		return cw.w.Write(p)
	}
	// Write the color code, then the content, then reset
	if _, err := cw.w.Write([]byte(cw.color)); err != nil {
		return 0, err
	}
	n, err = cw.w.Write(p)
	if err != nil {
		return n, err
	}
	_, err = cw.w.Write([]byte(colorReset))
	return n, err
}

func (cmd *Cmd) getColor(col string) string {
	if !cmd.Color {
		return ""
	}
	return col
}
