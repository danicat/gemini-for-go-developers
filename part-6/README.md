# Chibi Dani Public Agent (`part-6`)

This directory holds the complete source code for **Chibi Dani**, the public blog companion agent built for [Part 6 of Gemini for Go Developers](https://danicat.dev/posts/chibi-dani-architecture/).

It combines Genkit Go token budgets, HTTP perimeter checks, Google Cloud Model Armor screening, PII-redacted Firestore session telemetry, and direct binary deployments to Cloud Run.

## Architecture and Files

- [`main.go`](main.go): Initializes Gemini Enterprise (`gemini-3.5-flash-lite`), `read_corpus`, turn caps (`ai.WithMaxTurns(5)`), and `mwexp.Budget` token limits per run and per session.
- [`system_prompt.md`](system_prompt.md): Defines Chibi Dani's persona, topic scope, citation format, and refusal rules embedded ahead of the content index.
- [`corpus.go`](corpus.go): Loads published Markdown articles, codelabs, talks, and videos into memory at startup (`LoadCorpus`). A small sample dataset is bundled under `content/` and `data/` for local testing.
- [`server.go`](server.go): Chains `http.CrossOriginProtection`, CORS response headers, per-IP token-bucket rate limiting (`golang.org/x/time/rate`), trailing `X-Forwarded-For` parsing (`clientIP`), an `8 KB` / `1,200`-rune strict JSON envelope check, and server-minted UUIDv4 `sessionId` verification against `sessionStore`.
- [`armor.go`](armor.go): Wraps model calls in `WrapModel` with regional Google Cloud Model Armor screening (`SanitizeUserPrompt` and `SanitizeModelResponse`), surfacing `PERMISSION_DENIED` on policy hits.
- [`telemetry.go`](telemetry.go): Saves PII-redacted turn outcomes (`ok`, `armor_blocked`, `budget_exceeded`, `error`), cited URLs, and latency to Cloud Firestore while turning off Genkit prompt/reply logging in Cloud Trace.

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

Start the HTTP server locally using the bundled sample corpus (or point `CORPUS_DIR` to a full Hugo checkout):

```bash
go run .
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
