# Retro Game Appraiser: Native SDK Agent Loop

This example demonstrates how to build an autonomous agent from scratch in Go using the official **[Go GenAI SDK](https://pkg.go.dev/google.golang.org/genai)** (`google.golang.org/genai`).

The agent implements an explicit control loop (Reasoning $\rightarrow$ Tool Dispatch $\rightarrow$ Observation $\rightarrow$ Synthesis) with:
* **Custom Function Tool**: Searches a local personal inventory catalog for retro video games.
* **Server-Side Grounding**: Uses Google Search to retrieve live market values and price comparisons.
* **Interactive CLI Interface**: Multi-turn chat session with graceful shutdown on `exit`, `Ctrl+C`, or `Ctrl+D`.

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

Run the agent directly from this directory:

```bash
go run main.go
```

The interactive chat session will start in your terminal:

```text
Retro Game Appraiser (SDK Agent)
Type your question below, or 'exit' (Ctrl+C / Ctrl+D) to quit.
-----------------------------------------------------------------

User: I found a copy of EarthBound for SNES in mint Complete-in-Box (CIB) condition for $350. Do I already own it, and is $350 a good deal compared to current market prices?
[Harness] Executing tool: search_catalog(args=map[query:EarthBound])

Appraiser: Here is your collection check and appraisal for EarthBound (SNES):
...
```

To exit the session at any time:
* Type `exit` and press Enter
* Press `Ctrl+C` (SIGINT)
* Press `Ctrl+D` (EOF)
