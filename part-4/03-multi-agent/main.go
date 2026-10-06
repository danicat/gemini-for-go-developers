package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/ai/exp/localstore"
	"github.com/firebase/genkit/go/genkit"
	genkitx "github.com/firebase/genkit/go/genkit/exp"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/firebase/genkit/go/plugins/middleware"
	middlewarex "github.com/firebase/genkit/go/plugins/middleware/exp"
	"github.com/firebase/genkit/go/plugins/server"
	"google.golang.org/genai"
)

type OsqueryInput struct {
	Query string `json:"query" jsonschema_description:"The SQL query to execute in osquery."`
}

func main() {
	ctx := context.Background()

	g := genkit.Init(ctx,
		genkit.WithExperimental(),
		genkit.WithPlugins(&googlegenai.VertexAI{}, &middleware.Middleware{}, &middlewarex.Middleware{}),
	)

	runOsquery := genkit.DefineTool(g, "runOsquery", "Runs a system inspection query using osquery.",
		func(ctx *ai.ToolContext, in OsqueryInput) (string, error) {
			out, err := exec.CommandContext(ctx, "osqueryi", "--json", in.Query).CombinedOutput()
			if err != nil {
				return "", fmt.Errorf("osqueryi failed: %w: %s", err, out)
			}
			return string(out), nil
		},
	)

	inspector := genkitx.DefineAgent(g, "inspector",
		aix.InlinePrompt{
			ai.WithModelName("vertexai/gemini-3.8-flash"),
			ai.WithSystem(fmt.Sprintf("Inspect %s system telemetry using osquery and report diagnostic findings.", runtime.GOOS)),
			ai.WithConfig(&genai.GenerateContentConfig{
				ThinkingConfig: &genai.ThinkingConfig{
					ThinkingLevel: genai.ThinkingLevelLow,
				},
			}),
			ai.WithTools(runOsquery),
			ai.WithUse(
				&middleware.Retry{MaxRetries: 5},
				&middleware.Skills{SkillPaths: []string{"./skills"}},
			),
		},
		aix.WithDescription[struct{}]("Inspects host metrics, running processes, and system configuration using osquery."),
	)

	researcher := genkitx.DefineAgent(g, "researcher",
		aix.InlinePrompt{
			ai.WithModelName("vertexai/gemini-3.8-flash"),
			ai.WithSystem("Research unfamiliar processes, binaries, error messages, and CVEs using Google Search and report concise findings."),
			ai.WithConfig(&genai.GenerateContentConfig{
				ThinkingConfig: &genai.ThinkingConfig{
					ThinkingLevel: genai.ThinkingLevelLow,
				},
				Tools: []*genai.Tool{
					{GoogleSearch: &genai.GoogleSearch{}},
				},
			}),
		},
		aix.WithDescription[struct{}]("Searches the web for information on unfamiliar processes, binaries, error codes, and CVEs."),
	)

	store := localstore.NewInMemorySessionStore[struct{}]()

	aida := genkitx.DefineAgent(g, "aida",
		aix.InlinePrompt{
			ai.WithModelName("vertexai/gemini-3.8-flash"),
			ai.WithSystem(fmt.Sprintf("You are AIDA, a %s operating system diagnostic assistant. Delegate host inspection to inspector and web research to researcher, then synthesize findings for the user.", runtime.GOOS)),
			ai.WithConfig(&genai.GenerateContentConfig{
				ThinkingConfig: &genai.ThinkingConfig{
					ThinkingLevel: genai.ThinkingLevelLow,
				},
			}),
			ai.WithUse(
				&middleware.Retry{MaxRetries: 5},
				&middlewarex.Agents{
					Agents:         []aix.AgentRef{inspector.Ref(), researcher.Ref()},
					MaxDelegations: 5,
				},
			),
		},
		aix.WithSessionStore(store),
	)

	mux := http.NewServeMux()
	for _, route := range genkitx.AllAgentRoutes(g) {
		mux.HandleFunc(route.Pattern(), route.Handler())
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3400"
	}

	srvErr := make(chan error, 1)
	go func() {
		srvErr <- server.Start(ctx, "0.0.0.0:"+port, mux)
	}()

	// Interactive terminal session maintaining state across turns
	scanner := bufio.NewScanner(os.Stdin)
	var sessionID string
	var readAny bool

	fmt.Print("AIDA> ")
	for scanner.Scan() {
		readAny = true
		input := strings.TrimSpace(scanner.Text())
		if input == "" || input == "exit" {
			break
		}

		var opts []aix.InvocationOption[struct{}]
		if sessionID != "" {
			opts = append(opts, aix.WithSessionID[struct{}](sessionID))
		}

		out, err := aida.RunText(ctx, input, opts...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			fmt.Print("AIDA> ")
			continue
		}
		if out.Error != nil {
			fmt.Fprintf(os.Stderr, "agent turn error (%s): %v\n", out.FinishReason, out.Error)
			fmt.Print("AIDA> ")
			continue
		}

		sessionID = out.SessionID
		fmt.Printf("\n%s\n\nAIDA> ", out.Message.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("reading standard input: %v", err)
	}

	if !readAny || os.Getenv("GENKIT_ENV") == "dev" {
		if err := <-srvErr; err != nil {
			log.Fatal(err)
		}
	}
}
