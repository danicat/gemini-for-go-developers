package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
)

// GameItem represents a collectible item in the user's personal inventory.
type GameItem struct {
	Title     string  `json:"title"`
	Platform  string  `json:"platform"`
	Year      int     `json:"year"`
	Condition string  `json:"condition"` // e.g. "Loose Cartridge", "CIB (Complete in Box)", "Mint"
	PricePaid float64 `json:"price_paid"`
	Notes     string  `json:"notes"`
}

// localCatalog simulates an inventory database for retro games.
var localCatalog = []GameItem{
	{
		Title:     "Chrono Trigger",
		Platform:  "Super Nintendo (SNES)",
		Year:      1995,
		Condition: "CIB (Complete in Box)",
		PricePaid: 210.00,
		Notes:     "Includes original map and registration card.",
	},
	{
		Title:     "EarthBound",
		Platform:  "Super Nintendo (SNES)",
		Year:      1994,
		Condition: "Loose Cartridge",
		PricePaid: 180.00,
		Notes:     "Authentic board verified; label in excellent shape.",
	},
	{
		Title:     "Castlevania: Symphony of the Night",
		Platform:  "Sony PlayStation",
		Year:      1997,
		Condition: "CIB (Black Label)",
		PricePaid: 135.00,
		Notes:     "Original soundtrack disc included.",
	},
}

type CatalogRequest struct {
	Query string `json:"query" jsonschema:"description=The game title or platform to search in the inventory"`
}

type CatalogResponse struct {
	Found   bool       `json:"found"`
	Message string     `json:"message,omitempty"`
	Count   int        `json:"count,omitempty"`
	Results []GameItem `json:"results,omitempty"`
}

type AppraiserRequest struct {
	Prompt string `json:"prompt"`
}

type AppraiserResponse struct {
	Appraisal string `json:"appraisal"`
}

func main() {
	ctx := context.Background()

	// 1. Initialise Genkit with Vertex AI / Gemini Enterprise plugin
	g := genkit.Init(ctx,
		genkit.WithPlugins(&googlegenai.VertexAI{
			ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
			Location:  "global",
		}),
		genkit.WithDefaultModel("vertexai/gemini-3.7-flash"),
	)

	// 2. Define strongly-typed tool with automatic schema generation
	catalogTool := genkit.DefineTool(
		g,
		"search_catalog",
		"Search the collector's personal inventory for owned games by title or platform.",
		func(ctx *ai.ToolContext, req CatalogRequest) (CatalogResponse, error) {
			queryLower := strings.ToLower(strings.TrimSpace(req.Query))
			var matches []GameItem

			for _, item := range localCatalog {
				if strings.Contains(strings.ToLower(item.Title), queryLower) ||
					strings.Contains(strings.ToLower(item.Platform), queryLower) {
					matches = append(matches, item)
				}
			}

			if len(matches) == 0 {
				return CatalogResponse{
					Found:   false,
					Message: fmt.Sprintf("No items matching %q found in personal collection.", req.Query),
				}, nil
			}

			return CatalogResponse{
				Found:   true,
				Count:   len(matches),
				Results: matches,
			}, nil
		},
	)

	// 3. Define the Appraiser Flow
	appraiserFlow := genkit.DefineFlow(
		g,
		"retroAppraiserFlow",
		func(ctx context.Context, input string) (string, error) {
			systemInstruction := "You are an expert Retro Game Appraiser. When evaluating purchases, " +
				"check the user's collection catalog first, then evaluate current market value " +
				"to provide an accurate appraisal and buying recommendation."

			resp, err := genkit.GenerateText(ctx, g,
				ai.WithSystem(systemInstruction),
				ai.WithPrompt(input),
				ai.WithTools(catalogTool),
			)
			if err != nil {
				return "", fmt.Errorf("appraiser generation error: %w", err)
			}

			return resp, nil
		},
	)

	// 4. Expose the Flow as a standard HTTP API
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/appraise", func(w http.ResponseWriter, r *http.Request) {
		var req AppraiserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		verdict, err := appraiserFlow.Run(r.Context(), req.Prompt)
		if err != nil {
			http.Error(w, fmt.Sprintf("evaluation error: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(AppraiserResponse{Appraisal: verdict})
	})

	// Also attach Genkit's standard flow handler for Developer UI inspection
	mux.Handle("/flow/retroAppraiserFlow", genkit.Handler(appraiserFlow))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Graceful shutdown on Ctrl+C (SIGINT) or SIGTERM
	serverCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Retro Game Appraiser (Genkit) listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	<-serverCtx.Done()
	log.Println("\nShutting down server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced shutdown: %v", err)
	}
	log.Println("Server exited cleanly.")
}
