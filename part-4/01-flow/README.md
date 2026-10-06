# AIDA Diagnostic Flow (`part-4/01-flow`)

This example demonstrates how to build **AIDA (AI Diagnostic Agent)** as a strongly typed Genkit Go flow with `dotprompt`, built-in middleware (`retry` and `skills`), and HTTP deployment via `genkit.Handler`.

Key features demonstrated:
* **Typed Flow & Data Prompts**: Strongly typed input (`AIDARequest`) and output (`AIDAResponse`) schemas registered with `genkit.DefineSchemaFor` and bound to `prompts/aida.prompt` via `genkit.LookupDataPrompt`.
* **Embedded Dotprompt**: Prompt templates (`prompts/*`) embedded into the Go binary via `//go:embed` and `genkit.WithPromptFS`.
* **Custom `osquery` Tool**: Registers `runOsquery` with `genkit.DefineTool` to execute read-only system queries via `osqueryi --json`.
* **Middleware & Agent Skills**: Uses `genkit-middleware/retry` and `genkit-middleware/skills` to load operating system specific `osquery` schemas (`skills/darwin`, `skills/linux`, `skills/windows`) on demand.
* **HTTP Endpoint**: Exposes registered flows over HTTP using `genkit.ListFlows`, `genkit.Handler`, and `server.Start`.

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

## Running with the Genkit Developer UI

Start the application under the Genkit CLI to launch the Developer UI (`http://localhost:4000`):

```bash
genkit start -- go run .
```

Select the **diagnose** flow in the Developer UI and run it with:

```json
{
  "userQuestion": "My machine is running slow. Which top 5 processes are consuming the most memory right now?"
}
```

Or run the flow once from the terminal:

```bash
genkit flow:run diagnose '{"userQuestion": "My machine is running slow. Which top 5 processes are consuming the most memory right now?"}' --non-interactive -- go run .
```

## Running as an HTTP Server

Start the HTTP server directly (defaults to port `3400`, or `PORT` if set):

```bash
go run .
```

Send a `POST` request to `/diagnose`:

```bash
curl -X POST http://localhost:3400/diagnose \
  -H "Content-Type: application/json" \
  -d '{"data": {"userQuestion": "My machine is running slow. Which top 5 processes are consuming the most memory right now?"}}'
```
