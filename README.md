<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="./assets/logo.png">
    <img src="./assets/logo.png" width="220" alt="Whaeh logo">
  </picture>
</p>

<h1 align="center">whaeh</h1>

`whaeh` is a Unix-style CLI for asking an AI about piped input or files.

## Installation

### Requirements

- Go 1.23 or later
- An AI provider:
  - [Ollama](https://ollama.com/) running locally (the default)
  - [Google Gemini](https://ai.google.dev/)
  - [OpenAI](https://platform.openai.com/)

### Install with Go

```sh
go install github.com/veenoise/whaeh@latest
```

The `whaeh` executable is installed in Go's binary directory. Make sure that
directory is on your `PATH`, then verify the installation:

```sh
whaeh version
```

### Install from source

```sh
git clone https://github.com/veenoise/whaeh.git
cd whaeh
go install .
```

To build a local executable instead:

```sh
go build -o whaeh .
```

## Quick start

Ollama is the default provider and uses `http://localhost:11434`. Make sure
Ollama is running and that the selected model is available:

```sh
ollama pull llama3.2
whaeh "find the relevant command"
```

Ask about piped input or a file:

```sh
man tar | whaeh "how do I create a gzipped archive?"
whaeh --file README.md "summarize this"
```

For a one-off provider or model, use command-line flags:

```sh
whaeh --provider ollama --model llama3.2 "explain this"
```

## Configuration

Configuration is read from `~/.whaeh/config.yaml`, or from
`$XDG_CONFIG_HOME/whaeh/config.yaml` when `XDG_CONFIG_HOME` is set.

### Configure with the command line

```sh
whaeh config set provider openai
whaeh config set model gpt-4o-mini
whaeh config set api_key "$OPENAI_API_KEY"
```

Supported settings are `provider`, `model`, `api_key`, and `base_url`.

### Configure with environment variables

```sh
export WHAEH_PROVIDER=gemini
export WHAEH_MODEL=gemini-2.5-flash
export GEMINI_API_KEY="your-api-key"
```

API keys are read from `GEMINI_API_KEY` or `OPENAI_API_KEY` when the selected
provider does not have an `api_key` saved in the configuration file. Ollama
does not require an API key.

The default base URLs are:

| Provider | Base URL |
| --- | --- |
| Ollama | `http://localhost:11434` |
| Gemini | `https://generativelanguage.googleapis.com/v1beta` |
| OpenAI | `https://api.openai.com/v1` |

Override a provider's base URL with `WHAEH_BASE_URL` or:

```sh
whaeh config set base_url "https://example.com/v1"
```

## Command reference

```text
whaeh [query]                         Ask about stdin or a file
whaeh --file, -f <path> [query]       Read context from a file
whaeh --provider, -p <name> [query]   Override the provider
whaeh --model, -m <name> [query]      Override the model
whaeh --raw [query]                   Disable terminal markdown rendering
whaeh config set <key> <value>        Save a configuration value
whaeh version                         Print the version
```

Output is always raw when stdout is not a TTY.

## Development

```sh
go test ./...
go build -o whaeh .
```
