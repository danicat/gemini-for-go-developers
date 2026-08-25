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

	"google.golang.org/genai"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model/gemini"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"
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
	// Query is the game title or platform to search in the inventory.
	Query string `json:"query"`
}

type CatalogResponse struct {
	Found   bool       `json:"found"`
	Message string     `json:"message,omitempty"`
	Count   int        `json:"count,omitempty"`
	Results []GameItem `json:"results,omitempty"`
}

type AppraiserRequest struct {
	Prompt    string `json:"prompt"`
	SessionID string `json:"session_id,omitempty"`
}

type AppraiserResponse struct {
	Appraisal string `json:"appraisal"`
	SessionID string `json:"session_id,omitempty"`
}

func main() {
	ctx := context.Background()

	// 1. Initialise Gemini Model adapter for Gemini Enterprise / Vertex AI
	model, err := gemini.NewModel(ctx, "gemini-3.7-flash", &genai.ClientConfig{
		Project:  os.Getenv("GOOGLE_CLOUD_PROJECT"),
		Location: "global",
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		log.Fatalf("failed to create Gemini model: %v", err)
	}

	// 2. Wrap collection lookup as an ADK Function Tool
	catalogTool, err := functiontool.New(functiontool.Config{
		Name:        "search_catalog",
		Description: "Search the collector's personal inventory for owned games by title or platform.",
	}, func(ctx agent.Context, req CatalogRequest) (CatalogResponse, error) {
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
	})
	if err != nil {
		log.Fatalf("failed to create catalog tool: %v", err)
	}

	// 3. Define autonomous LLM Agent
	appraiserAgent, err := llmagent.New(llmagent.Config{
		Name:        "retro_game_appraiser",
		Model:       model,
		Description: "Expert appraiser that analyzes retro video game purchases and collection inventory.",
		Instruction: "You are an expert Retro Game Appraiser. Assist collectors by verifying collection " +
			"status with search_catalog, assessing condition variants, and offering objective buying recommendations.",
		Tools: []tool.Tool{catalogTool},
	})
	if err != nil {
		log.Fatalf("failed to create appraiser agent: %v", err)
	}

	// 4. Initialise ADK Runner with Session Service for state management
	r, err := runner.New(runner.Config{
		AppName:           "retro_game_vault",
		Agent:             appraiserAgent,
		SessionService:    session.InMemoryService(),
		AutoCreateSession: true,
	})
	if err != nil {
		log.Fatalf("failed to create ADK runner: %v", err)
	}

	// 5. HTTP Handler streaming multi-turn responses
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/chat", func(w http.ResponseWriter, req *http.Request) {
		var chatReq AppraiserRequest
		if err := json.NewDecoder(req.Body).Decode(&chatReq); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		userMsg := &genai.Content{
			Role: genai.RoleUser,
			Parts: []*genai.Part{
				{Text: chatReq.Prompt},
			},
		}
		sessionID := chatReq.SessionID
		if sessionID == "" {
			sessionID = fmt.Sprintf("session-%d", time.Now().UnixNano())
		}

		var responseBuilder strings.Builder
		for event, err := range r.Run(req.Context(), "collector_user", sessionID, userMsg, agent.RunConfig{}) {
			if err != nil {
				log.Printf("runner error: %v", err)
				http.Error(w, fmt.Sprintf("agent error: %v", err), http.StatusInternalServerError)
				return
			}
			if event != nil && event.Content != nil {
				for _, part := range event.Content.Parts {
					if part.Text != "" {
						responseBuilder.WriteString(part.Text)
					}
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(AppraiserResponse{
			Appraisal: responseBuilder.String(),
			SessionID: sessionID,
		})
	})

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
		log.Printf("Retro Game Appraiser (ADK) listening on :%s", port)
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
