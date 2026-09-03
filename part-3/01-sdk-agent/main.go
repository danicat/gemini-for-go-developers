package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"google.golang.org/genai"
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

// searchCatalogTool searches the local collection for matching games.
func searchCatalogTool(args map[string]any) map[string]any {
	query, _ := args["query"].(string)
	queryLower := strings.ToLower(strings.TrimSpace(query))

	var matches []GameItem
	for _, item := range localCatalog {
		if strings.Contains(strings.ToLower(item.Title), queryLower) ||
			strings.Contains(strings.ToLower(item.Platform), queryLower) {
			matches = append(matches, item)
		}
	}

	if len(matches) == 0 {
		return map[string]any{
			"found":   false,
			"message": fmt.Sprintf("No items matching %q found in your collection.", query),
		}
	}

	return map[string]any{
		"found":   true,
		"count":   len(matches),
		"results": matches,
	}
}

func main() {
	ctx := context.Background()

	// Initialise GenAI client for Gemini Enterprise
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  os.Getenv("GOOGLE_CLOUD_PROJECT"),
		Location: "global",
		Backend:  genai.BackendEnterprise,
	})
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	// 1. Declare custom function schema for collection lookup
	catalogToolDecl := &genai.FunctionDeclaration{
		Name:        "search_catalog",
		Description: "Search the collector's personal inventory for owned games by title or platform.",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"query": {
					Type:        genai.TypeString,
					Description: "Game title or platform to search (e.g. 'EarthBound', 'SNES').",
				},
			},
			Required: []string{"query"},
		},
	}

	// 2. Configure model tools: custom function declaration + Google Search grounding
	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{Text: "You are an expert Retro Game Appraiser. When evaluating purchases, check the user's " +
					"collection catalog first to see if they already own the item, then check current market " +
					"prices using Google Search to evaluate whether the deal is fair, overpriced, or a bargain."},
			},
		},
		Tools: []*genai.Tool{
			{
				FunctionDeclarations: []*genai.FunctionDeclaration{catalogToolDecl},
			},
			{
				GoogleSearch: &genai.GoogleSearch{},
			},
		},
	}

	// 3. Prepare initial user prompt
	prompt := "I found a copy of EarthBound for SNES in mint Complete-in-Box (CIB) condition for $350. " +
		"Do I already own it, and is $350 a good deal compared to current market prices?"

	contents := []*genai.Content{
		{
			Role:  "user",
			Parts: []*genai.Part{genai.NewPartFromText(prompt)},
		},
	}

	model := "gemini-3.7-flash"
	maxTurns := 6

	// 4. The Agent Loop (LLM Request -> Tool Dispatch -> Tool Result -> LLM Request)
	for turn := 0; turn < maxTurns; turn++ {
		resp, err := client.Models.GenerateContent(ctx, model, contents, config)
		if err != nil {
			log.Fatalf("error generating content: %v", err)
		}

		if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
			log.Fatal("received empty response candidate from model")
		}

		// Append the model's response to the conversation history
		modelContent := resp.Candidates[0].Content
		contents = append(contents, modelContent)

		// Check if the model requested any client-side tool executions
		funcCalls := resp.FunctionCalls()
		if len(funcCalls) == 0 {
			// Terminal condition: model produced its final natural language answer
			fmt.Println("\n=== Appraiser Verdict ===")
			fmt.Println(resp.Text())
			return
		}

		// Execute each requested tool and prepare response parts
		var responseParts []*genai.Part
		for _, call := range funcCalls {
			fmt.Printf("[Harness] Executing tool: %s(args=%v)\n", call.Name, call.Args)

			var result map[string]any
			switch call.Name {
			case "search_catalog":
				result = searchCatalogTool(call.Args)
			default:
				result = map[string]any{"error": fmt.Sprintf("unsupported tool: %s", call.Name)}
			}

			responseParts = append(responseParts, genai.NewPartFromFunctionResponse(call.Name, result))
		}

		// Return tool execution results as a user turn
		contents = append(contents, &genai.Content{
			Role:  "user",
			Parts: responseParts,
		})
	}

	log.Fatal("exceeded maximum conversation turns without reaching terminal state")
}
