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
whaeh "summarize this" --file README.md
```

## Configuration

Configuration is read from the OS config directory, under `whaeh/config.yaml`.
On Unix, `XDG_CONFIG_HOME` is honored.

Interactive configuration supports multiple profiles:

```yaml
active_profile: default
profiles:
  default:
    provider: openai
    model: gpt-4o
    api_key: sk-random
    base_url:
    system_prompt: You are a helpful assistant.
```

Run `whaeh config` to manage profiles:

- Use `↑`/`↓` to select a profile.
- Select `+ Create new profile` to add one.
- Selected profiles show `Set Active`, `Edit`, and `Delete` actions.
- Use `←`/`→` to select an action, then press `Enter`.
- Select `Exit` or press `Esc` to leave interactive configuration.

When no profiles exist, the selector opens with `+ Create new profile`
highlighted. A single profile becomes active automatically. Deleting the
active profile selects the first remaining profile. If no profiles remain, the
selector opens again without starting profile creation automatically.

Profile creation and editing prompt for profile name, provider, API key, model,
base URL, and system prompt. Provider uses a dropdown. API key input is hidden.
Press `Esc` during creation or editing to cancel without saving changes.

The default base URLs are:

| Provider | Base URL |
| --- | --- |
| Ollama | `http://localhost:11434` |
| Gemini | `https://generativelanguage.googleapis.com/v1beta` |
| OpenAI | `https://api.openai.com/v1` |

## Command reference

```text
whaeh [query]                         Ask about stdin or a file
whaeh [query] --file, -f <path>       Read context from a file
whaeh [query] --raw, -r               Disable terminal markdown rendering
whaeh config                          Manage interactive profiles
whaeh version                         Print the version
```

Output is always raw when stdout is not a TTY.

## Development

```sh
go test ./...
go build -o whaeh .
```
