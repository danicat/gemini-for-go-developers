# Retro Game Appraiser: Agent Development Kit (ADK)

This example demonstrates how to build a production-grade autonomous agent using Google's **[Agent Development Kit (ADK)](https://adk.dev)** (`google.golang.org/adk`).

The agent is constructed using ADK's modular component model:
* **LLM Agent**: Defined with `llmagent.New`, backed by Gemini 3.8 Flash (`gemini-3.8-flash`) and Gemini Enterprise.
* **Function Tools**: Registered with `functiontool.New` to query the collector's inventory catalog.
* **Search Grounding**: Integrated server-side via `geminitool.GoogleSearch{}`.
* **Canonical Launcher**: Built with `full.NewLauncher()`, supporting both interactive terminal sessions and full web hosting.

## Prerequisites

1. Authenticate with Google Cloud using Application Default Credentials (ADC):
   ```bash
   gcloud auth application-default login
   ```

2. Export your Google Cloud project ID:
   ```bash
   export GOOGLE_CLOUD_PROJECT="your-project-id"
   ```

## Running the Agent

The ADK launcher supports two primary modes:

### 1. Interactive Console Mode (Default)

Launch an interactive chat session in your terminal:

```bash
go run main.go
# or explicitly:
go run main.go console
```

Interact with the agent directly in the terminal prompt. Type `exit` or press `Ctrl+C` to quit.

### 2. Web UI & REST API Mode

Launch the web server with the developer browser UI and REST endpoints mounted:

```bash
go run main.go web webui api
```

Once running:
* **Developer Web UI**: Open [http://localhost:8080/ui/](http://localhost:8080/ui/) in your browser.
* **REST API**: Available at [http://localhost:8080/api](http://localhost:8080/api) (e.g., [http://localhost:8080/api/list-apps](http://localhost:8080/api/list-apps)).

To run on a different port:

```bash
go run main.go web -port 8081 webui api
```

### 3. Headless REST API Mode (Production)

To deploy as a headless API microservice without mounting the browser UI:

```bash
go run main.go web api
```
