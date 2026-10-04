package protocol

// agent_badge_view_test.go: ADR 0122 §7. PlayerView carries the agent
// badge, public and identical for every viewer, like is_bot.

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestPlayerViewCarriesTheAgentBadgeForEveryViewer(t *testing.T) {
	g := game.NewGame()
	var ids []string
	for i := range 3 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Commander %d", i+1), uuid.Nil)}
		for j := range 10 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", j+1), uuid.Nil))
		}
		p, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck)
		if err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
		ids = append(ids, p.ID.String())
	}
	agentID := ids[1]
	if err := g.SetAgent(uuid.MustParse(agentID), "claude-code"); err != nil {
		t.Fatalf("SetAgent: %v", err)
	}
	if err := g.SetBot(uuid.MustParse(ids[2]), "random", ""); err != nil {
		t.Fatalf("SetBot: %v", err)
	}
	if err := g.Start(rand.New(rand.NewPCG(5, 6))); err != nil {
		t.Fatalf("Start: %v", err)
	}

	viewers := map[string]string{
		"the agent's own seat": agentID,
		"another seat":         ids[0],
		"the bot seat":         ids[2],
		"a spectator":          SpectatorViewerID,
		"the admin":            "",
	}
	for label, viewer := range viewers {
		v := ViewOfGameFor(g, viewer)
		for _, s := range v.Seats {
			isAgent := s.ID == agentID
			if s.IsAgent != isAgent {
				t.Errorf("%s: seat %s is_agent = %v, want %v", label, s.Name, s.IsAgent, isAgent)
			}
			wantClient := ""
			if isAgent {
				wantClient = "claude-code"
			}
			if s.AgentClient != wantClient {
				t.Errorf("%s: seat %s agent_client = %q, want %q", label, s.Name, s.AgentClient, wantClient)
			}
			if s.IsAgent && s.IsBot {
				t.Errorf("%s: seat %s is both a bot and an agent", label, s.Name)
			}
		}
	}

	// On the wire: the two keys on the agent's seat, and absent on
	// every other seat (omitempty), so a table with no agent sends
	// exactly what it sent before.
	raw, err := json.Marshal(ViewOfGameFor(g, ids[0]))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var wire struct {
		Seats []map[string]any `json:"seats"`
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, s := range wire.Seats {
		if s["id"] == agentID {
			if s["is_agent"] != true || s["agent_client"] != "claude-code" {
				t.Errorf("agent seat on the wire: is_agent %v, agent_client %v", s["is_agent"], s["agent_client"])
			}
			continue
		}
		for k := range s {
			if strings.Contains(k, "agent") {
				t.Errorf("seat %v carries %q on the wire", s["name"], k)
			}
		}
	}
}
