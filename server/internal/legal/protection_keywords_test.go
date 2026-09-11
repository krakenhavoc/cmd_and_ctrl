package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// protection_keywords_test.go — S23. The enumerator is the bot's
// whole view of what it may do, and #347 was the lesson that a gate
// added to the engine and not to the enumerator produces a bot that
// proposes moves the server then rejects. Hexproof and shroud are
// the same shape of gate, so they get the same guard.

func protectedCreature(name, cost string, keywords ...string) game.Card {
	c := creature(name, cost, 2, 2)
	c.Keywords = keywords
	return c
}

// TestEnumeratorSkipsHexproofAndShroudTargets: the bot must not
// offer Lightning Bolt at an opponent's hexproof creature, nor at a
// shrouded one of its own, but must still offer it at its own
// hexproof creature (the asymmetry) and at every player.
func TestEnumeratorSkipsHexproofAndShroudTargets(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	clearHand(active)
	bolt := handCard(active, game.Card{
		Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}",
		OracleID: oracleLightningBolt,
	})
	battlefieldCard(g, active, basic("Mountain", "Mountain"))

	plain := battlefieldCard(g, opp, creature("Plain Bear", "{1}{G}", 2, 2))
	theirHexproof := battlefieldCard(g, opp, protectedCreature("Slippery Bogle", "{G/U}", "hexproof"))
	myHexproof := battlefieldCard(g, active, protectedCreature("My Bogle", "{G/U}", "hexproof"))
	myShroud := battlefieldCard(g, active, protectedCreature("Silhana Ledgewalker", "{1}{G}", "shroud"))
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
		t.Errorf("an unprotected creature must still be enumerated")
	}
	if !targeted[myHexproof.String()] {
		t.Errorf("your own hexproof creature is a legal target for your own Bolt")
	}
	if targeted[theirHexproof.String()] {
		t.Errorf("the bot proposed a Bolt at an opponent's hexproof creature — the engine would reject it")
	}
	if targeted[myShroud.String()] {
		t.Errorf("the bot proposed a Bolt at a shrouded creature — shroud stops everyone")
	}
}

// TestEnumeratorStillSacrificesProtectedCreatures is the boundary
// the gate must not cross: Village Rites sacrifices, it does not
// target, so a hexproof or shrouded creature stays a legal cost.
// Narrowing the enumerator here would make the bot unable to pay a
// cost the engine accepts — the same drift, pointing the other way.
func TestEnumeratorStillSacrificesProtectedCreatures(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	rites := handCard(active, game.Card{
		Name: "Village Rites", TypeLine: "Instant", ManaCost: "{B}",
		OracleID: oracleVillageRites,
	})
	battlefieldCard(g, active, basic("Swamp", "Swamp"))
	bogle := battlefieldCard(g, active, protectedCreature("My Bogle", "{G/U}", "hexproof"))
	ledge := battlefieldCard(g, active, protectedCreature("Ledgewalker", "{1}{G}", "shroud"))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	sacs := map[string]bool{}
	for _, m := range moves {
		if m.Source != rites {
			continue
		}
		var p struct {
			SacrificeIDs []string `json:"sacrifice_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("bad Village Rites params: %s", string(m.Params))
		}
		for _, id := range p.SacrificeIDs {
			sacs[id] = true
		}
	}
	if !sacs[bogle.String()] || !sacs[ledge.String()] {
		t.Errorf("protected creatures must stay sacrificeable: %v", sacs)
	}
}
