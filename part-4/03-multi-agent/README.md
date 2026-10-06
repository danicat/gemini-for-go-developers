# AIDA Multi-Agent Delegation (`part-4/03-multi-agent`)

This example extends the stateful **AIDA (AI Diagnostic Agent)** with multi-agent delegation using Genkit Go's experimental `&middlewarex.Agents{}` middleware (`github.com/firebase/genkit/go/plugins/middleware/exp`).

Key features demonstrated:
* **Specialist Sub-Agents (`genkitx.DefineAgent`)**:
  * **`inspector`**: Queries local host telemetry using `runOsquery` and OS-specific `osquery` skills (`&middleware.Skills{SkillPaths: []string{"./skills"}}`).
  * **`researcher`**: Uses Google Search grounding (`&genai.GoogleSearch{}`) to look up unfamiliar processes, binaries, error messages, and CVEs without bloating AIDA's main conversation history.
* **Coordinator Agent (`aida`)**: Uses `&middlewarex.Agents{Agents: []aix.AgentRef{inspector.Ref(), researcher.Ref()}, MaxDelegations: 5}` to expose `delegate_to_inspector` and `delegate_to_researcher` tools, delegating host inspection and web research and synthesizing the results across turns.

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

Start the interactive multi-agent diagnostic REPL:

```bash
go run .
```

Example session:

```text
AIDA> Check what OS version this host is running and what process is using the most resident memory.
...
AIDA> Search the web to explain what that top process does on macOS.
...
AIDA> exit
```

## Running with the Genkit Developer UI

Start the application under the Genkit CLI to inspect multi-agent delegation traces in the Developer UI (`http://localhost:4000`):

```bash
genkit start -- go run .
```
