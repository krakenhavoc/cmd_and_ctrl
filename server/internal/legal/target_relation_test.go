package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

func TestGoblinWelderEnumeratesOnlyRelatedArtifactPairs(t *testing.T) {
	g := newTable(t)
	me, opp := g.Seats[0], g.Seats[1]
	welder := battlefieldCard(g, me, game.Card{Name: "Goblin Welder", TypeLine: "Creature — Goblin Artificer", Power: 1, Toughness: 1, OracleID: "9fc28e1f-f4ea-4bf8-9442-a36dba3b3b29"})
	a := battlefieldCard(g, me, game.Card{Name: "Mine", TypeLine: "Artifact"})
	b := battlefieldCard(g, opp, game.Card{Name: "Theirs", TypeLine: "Artifact"})
	x, y := game.NewCard("grave a", me.ID), game.NewCard("grave b", opp.ID)
	x.TypeLine, y.TypeLine = "Artifact", "Artifact"
	me.Graveyard.PushTop(x)
	opp.Graveyard.PushTop(y)
	advanceTo(t, g, game.StepPrecombatMain)
	seen := map[uuid.UUID]bool{}
	moves := legal.EnumerateFor(g, me.ID)
	for _, m := range moves {
		if m.Type != "activate_ability" {
			continue
		}
		var params struct {
			CardID  string `json:"source_card_id"`
			Targets []struct {
				ID string `json:"id"`
			} `json:"targets"`
		}
		if err := json.Unmarshal(m.Params, &params); err != nil {
			t.Fatal(err)
		}
		if params.CardID != welder.String() {
			continue
		}
		if len(params.Targets) != 2 {
			t.Fatalf("targets: %+v", params.Targets)
		}
		first, second := params.Targets[0].ID, params.Targets[1].ID
		if first == a.String() && second == x.InstanceID.String() {
			seen[a] = true
		} else if first == b.String() && second == y.InstanceID.String() {
			seen[b] = true
		} else {
			t.Fatalf("mismatched offered pair: %+v", params.Targets)
		}
	}
	if !seen[a] || !seen[b] {
		t.Fatalf("offered pairs: %v", seen)
	}
	dispatchAll(t, g, me.ID, moves)
}
