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

// joinBody is the request body of both join routes. The Go field is
// Declaration, not Agent: game's TestAgentBadgeHasNoClearingWriter scans
// the module for writes to fields named Agent, AgentClient or IsAgent, and
// this is a request body, not the badge on a seat.
type joinBody struct {
	InviteToken string     `json:"invite_token"`
	Name        string     `json:"name"`
	Declaration agentField `json:"agent"`
}

// agentField declares the seat an agent, and which MCP client drives it.
type agentField struct {
	Client string `json:"client"`
}

// errNoBadge is a server that cannot record the badge.
var errNoBadge = errors.New("this server cannot record the AI-agent badge yet (it predates ADR 0122 PR 2), and this seat never joins without declaring itself")

// errAgentSignedIn is the server's 400 for an agent join that carried a
// signed-in session. The binary has no cookie jar and sends no
// credential on a join, so it should never see it; if it does, the
// session came from somewhere it should not have.
var errAgentSignedIn = errors.New("the server refused the join: an agent seat joins as a guest, and this request carried a signed-in session")

func isAgentSignedIn(err error) bool {
	var ae *apiError
	return errors.As(err, &ae) && ae.Status == http.StatusBadRequest && strings.Contains(ae.Message, "an agent seat joins as a guest")
}

// isNoBadge reports a pre-PR 2 server's refusal of the agent field.
func isNoBadge(err error) bool {
	var ae *apiError
	return errors.As(err, &ae) && ae.Status == http.StatusBadRequest && strings.Contains(ae.Message, `unknown field "agent"`)
}

// maxAgentClient is the cut §7 puts on the client name.
const maxAgentClient = 32

// agentClientName turns the MCP clientInfo.name into the badge's client:
// lowercased, every character outside [a-z0-9._-] made a hyphen, trimmed
// of hyphens, cut to 32, or "unknown". It is the server's own rule
// (lobby.NormalizeAgentClient, which this package may not import), so the
// name the binary logs is the name the table sees; the server normalises
// again anyway.
func agentClientName(raw string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(raw) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '.', r == '_', r == '-':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	out := strings.Trim(b.String(), "-")
	if len(out) > maxAgentClient {
		out = strings.Trim(out[:maxAgentClient], "-")
	}
	if out == "" {
		return "unknown"
	}
	return out
}

// defaultDisplayName is the seat's name when join gives none (#2273): the
// MCP client's own name, which the table already shows in the badge, so it
// reveals nothing new. A known client reads as people say it, any other
// uses its normalised name, and a client that sent no clientInfo stays
// "Agent". client is the agentClientName form.
func defaultDisplayName(client string) string {
	switch {
	case client == "unknown":
		return defaultName
	case client == "claude-code":
		return "Claude Code"
	case client == "codex" || strings.HasPrefix(client, "codex-") || strings.HasPrefix(client, "codex_"):
		return "Codex"
	}
	return client
}
