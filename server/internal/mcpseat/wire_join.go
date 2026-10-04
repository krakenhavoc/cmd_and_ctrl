package mcpseat

import "strings"

// The join body's `agent` field, ADR 0122 §7 (Delivery PR 2).
//
// Written here to the ADR because PR 2 has not merged into this branch.
// Both join routes (POST /games/{id}/join and POST /join) accept it:
//
//	{"invite_token": "…", "name": "Claude", "agent": {"client": "claude-code"}}
//
// The binary always sends it, and has no flag to leave it out. The server
// records it on the seat for good, and shows the table an "AI agent" chip.
// An older server ignores the unknown field, which is why the PR 2 server
// change matters: the badge is the server's record, not this struct.

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
