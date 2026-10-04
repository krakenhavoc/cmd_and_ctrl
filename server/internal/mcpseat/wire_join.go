package mcpseat

import (
	"errors"
	"net/http"
	"strings"
)

// The join body's `agent` field, ADR 0122 §7 (Delivery PR 2).
//
// Written here to the ADR because PR 2 has not merged into this branch.
// Both join routes (POST /games/{id}/join and POST /join) accept it:
//
//	{"invite_token": "…", "name": "Claude", "agent": {"client": "claude-code"}}
//
// The binary always sends it, and has no flag to leave it out. The server
// records it on the seat for good, and shows the table an "AI agent" chip.
//
// A server from before PR 2 decodes join bodies strictly and answers 400
// `unknown field "agent"`. The seat then refuses to join rather than
// retrying without the field (errNoBadge): a seat that cannot declare
// itself does not sit down.

// joinBody is the request body of both join routes.
type joinBody struct {
	InviteToken string     `json:"invite_token"`
	Name        string     `json:"name"`
	Agent       agentField `json:"agent"`
}

// agentField declares the seat an agent, and which MCP client drives it.
type agentField struct {
	Client string `json:"client"`
}

// errNoBadge is a server that cannot record the badge.
var errNoBadge = errors.New("this server cannot record the AI-agent badge yet (it predates ADR 0122 PR 2), and this seat never joins without declaring itself")

// isNoBadge reports a pre-PR 2 server's refusal of the agent field.
func isNoBadge(err error) bool {
	var ae *apiError
	return errors.As(err, &ae) && ae.Status == http.StatusBadRequest && strings.Contains(ae.Message, `unknown field "agent"`)
}

// maxAgentClient is the cut §7 puts on the client name.
const maxAgentClient = 32

// agentClientName turns the MCP clientInfo.name into the badge's client:
// lowercased, cut to 32 characters of [a-z0-9._-], or "unknown". Spaces
// become hyphens so "Claude Code" reads as "claude-code".
func agentClientName(raw string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(raw)) {
		if b.Len() >= maxAgentClient {
			break
		}
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '.', r == '_', r == '-':
			b.WriteRune(r)
		case r == ' ':
			b.WriteByte('-')
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "unknown"
	}
	return out
}
