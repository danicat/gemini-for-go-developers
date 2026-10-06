# AIDA Stateful Diagnostic Agent (`part-4/02-agent`)

This example demonstrates how to build **AIDA (AI Diagnostic Agent)** as a stateful, multi-turn agent using Genkit Go's experimental agent API (`genkit/exp`, `ai/exp`, and `ai/exp/localstore`).

Key features demonstrated:
* **Declarative Agent Definition**: Registers `aida` using `genkitx.DefineAgent` with `aix.InlinePrompt`, `ai.WithTools(runOsquery)`, and `ai.WithUse(&middleware.Retry{MaxRetries: 2}, &middleware.Skills{SkillPaths: []string{"./skills"}})`.
* **Server-Managed Session Persistence**: Uses `localstore.NewInMemorySessionStore[struct{}]()` and `aix.WithSessionStore(store)` to persist multi-turn conversation history across turns via `aix.WithSessionID[struct{}](sessionID)`.
* **Interactive Terminal REPL & HTTP Server**: Runs an interactive `AIDA>` terminal loop while simultaneously serving agent HTTP endpoints (`/agents/aida`, `/agents/aida/getSnapshot`, `/agents/aida/abort`) via `genkitx.AllAgentRoutes(g)` and `server.Start`.

## Prerequisites

1. Install [`osquery`](https://osquery.io/) so `osqueryi` is available on your `PATH`.
2. Authenticate with Google Cloud using Application Default Credentials (ADC):
   ```bash
   gcloud auth application-default login
   ```
3. Export your Google Cloud project ID and location:
   ```bash
   export GOOGLE_CLOUD_PROJECT="your-project-id"
   export GOOGLE_CLOUD_LOCATION="global"
   ```

## Interactive Terminal Mode

Start the interactive multi-turn diagnostic REPL:

```bash
go run .
```

Example multi-turn session:

```text
AIDA> Which top 3 processes are using the most resident memory right now?
...
AIDA> What is the PID and command line of the #1 process from that list?
...
AIDA> exit
```

## Running with the Genkit Developer UI

Start the application under the Genkit CLI to inspect agent execution traces in the Developer UI (`http://localhost:4000`):

```bash
genkit start -- go run .
```

## Calling the HTTP Agent Endpoint

While the agent is running (defaults to port `3400`, or `PORT` if set), you can also invoke turns over HTTP:

```bash
# Start a new session (Turn 1)
curl -X POST http://localhost:3400/agents/aida \
  -H "Content-Type: application/json" \
  -d '{"data": {"message": {"role": "user", "content": [{"text": "What is my battery status?"}]}}}'

# Resume an existing session (Turn 2)
curl -X POST http://localhost:3400/agents/aida \
  -H "Content-Type: application/json" \
  -d '{"data": {"message": {"role": "user", "content": [{"text": "How many charge cycles does it have?"}]}}, "init": {"sessionId": "<SESSION_ID>"}}'
```
