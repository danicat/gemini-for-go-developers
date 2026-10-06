package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/firebase/genkit/go/plugins/middleware"
	"github.com/firebase/genkit/go/plugins/server"
)

//go:embed prompts/*
var promptsFS embed.FS

type AIDARequest struct {
	OS           string `json:"os,omitempty" jsonschema_description:"The target operating system (defaults to runtime.GOOS)"`
	UserQuestion string `json:"userQuestion" jsonschema_description:"The diagnostic question to investigate"`
}

type AIDAResponse struct {
	Message string `json:"message" jsonschema_description:"The diagnostic findings and recommendations"`
}

type OsqueryInput struct {
	Query string `json:"query" jsonschema_description:"The SQL query to execute in osquery."`
}

func main() {
	ctx := context.Background()

	g := genkit.Init(ctx,
		genkit.WithPlugins(
			&googlegenai.VertexAI{},
			&middleware.Middleware{},
		),
		genkit.WithPromptFS(promptsFS),
	)

	genkit.DefineSchemaFor[AIDARequest](g)
	genkit.DefineSchemaFor[AIDAResponse](g)

	genkit.DefineTool(g, "runOsquery", "Runs a system inspection query using osquery.",
		func(ctx *ai.ToolContext, in OsqueryInput) (string, error) {
			out, err := exec.CommandContext(ctx, "osqueryi", "--json", in.Query).CombinedOutput()
			if err != nil {
				return "", fmt.Errorf("osqueryi failed: %w: %s", err, out)
			}
			return string(out), nil
		},
	)

	aida := genkit.LookupDataPrompt[AIDARequest, *AIDAResponse](g, "aida")

	genkit.DefineFlow(
		g,
		"diagnose",
		func(ctx context.Context, req AIDARequest) (AIDAResponse, error) {
			if req.OS == "" {
				req.OS = runtime.GOOS
			}
			resp, _, err := aida.Execute(ctx, req)
			if err != nil {
				return AIDAResponse{}, fmt.Errorf("diagnose flow failed: %w", err)
			}
			if resp == nil {
				return AIDAResponse{}, fmt.Errorf("diagnose flow returned empty response")
			}
			return *resp, nil
		},
	)

	mux := http.NewServeMux()
	for _, flow := range genkit.ListFlows(g) {
		mux.HandleFunc("POST /"+flow.Name(), genkit.Handler(flow))
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3400"
	}

	log.Fatal(server.Start(ctx, "0.0.0.0:"+port, mux))
}
