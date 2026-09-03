# Gemini for Go Developers

Companion source code and runnable examples for the **[Gemini for Go Developers](https://danicat.dev/series/gemini-for-go-developers/)** tutorial series.

---

## Repository Structure

```text
.
└── part-3/
    ├── 01-sdk-agent/     # Low-level agent loop using google.golang.org/genai
    ├── 02-genkit-flow/    # Flow-centric pipeline using Firebase Genkit Go
    └── 03-adk-agent/      # Modular multi-agent using Google Agent Development Kit (ADK)
```

---

## Part 3: Building Autonomous Agents in Go

In [Part 3](https://danicat.dev/posts/20260825-gemini-for-go-developers-part-3-building-agents/), we build a **Retro Game Appraiser** across three distinct paradigms in Go. The agent inspects a collector's personal inventory to check for owned titles, queries live market valuations with Google Search grounding, and provides actionable purchase advice.

### Prerequisites

All examples require a Google Cloud Project with the Vertex AI / Gemini API enabled and Application Default Credentials configured:

```bash
# Authenticate with Google Cloud
gcloud auth application-default login

# Export your Google Cloud Project ID
export GOOGLE_CLOUD_PROJECT="your-project-id"
```

---

### 1. Low-Level Agent Loop (`part-3/01-sdk-agent`)

Built directly with the **[Go GenAI SDK](https://pkg.go.dev/google.golang.org/genai)** (`google.golang.org/genai`), implementing the core agent control loop (Reasoning $\rightarrow$ Tool Dispatch $\rightarrow$ Observation $\rightarrow$ Synthesis).

```bash
cd part-3/01-sdk-agent
go run main.go
```

---

### 2. Flow-Centric Pipeline (`part-3/02-genkit-flow`)

Built with **[Genkit Go](https://genkit.dev)** (`github.com/firebase/genkit/go`), defining structured schemas, strongly typed tools, and an HTTP API endpoint with graceful shutdown.

```bash
cd part-3/02-genkit-flow
go run main.go
```

Send a test request to the endpoint:

```bash
curl -X POST http://localhost:8080/api/appraise \
  -H "Content-Type: application/json" \
  -d '{"data": {"prompt": "I found a copy of Chrono Trigger for SNES for $250. Is this a good deal?"}}'
```

---

### 3. Modular Multi-Agent System (`part-3/03-adk-agent`)

Built with Google's **[Agent Development Kit (ADK)](https://adk.dev)**, using session-driven state management, event streaming runners, and modular function tools.

Run the agent in interactive console mode:

```bash
cd part-3/03-adk-agent
go run main.go
```

Or launch the developer web interface and REST API:

```bash
go run main.go web webui api
```

Open `http://localhost:8080` in your browser to interact with the agent via the ADK Web UI.

---

## License

This project is licensed under the [Apache 2.0 License](LICENSE).
