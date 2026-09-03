# Retro Game Appraiser: Genkit Flow Pipeline

This example demonstrates how to build an agent workflow in Go using **[Genkit Go](https://genkit.dev)** (`github.com/firebase/genkit/go`).

The service structures the agent as a strongly typed, observable pipeline:
* **Typed Flow**: Defined using `genkit.DefineFlow` with automatic schema reflection.
* **Custom Tools**: Strongly typed Go functions registered with `ai.DefineTool`.
* **Server-Side Grounding & Thinking**: Configured with `GoogleSearch` grounding and thinking mode enabled.
* **HTTP API Server**: Mounted via `genkit.Handler` on `POST /api/appraise` with graceful shutdown handling.

## Prerequisites

1. Authenticate with Google Cloud using Application Default Credentials (ADC):
   ```bash
   gcloud auth application-default login
   ```

2. Export your Google Cloud project ID:
   ```bash
   export GOOGLE_CLOUD_PROJECT="your-project-id"
   ```

## Running the Server

Start the HTTP server directly:

```bash
go run main.go
```

By default, the server listens on port `8080` (or the port specified by the `PORT` environment variable).

## Testing the Endpoint

Send a POST request to `/api/appraise` using `curl`:

```bash
curl -X POST http://localhost:8080/api/appraise \
  -H "Content-Type: application/json" \
  -d '{"data": {"prompt": "I found a copy of Chrono Trigger for SNES for $250. Is this a good deal?"}}'
```

Example JSON response:

```json
{
  "result": {
    "appraisal": "Based on your personal collection, you already own Chrono Trigger (SNES) in Complete in Box (CIB) condition, for which you paid $210.00. Current market prices for a loose cartridge range from $180-$220, while CIB copies often command $400+. At $250, if it is CIB it represents a great price, but since you already own a CIB copy with original map and registration card, purchasing a duplicate may not be necessary."
  }
}
```

To stop the server, press `Ctrl+C`. The process will execute a graceful shutdown, draining existing connections before exiting.
