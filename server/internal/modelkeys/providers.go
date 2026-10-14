// Package modelkeys is a signed-in person's own model API key for the AI
// seats they add (ADR 0182, migration 0025): the providers it may point
// at, the checks on what a person types, the sealed store, and the model
// client a seat on that key calls through.
//
// # What the package promises
//
//   - The key is sealed at rest (AES-256-GCM under CMDCTRL_IDENTITY_KEY,
//     ADR 0051 decision 5) and opened only to make a call. With no
//     sealing key the server builds no Service and the feature is off:
//     there is no plaintext fallback.
//   - The key never leaves the server again: View, the one shape the
//     HTTP layer gets, holds the provider, the model, the last four
//     characters and dates, and no field that could carry the key.
//   - The key is never in a log line or an error string. A provider's
//     own error text can echo part of a key ("Incorrect API key
//     provided: sk-...abcd"), so nothing a provider says is passed on:
//     every failure becomes one of this package's fixed sentences.
//   - Requests go through netguard: https only, a public address checked
//     at dial time, no redirects, a short timeout and a response cap.
//     That holds for the listed providers too, not only for "custom".
//   - Only the owner's own seats use it: a seat names its key owner,
//     and the lobby sets that to the person who added the seat.
package modelkeys

import "strings"

// Transport is which wire format a provider speaks.
type Transport string

const (
	// TransportAnthropic is the Messages API (aiseat/model.AnthropicClient).
	TransportAnthropic Transport = "anthropic"
	// TransportOpenAI is the OpenAI chat-completions API
	// (aiseat/model.OpenAIClient), which most hosted providers serve.
	TransportOpenAI Transport = "openai"
)

// CustomProvider is the id of an OpenAI-compatible endpoint the person
// types in. It is held to every check the listed ones are, plus the URL
// checks in ValidateCustomURL.
const CustomProvider = "custom"

// Provider is one choice in the Settings picker.
type Provider struct {
	ID        string    `json:"id"`
	Label     string    `json:"label"`
	Transport Transport `json:"transport"`
	// BaseURL is where the requests go. Fixed in code for every listed
	// provider; empty for "custom", whose address the person supplies.
	BaseURL string `json:"base_url,omitempty"`
	// ModelHint is an example model id for the placeholder. A hint, not
	// a default anyone is held to: providers rename models often.
	ModelHint string `json:"model_hint,omitempty"`
}

// providers is the allowlist. Every base URL is https on port 443.
var providers = []Provider{
	{ID: "anthropic", Label: "Anthropic", Transport: TransportAnthropic,
		BaseURL: "https://api.anthropic.com/v1/messages", ModelHint: "claude-haiku-4-5"},
	{ID: "openai", Label: "OpenAI", Transport: TransportOpenAI,
		BaseURL: "https://api.openai.com/v1", ModelHint: "gpt-4o-mini"},
	{ID: "openrouter", Label: "OpenRouter", Transport: TransportOpenAI,
		BaseURL: "https://openrouter.ai/api/v1", ModelHint: "anthropic/claude-haiku-4.5"},
	{ID: "groq", Label: "Groq", Transport: TransportOpenAI,
		BaseURL: "https://api.groq.com/openai/v1", ModelHint: "llama-3.3-70b-versatile"},
	{ID: "together", Label: "Together AI", Transport: TransportOpenAI,
		BaseURL: "https://api.together.xyz/v1", ModelHint: "meta-llama/Llama-3.3-70B-Instruct-Turbo"},
	{ID: "mistral", Label: "Mistral", Transport: TransportOpenAI,
		BaseURL: "https://api.mistral.ai/v1", ModelHint: "mistral-small-latest"},
	{ID: "deepseek", Label: "DeepSeek", Transport: TransportOpenAI,
		BaseURL: "https://api.deepseek.com/v1", ModelHint: "deepseek-chat"},
	{ID: CustomProvider, Label: "Other (OpenAI-compatible)", Transport: TransportOpenAI},
}

// Providers lists the picker's choices, in order. The slice is a copy.
func Providers() []Provider { return append([]Provider(nil), providers...) }

// LookupProvider finds a provider by id, ignoring case and spaces.
func LookupProvider(id string) (Provider, bool) {
	id = strings.ToLower(strings.TrimSpace(id))
	for _, p := range providers {
		if p.ID == id {
			return p, true
		}
	}
	return Provider{}, false
}
