package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// protection_source_test.go — #662, the enumerator half.
//
// Protection is the first targeting gate that depends on WHAT is
// casting rather than on who. The enumerator reaches targeting
// through LegalTargetsForEffect, so it inherits the filter the way
// ADR 0038 §5 describes — but only once it passes a source. An
// enumerator that passed a bare seat would happily offer the bot a
// Lightning Bolt at a pro-red creature and the server would then
// refuse the move, which is the #347 / #544 failure mode exactly.
//
// dispatchAll is the other half of every assertion here: it sends
// each enumerated move to the dispatcher and fails if the engine
// rejects one, so "the bot never enumerates an illegal target" is
// checked rather than assumed.

func TestEnumeratorSkipsATargetProtectedFromTheSpellsColour(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	clearHand(active)
	bolt := handCard(active, game.Card{
		Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}",
		Colors: []string{"R"}, OracleID: oracleLightningBolt,
	})
	battlefieldCard(g, active, basic("Mountain", "Mountain"))

	plain := battlefieldCard(g, opp, creature("Plain Bear", "{1}{G}", 2, 2))
	proRed := battlefieldCard(g, opp, protectedCreature("Kor Firewalker", "{W}{W}", "protection from red"))
	proWhite := battlefieldCard(g, opp, protectedCreature("Paladin en-Vec", "{1}{W}{W}", "protection from white"))
	// Your OWN pro-red creature is out too: protection has no "your
	// opponents" clause, which is where it differs from hexproof.
	myProRed := battlefieldCard(g, active, protectedCreature("My Firewalker", "{W}{W}", "protection from red"))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)

	targeted := map[string]bool{}
	for _, m := range moves {
		if m.Source != bolt {
			continue
		}
		var p struct {
			Targets []struct {
				ID string `json:"id"`
			} `json:"targets"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("bad Bolt params: %s", string(m.Params))
		}
		for _, tr := range p.Targets {
			targeted[tr.ID] = true
		}
	}
	if !targeted[plain.String()] {
		t.Error("an unprotected creature must still be enumerated")
	}
	if !targeted[proWhite.String()] {
		t.Error("protection from WHITE does not stop a red Bolt")
	}
	if targeted[proRed.String()] {
		t.Error("the bot proposed a red Bolt at a pro-red creature — the engine would reject it")
	}
	if targeted[myProRed.String()] {
		t.Error("your own pro-red creature is no more boltable than theirs (CR 702.16b has no controller clause)")
	}
}
