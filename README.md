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

## Configuration

Flags are below. Any flag can also be set as an environment variable, just make it all uppercase and replace dashes with underscores.

```
$ aicli -help
Usage of aicli:
  -openai-api-key string
    	Your API key for OpenAI.
  -openai-model string
    	Model name for OpenAI. (default "gpt-3.5-turbo")
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

## Future 

- Send whole files
- Write conversation, or single response to file
- Load old conversation from file
- functions
- other OpenAI features?
- I dunno... open an issue if something interests you.
