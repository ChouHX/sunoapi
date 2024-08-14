package models

type GeneralOpenAIRequest struct {
	Model            string      `json:"model,omitempty"`
	Messages         []Message   `json:"messages,omitempty"`
	Stream           bool        `json:"stream,omitempty"`
	MaxTokens        uint        `json:"max_tokens,omitempty"`
	Temperature      float64     `json:"temperature,omitempty"`
	TopP             float64     `json:"top_p,omitempty"`
	TopK             int         `json:"top_k,omitempty"`
	FunctionCall     interface{} `json:"function_call,omitempty"`
	FrequencyPenalty float64     `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64     `json:"presence_penalty,omitempty"`
	ToolChoice       string      `json:"tool_choice,omitempty"`
	Tools            []Tool      `json:"tools,omitempty"`
}
type Message struct {
	Role    string  `json:"role"`
	Content string  `json:"content"`
	Name    *string `json:"name,omitempty"`
}

type Tool struct {
	Id       string   `json:"id,omitempty"`
	Type     string   `json:"type"`
	Function Function `json:"function"`
}
type Function struct {
	Url         string    `json:"url,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Parameters  Parameter `json:"parameters"`
	Arguments   string    `json:"arguments,omitempty"`
}
type Parameter struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}
