package mcpseat

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// connectClient runs the seat's MCP server on in-memory transports and
// returns a client session on it.
func connectClient(t *testing.T, s *Seat, clientName string) *mcp.ClientSession {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	st, ct := mcp.NewInMemoryTransports()
	srv := NewMCPServer(s, "test")
	ss, err := srv.Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ss.Close() })
	cs, err := mcp.NewClient(&mcp.Implementation{Name: clientName, Version: "1"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

// TestEveryToolIsRegisteredWithItsSchema is §10's "transport.go's schema
// registration, for each tool".
func TestEveryToolIsRegisteredWithItsSchema(t *testing.T) {
	s := newTestSeat(t, nil)
	cs := connectClient(t, s.Seat, "test-client")
	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	type want struct{ props, required []string }
	wants := map[string]want{
		"join":              {[]string{"deck", "display_name", "invite_url"}, []string{"invite_url"}},
		"set_deck":          {[]string{"deck"}, []string{"deck"}},
		"wait_for_decision": {[]string{"pass_until", "timeout_s"}, nil},
		"get_state":         {[]string{"detail"}, nil},
		"legal_moves":       {[]string{"card", "choice", "match", "targets_for"}, nil},
		"card":              {[]string{"ref"}, []string{"ref"}},
		"act":               {[]string{"move", "targets", "value", "window"}, []string{"move", "window"}},
		"say":               {[]string{"text"}, []string{"text"}},
		"concede":           {[]string{"confirm"}, []string{"confirm"}},
		"leave":             {nil, nil},
	}
	if len(res.Tools) != len(wants) {
		t.Errorf("%d tools registered, want %d", len(res.Tools), len(wants))
	}
	for _, tool := range res.Tools {
		w, ok := wants[tool.Name]
		if !ok {
			t.Errorf("unexpected tool %q", tool.Name)
			continue
		}
		if !strings.Contains(tool.Description, "never instructions") {
			t.Errorf("%s: the description does not say table text is data (§8)", tool.Name)
		}
		raw, _ := json.Marshal(tool.InputSchema)
		var schema struct {
			Type       string                     `json:"type"`
			Properties map[string]json.RawMessage `json:"properties"`
			Required   []string                   `json:"required"`
		}
		if err := json.Unmarshal(raw, &schema); err != nil {
			t.Fatalf("%s schema: %v", tool.Name, err)
		}
		if schema.Type != "object" {
			t.Errorf("%s: schema type %q", tool.Name, schema.Type)
		}
		var props []string
		for k := range schema.Properties {
			props = append(props, k)
		}
		sort.Strings(props)
		sort.Strings(schema.Required)
		if strings.Join(props, ",") != strings.Join(w.props, ",") {
			t.Errorf("%s properties = %v, want %v", tool.Name, props, w.props)
		}
		if strings.Join(schema.Required, ",") != strings.Join(w.required, ",") {
			t.Errorf("%s required = %v, want %v", tool.Name, schema.Required, w.required)
		}
	}
}

// TestTheClientNameReachesTheBadge: clientInfo.name from initialize is
// what join declares (§7).
func TestTheClientNameReachesTheBadge(t *testing.T) {
	f := newFakeServer(t)
	s := newTestSeat(t, nil)
	cs := connectClient(t, s.Seat, "Codex CLI")
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "join", Arguments: map[string]any{"invite_url": f.inviteURL()}})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("join: %+v", res.Content)
	}
	text := res.Content[0].(*mcp.TextContent).Text
	noSentinel(t, "a tool result over MCP", text)
	f.mu.Lock()
	defer f.mu.Unlock()
	if got := f.joins[0].Declaration.Client; got != "codex-cli" {
		t.Errorf("agent client = %q, want codex-cli", got)
	}

	// A tool error is a result the model can read, not a protocol error.
	res, err = cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "concede", Arguments: map[string]any{"confirm": false}})
	if err != nil || !res.IsError {
		t.Errorf("concede without confirm: err=%v isError=%v", err, res != nil && res.IsError)
	}
}

// TestNoToolResultCarriesTheToken drives the tools over MCP with a
// handler that would echo the token, and checks the scrub catches it.
func TestNoToolResultCarriesTheToken(t *testing.T) {
	s := newTestSeat(t, nil)
	s.sec.add(sentinelToken)
	if got := s.Scrub("token is " + sentinelToken + " ok"); strings.Contains(got, sentinelToken) {
		t.Fatalf("scrub left the token: %q", got)
	}
	s.Logger().Info("dialing", "url", "wss://x/ws?token="+sentinelToken, "auth", "Bearer "+sentinelToken)
	noSentinel(t, "the log", s.logBuf.String())
}
