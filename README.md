# Gemini for Go Developers

Companion source code and runnable examples for the **[Gemini for Go Developers](https://danicat.dev/series/gemini-for-go-developers/)** tutorial series.

---

## Repository Structure

```text
.
├── part-3/
│   ├── 01-sdk-agent/      # Low-level agent loop using google.golang.org/genai
│   ├── 02-genkit-flow/    # Flow-centric pipeline using Firebase Genkit Go
│   └── 03-adk-agent/      # Modular multi-agent using Google Agent Development Kit (ADK)
├── part-4/
│   ├── 01-flow/           # Stateless diagnostic flow with dotprompt, retry & Agent Skills
│   ├── 02-agent/          # Stateful multi-turn terminal & HTTP agent (genkitx.DefineAgent)
│   ├── 03-multi-agent/    # Multi-agent delegation (aida -> inspector & researcher)
│   └── 04-interruptions/  # Human-in-the-loop tool interrupts
└── part-6/
    └── ...                # Hardened public Cloud Run agent (budgets, HTTP gates & Model Armor)
```

---

## Part 3: Building Autonomous Agents in Go

In [Part 3](https://danicat.dev/posts/gemini-for-go-developers-part-3-building-agents/), we build a **Retro Game Appraiser** across three distinct paradigms in Go. The agent inspects a collector's personal inventory to check for owned titles, queries live market valuations with Google Search grounding, and provides actionable purchase advice.

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

Open `http://localhost:8080/ui/` in your browser to interact with the agent via the ADK Web UI.

---

## Part 4: Building Agentic Backends with Genkit

In [Part 4](https://danicat.dev/posts/building-agentic-backends-with-genkit-go/), we rebuild **AIDA** (the AI Diagnostic Agent powered by [osquery](https://osquery.io/)) in Go with Genkit, progressing from a single-shot typed flow to a stateful multi-agent backend:

```bash
export GOOGLE_CLOUD_PROJECT="your-project-id"
export GOOGLE_CLOUD_LOCATION="global"
```

- **`part-4/01-flow`**: Single-shot `diagnose` flow using embedded `dotprompt` (`prompts/aida.prompt`), typed schemas (`DefineSchemaFor` + `LookupDataPrompt`), `runOsquery` tool, and OS-specific `osquery` Agent Skills (`darwin`, `linux`, `windows`).
- **`part-4/02-agent`**: Stateful multi-turn agent (`genkitx.DefineAgent`) backed by an in-memory session store (`localstore.NewInMemorySessionStore`) with both an interactive REPL and HTTP endpoints (`genkitx.AllAgentRoutes`).
- **`part-4/03-multi-agent`**: Multi-agent delegation via `middlewarex.Agents`, splitting AIDA's work between `inspector` (host telemetry via `osquery`) and `researcher` (web lookup via Google Search grounding).

---

## Part 5: Designing Agentic Interfaces with A2UI and Genkit

In [Part 5](https://danicat.dev/posts/designing-agentic-interfaces-with-a2ui-and-genkit/), AIDA graduates to a custom [A2UI](https://a2ui.org/) diagnostic cockpit with human-in-the-loop tool interrupts. The full desktop and web cockpit lives in its own repository: [`danicat/go-aida`](https://github.com/danicat/go-aida).

---

## Part 6: How to Ship a Public AI Agent Without Breaking the Budget

In [Part 6](https://danicat.dev/posts/chibi-dani-architecture/), we harden **Chibi Dani** for unauthenticated public traffic on Cloud Run using minimal token/compute budgets, least-privilege IAM, HTTP perimeter middleware, Google Cloud Model Armor, and PII-redacted Firestore telemetry.

```bash
cd part-6
go test -count=1 ./...

export GOOGLE_CLOUD_PROJECT="your-project-id"
export MODEL_ARMOR_TEMPLATE="projects/your-project-id/locations/us-central1/templates/chibi-dani-template"
go run .
```

See [`part-6/README.md`](part-6/README.md) for local `curl` examples and setup notes.

---

## License

This project is licensed under the [Apache 2.0 License](LICENSE).

## Disclaimer

This is not an officially supported Google product. It is an experimental project created for educational and development purposes.
