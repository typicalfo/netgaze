# Addendum 11 – ADK Agent Setup & Tool Signatures

**Responsibility**: Configure Google ADK with OpenRouter integration and define AI tools.

**Files**: `internal/agent/agent.go`, `internal/agent/tools.go`

**Behavior**
- Only initialized when NOT in --no-agent mode
- Uses OpenRouter API with grok-4.1 model (free tier)
- Environment variable: `OPENROUTER_API_KEY`
- Single agent with 4 custom tools for network intelligence

**Agent initialization**
```go
// agent.go
package agent

import (
    "context"
    "os"
    "github.com/google/agent-toolkit-go/pkg/agent"
    "github.com/google/agent-toolkit-go/pkg/llm"
)

type Agent struct {
    client *agent.Client
}

func New() (*Agent, error) {
    apiKey := os.Getenv("OPENROUTER_API_KEY")
    if apiKey == "" {
        return nil, fmt.Errorf("OPENROUTER_API_KEY required for AI mode")
    }
    
    llmClient, err := llm.NewOpenRouterClient(llm.OpenRouterConfig{
        APIKey: apiKey,
        Model:  "x-ai/grok-4.1",
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create OpenRouter client: %w", err)
    }
    
    client := agent.New(llmClient, agent.Config{
        Tools: []agent.Tool{
            SummarizeFindings{},
            DetectAnomalies{},
            SuggestNextSteps{},
            AnswerQuestion{},
        },
    })
    
    return &Agent{client: client}, nil
}
```

**Tool 1: Summarize Findings**
```go
// tools.go
type SummarizeFindings struct{}

func (t SummarizeFindings) Name() string { return "summarize_findings" }
func (t SummarizeFindings) Description() string { 
    return "Generate a concise summary of network reconnaissance findings" 
}

type SummarizeFindingsInput struct {
    JsonReport string `json:"json_report" description:"Complete netgaze report as JSON"`
}

func (t SummarizeFindings) Run(ctx context.Context, input SummarizeFindingsInput) (string, error) {
    // AI processes the JSON report and generates human-readable summary
    return t.client.Process(ctx, fmt.Sprintf(`
        Analyze this network reconnaissance data and provide a concise summary:
        %s
        
        Focus on:
        - Key network characteristics
        - Security-relevant findings
        - Notable infrastructure details
        Keep it under 200 words.
    `, input.JsonReport))
}
```

**Tool 2: Detect Anomalies**
```go
type DetectAnomalies struct{}

func (t DetectAnomalies) Name() string { return "detect_anomalies" }
func (t DetectAnomalies) Description() string { 
    return "Identify unusual or suspicious patterns in network data" 
}

type DetectAnomaliesInput struct {
    JsonReport string `json:"json_report" description:"Complete netgaze report as JSON"`
}

func (t DetectAnomalies) Run(ctx context.Context, input DetectAnomaliesInput) (string, error) {
    return t.client.Process(ctx, fmt.Sprintf(`
        Analyze this network reconnaissance data for anomalies:
        %s
        
        Look for:
        - Unusual port configurations
        - Mismatched geolocation vs ASN
        - Suspicious TLS certificates
        - Atypical network infrastructure
        Flag anything that seems abnormal or security-relevant.
    `, input.JsonReport))
}
```

**Tool 3: Suggest Next Steps**
```go
type SuggestNextSteps struct{}

func (t SuggestNextSteps) Name() string { return "suggest_next_steps" }
func (t SuggestNextSteps) Description() string { 
    return "Recommend follow-up investigation steps" 
}

type SuggestNextStepsInput struct {
    JsonReport string `json:"json_report" description:"Complete netgaze report as JSON"`
}

func (t SuggestNextSteps) Run(ctx context.Context, input SuggestNextStepsInput) (string, error) {
    return t.client.Process(ctx, fmt.Sprintf(`
        Based on this network reconnaissance data:
        %s
        
        Suggest 3-5 specific follow-up investigation steps:
        - Additional tools to run
        - Specific ports/services to investigate
        - Further research directions
        Be specific and actionable.
    `, input.JsonReport))
}
```

**Tool 4: Answer Question**
```go
type AnswerQuestion struct{}

func (t AnswerQuestion) Name() string { return "answer_question" }
func (t AnswerQuestion) Description() string { 
    return "Answer user questions about the network reconnaissance data" 
}

type AnswerQuestionInput struct {
    Question   string `json:"question" description:"User's specific question"`
    JsonReport string `json:"json_report" description:"Complete netgaze report as JSON"`
}

func (t AnswerQuestion) Run(ctx context.Context, input AnswerQuestionInput) (string, error) {
    return t.client.Process(ctx, fmt.Sprintf(`
        Answer this question about the network reconnaissance data:
        Question: %s
        
        Data: %s
        
        Provide a specific, helpful answer based only on the provided data.
        If the data doesn't support an answer, say so clearly.
    `, input.Question, input.JsonReport))
}
```

**Integration with TUI**
- Streaming responses displayed in real-time
- Tool selection automatic based on user action
- Error handling for API failures and rate limits
- Fallback to template mode if AI unavailable

**Error handling**
- Missing API key → graceful fallback to --no-agent mode
- Rate limiting → user-friendly message with retry suggestion
- Invalid JSON response → log error, continue with other tools
- Network timeouts → retry once, then fail gracefully

**Performance considerations**
- Lazy initialization (only when needed)
- Connection reuse for multiple tool calls
- Response streaming for better UX
- Timeout per tool call: 30 seconds