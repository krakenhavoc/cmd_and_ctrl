package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// combat_interacts_test.go — #2871. Crewing a Vehicle and animating a
// manland are not answers to a spell, so they never set interacts, but
// they make a blocker, so they set combat_interacts, which the client
// counts in a combat window only. Mind Stone's draw sets neither.
const (
	oracleSmugglersCopterCI   = "49136bdc-bc50-49a2-999a-1ef9c16ea130"
	oracleRestlessAnchorageCI = "91320daf-f69c-4350-b0fc-4bb37a6904b1"
	oracleVodalianKnightsCI   = "f6daa28f-e5ce-440c-8dc2-b36f59ae0d4f"
)

func TestCombatInteractsIsSetOnCombatAbilitiesOnly(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	battlefieldCard(g, seat, basic("Plains", "Plains"))
	battlefieldCard(g, seat, basic("Island", "Island"))
	battlefieldCard(g, seat, basic("Island", "Island"))
	battlefieldCard(g, seat, game.Card{
		Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2,
	})
	copter := battlefieldCard(g, seat, game.Card{
		Name: "Smuggler's Copter", TypeLine: "Artifact — Vehicle", ManaCost: "{2}",
		OracleID: oracleSmugglersCopterCI, Power: 3, Toughness: 3,
	})
	anchorage := battlefieldCard(g, seat, game.Card{
		Name: "Restless Anchorage", TypeLine: "Land", OracleID: oracleRestlessAnchorageCI,
	})
	knights := battlefieldCard(g, seat, game.Card{
		Name: "Vodalian Knights", TypeLine: "Creature — Merfolk Knight", ManaCost: "{1}{U}{U}",
		OracleID: oracleVodalianKnightsCI, Power: 2, Toughness: 2,
	})
	stone := battlefieldCard(g, seat, game.Card{
		Name: "Mind Stone", TypeLine: "Artifact", ManaCost: "{2}", OracleID: oracleMindStoneHT,
	})

	moves := legal.EnumerateFor(g, seat.ID)
	dispatchAll(t, g, seat.ID, moves)

	check := func(name string, source uuid.UUID, want bool) {
		t.Helper()
		got := movesFrom(moves, source, legal.KindActivate)
		if len(got) == 0 {
			t.Fatalf("%s offered no activation: %v", name, labels(moves))
		}
		for _, m := range got {
			if m.Interacts {
				t.Errorf("%s: interacts set on a combat-only ability (%s)", name, m.Label)
			}
			if m.CombatInteracts != want {
				t.Errorf("%s: combat_interacts = %v, want %v (%s)", name, m.CombatInteracts, want, m.Label)
			}
		}
	}
	check("Smuggler's Copter's crew", copter, true)
	check("Restless Anchorage's animation", anchorage, true)
	check("Vodalian Knights' flying", knights, true)
	check("Mind Stone's draw", stone, false)

	for _, m := range moves {
		if m.CombatInteracts && (m.Interacts || m.HasTargets) {
			t.Errorf("%q: combat_interacts set beside interacts or has_targets", m.Label)
		}
		if m.Kind == legal.KindMana && m.CombatInteracts {
			t.Errorf("mana move %q has combat_interacts set", m.Label)
		}
	}
}
