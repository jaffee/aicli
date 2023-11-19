[![Go Report Card](https://goreportcard.com/badge/github.com/jaffee/aicli)](https://goreportcard.com/report/github.com/jaffee/aicli) [![Go Coverage](https://github.com/jaffee/aicli/wiki/coverage.svg)](https://raw.githack.com/wiki/jaffee/aicli/coverage.html) 

# aicli

aicli is a command line interface for AI chatbots. Currently only OpenAI is supported. Think of it like ChatGPT, but in your terminal instead of in the browser. You need an OpenAI API key to use it.

I threw this together so that I could access GPT-4 and pay only for my usage without paying $20/month flat rate for access to the web version. Also I tend to live in the terminal, so this is more convenient than tracking down a browser tab, logging in, etc.

## Example

```
$ aicli
> say test and nothing else
Test.
> ok, but lower case and no period
test
> \messages
     user: say test and nothing else
assistant: Test.
     user: ok, but lower case and no period
assistant: test
> repeat your last response
test
> \reset
> repeat your last response
I'm an AI language model, and each interaction with me doesn't necessarily carry context from previous conversations unless it's within the same session. If you are looking for a continuation of a previous conversation, please provide some context or restate your question, and I'll do my best to help you. If this is the same session, please simply scroll up to see the previous responses, as they should be visible to you above.
>
```

## Usage

Whatever you type into the prompt will be sent to the AI, unless it's a meta command. The meta commands are all prefixed with a backslash:

- `\reset` Reset the conversation history to a blank slate. This leaves the system message.
- `\reset-system` Removes the system message.
- `\messages` Print out the entirety of the current conversation. (After a reset, this is blank except for system message if any)
- `\config` Prints out aicli's configuration.
- `\file <filepath>` Send the path and contents of a file on your local filesystem to the AI. It will be prefixed with a short message explaining that you'll refer to the file later. The AI should just respond with something like "ok".
- `\system <message>` Prepends a system message to the list of messages (or replaces if one is already there). Does not send anything to the AI, but the new system message will be sent with the next message.
- `\set <param> <value>` Set various config params. See `\config`.


## Configuration

Flags are below. Any flag can also be set as an environment variable, just make it all uppercase and replace dashes with underscores.

```
$ aicli -help
Usage of aicli:
  -ai string
    	Name of service (default "openai")
  -context-limit int
    	Maximum number of bytes of context to keep. Earlier parts of the conversation are discarded. (default 10000)
  -model string
    	Name of model to talk to. Most services have multiple options. (default "gpt-3.5-turbo")
  -openai-api-key string
    	Your API key for OpenAI.
  -temperature float
    	Passed to model, higher numbers tend to generate less probable responses. (default 0.7)
  -verbose
    	Enables debug output.
```

## Features

- Saves input history, standard readline support (reverse search, up and down, beginning and end of line, etc).
- Resettable message history.
- Streaming responses.
- Can choose temperature and model with command line arguments.
- Can send local files.
- Ability to set system message
- support an automatic context limit

## Future/TODO

- Need to use "model" param after startup
- support other services like Anthropic, Cohere
- Write conversation, or single response to file
- automatically save conversations and allow listing/loading of convos
- Load old conversation from file
- copy response to clipboard
- functions
- other OpenAI features?
- I dunno... open an issue if something interests you.

