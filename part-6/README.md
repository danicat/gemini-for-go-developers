# Chibi Dani Public Agent (`part-6`)

This directory holds the complete source code for **Chibi Dani**, the public blog companion agent built for [Part 6 of Gemini for Go Developers](https://danicat.dev/posts/chibi-dani-architecture/).

It combines Genkit Go token budgets, HTTP perimeter checks, Google Cloud Model Armor screening, PII-redacted Firestore session telemetry, and direct binary deployments to Cloud Run.

## Architecture and Files

- [`main.go`](main.go): Initializes Gemini Enterprise (`gemini-3.5-flash-lite` with `MaxOutputTokens: 1500`), dynamic system prompts (`ai.WithSystemFn` backed by `corpusStore.Current(ctx)`), `read_corpus`, turn caps (`ai.WithMaxTurns(5)`), and `mwexp.Budget` token limits per run and per session.
- [`system_prompt.md`](system_prompt.md): Defines Chibi Dani's persona, topic scope, citation format, refusal rules, and recency/supersession rules embedded ahead of the content index.
- [`corpus.go`](corpus.go) and [`corpus_test.go`](corpus_test.go): Preload `llms.txt` and localized `llms-full.txt` bundles (`en`, `pt-br`, `ja`) from `CORPUS_BASE_URL` (default `https://danicat.dev`) into an immutable in-memory snapshot with same-origin redirect protection and conditional `If-None-Match` (`24h` default via `CORPUS_REFRESH_INTERVAL`) ETag refreshes on incoming turns.
- [`session_store.go`](session_store.go): Wraps `localstore.InMemorySessionStore` in a bounded LRU (`256` sessions) and `30`-minute idle TTL store validated against Genkit's `sessionstoretest` suite so unauthenticated traffic cannot exhaust container RAM.
- [`server.go`](server.go): Chains `http.CrossOriginProtection`, strict `Origin` + CORS enforcement (with `CHIBI_ALLOW_LOCALHOST=true` opt-in for local dev), an `O(1)` LRU-bounded per-IP rate limiter (`6 req/min`, burst `4`, max `4,096` IPs), a global container rate limiter (`20 req/min`, burst `10`) plus rolling `24h` request cap (`1,500`/day), trailing `X-Forwarded-For` parsing (`clientIP`), an `8 KB` / `1,200`-rune strict JSON envelope check, and server-minted UUIDv4 `sessionId` verification.
- [`armor.go`](armor.go): Wraps generation in `WrapGenerate` with regional Google Cloud Model Armor screening: `SanitizeUserPrompt` once per turn (`Iteration == 0`), unary `SanitizeModelResponse` for non-streaming calls, and bidirectional gRPC `StreamSanitizeModelResponse` (`STREAMING_MODE_REALTIME`) for SSE streams so unchecked chunks never reach the client.
- [`telemetry.go`](telemetry.go): Truncates and redacts PII from user questions and writes per-turn outcomes (`ok`, `armor_blocked`, `budget_exceeded`, `error`), cited URLs, and latency to a Firestore `turns` subcollection while turning off Genkit prompt/reply logging in Cloud Trace.

## Prerequisites

1. Install Go `1.26+` and Google Cloud CLI (`gcloud`).
2. Authenticate with Google Cloud:
   ```bash
   gcloud auth application-default login
   ```
3. Configure a Google Cloud project with Gemini Enterprise (`aiplatform.googleapis.com`), Cloud Firestore, and a regional [Model Armor template](https://docs.cloud.google.com/model-armor/manage-templates).
4. Export the required environment variables:
   ```bash
   export GOOGLE_CLOUD_PROJECT="your-project-id"
   export MODEL_ARMOR_TEMPLATE="projects/your-project-id/locations/us-central1/templates/chibi-dani-template"
   ```

## Running Locally and Testing

Run the unit test suite:

```bash
go test -count=1 ./...
```

Start the HTTP server locally with localhost origins enabled (by default it preloads from `https://danicat.dev`; set `CORPUS_BASE_URL` to point at a local Hugo server or custom mirror, and optionally `CORPUS_REFRESH_INTERVAL` such as `24h`):

```bash
CHIBI_ALLOW_LOCALHOST=true go run .
```

Send a test request from an allowed local origin:

```bash
curl -i -X POST http://localhost:8080/agents/chibi-dani \
  -H "Content-Type: application/json" \
  -H "Origin: http://localhost:1313" \
  -d '{"data":{"message":{"role":"user","content":[{"text":"Which post covers Genkit Go?"}]}}}'
```

## Deploying to Cloud Run

See [Part 6 of Gemini for Go Developers](https://danicat.dev/posts/chibi-dani-architecture/) for the least-privilege IAM roles, Model Armor template setup, and Cloud Run flags used to run this pattern in production.
