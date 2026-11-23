package effectstest

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// opposition_agent_test.go — #3998, ADR 0166's 2026-10-16 amendment:
// Opposition Agent controls an opponent while they search (CR 723.2)
// and exiles what they find, for its controller to play.
//
// The searcher is the active seat (seat 0), the Agent's controller is
// seat 1, and seat 2 is a third player who is neither.

const (
	OppositionAgentOracle = "1f438b8f-fe23-4f3b-ab2e-f6c33676c462"
	OaExplosiveVegetation = "a0abd957-01a0-4aa9-8fc8-0e6840d21606"
	OaBuriedAlive         = "8203c621-a1a0-4865-8c9a-0d4064c86107"
	OaVampiricTutor       = "ededbdae-d9dc-4206-9335-d7158f2d7700"
	OaForest              = "Basic Land — Forest"
	OaBear                = "Creature — Bear"
)

// oaTable puts an Opposition Agent on seat 1's battlefield.
func OaTable(t *testing.T) (g *game.Game, searcher, ctl, third *game.Player, agent uuid.UUID) {
	t.Helper()
	g = NewCatalogGame(t)
	searcher = g.Seats[g.Turn.ActiveSeat]
	ctl, third = g.Seats[(g.Turn.ActiveSeat+1)%4], g.Seats[(g.Turn.ActiveSeat+2)%4]
	agent = PushCatalogPermanent(g, ctl.ID, "Opposition Agent", "Creature — Human Rogue", OppositionAgentOracle, false)
	return g, searcher, ctl, third, agent
}

// oaAnswer answers the open search prompt through the actions layer, as
// `caller`, for `player`.
func OaAnswer(g *game.Game, c *game.PendingChoice, caller, player uuid.UUID, picks ...uuid.UUID) error {
	ids := make([]string, 0, len(picks))
	for _, id := range picks {
		ids = append(ids, id.String())
	}
	raw, _ := json.Marshal(map[string]any{"choice_id": c.ID.String(), "card_ids": ids})
	return actions.Dispatch(g, actions.Action{Type: actions.TypeResolveChoice, Player: player, Caller: caller, Params: raw})
}

// oaInExile reports whether id is in exile, face up.
func OaInExile(g *game.Game, id uuid.UUID) bool {
	for _, c := range g.Exile.Cards {
		if c.InstanceID == id {
			return !c.FaceDown
		}
	}
	return false
}

// oaOpenSearch returns the open search prompt the searcher owes, failing
// the test when there is none, and checks the control is live.
func OaOpenSearch(t *testing.T, g *game.Game, searcher, ctl *game.Player) *game.PendingChoice {
	t.Helper()
	c := SearchChoiceFor(g, searcher.ID)
	if c == nil {
		t.Fatal("no search prompt for the searcher: prompts keep their real chooser (ADR 0166 §2)")
	}
	if got := g.ControllerOfPlayer(searcher.ID); got != ctl.ID {
		t.Fatalf("while searching, the searcher is controlled by %v, want the Agent's controller", got)
	}
	if got := g.ControlUntilOfPlayer(searcher.ID); got != game.ControlUntilSearch {
		t.Fatalf("control kind = %q, want %q", got, game.ControlUntilSearch)
	}
	return c
}

// oaAssertControlOver checks the control ended with the search.
func OaAssertControlOver(t *testing.T, g *game.Game, searcher *game.Player) {
	t.Helper()
	if got := g.ControllerOfPlayer(searcher.ID); got != searcher.ID {
		t.Errorf("after the search, the searcher is controlled by %v", got)
	}
	PassPriorityAroundTable(t, g)
	if len(g.PlayerControls) != 0 {
		t.Errorf("a control record outlived the search: %+v", g.PlayerControls)
	}
}
