# BubbleChat

BubbleChat is a prototype that uses the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework for a terminal-based program that interacts with the Gemini LLM. For interacting with the LLM, it uses [go-genai](https://github.com/googleapis/go-genai).

The chat interface supports both interacting with the LLM and using tools like `kubectl`, assuming they are present and configured on the machine. If the prompt from the user calls for a tool callout, the tool's execution is displayed in green on the terminal.

## Gemini Key

The Gemini key can be provided via the `GEMINI_API_KEY` environment variable or by specifying `GEMINI_API_KEY` in the `.env` file. The program uses [godotenv](https://github.com/joho/godotenv) to load the `.env` file if it exists.

## Building

The common building and test tasks are done via the `Taskfile.yml`. If you do not have it installed but have Go, the easiest way to install it is via:

```
go install github.com/go-task/task/v3/cmd/task@latest
```

## Usage

To run the application, use the following command:

```
task run
```

Make sure to set your `GEMINI_API_KEY` in your environment or in the `.env` file before running the application.