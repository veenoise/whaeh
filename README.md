<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="assets/logo-dark.png">
    <img src="assets/logo.png" width="220" alt="Whaeh logo">
  </picture>
</p>

<h1 align="center">whaeh</h1>

`whaeh` is a Unix-style CLI for asking an AI about piped input or files.

```sh
go run . --provider ollama --model llama3.2 "find the relevant command"
man tar | go run . "how do I create a gzipped archive?"
go run . --file README.md "summarize this"
```

The default provider is Ollama at `http://localhost:11434`. Configure other
providers with environment variables or the config command:

```sh
export WHAEH_PROVIDER=gemini
export GEMINI_API_KEY=...
export WHAEH_MODEL=gemini-2.5-flash

whaeh config set provider openai
whaeh config set model gpt-4o-mini
whaeh config set api_key "$OPENAI_API_KEY"
```

Configuration is read from `~/.whaeh/config.yaml` (or
`$XDG_CONFIG_HOME/whaeh/config.yaml`). `--raw` disables terminal markdown
rendering; output is always raw when stdout is not a TTY.

## Build and test

```sh
go test ./...
go build -o whaeh .
```
