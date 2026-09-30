package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// conditional_mode_count_test.go — #1590, ADR 0065's 2026-09-27
// amendment: the enumerator half. Jeska's Will's "if you control a
// commander as you cast this spell, you may choose both instead" is a
// bound the announce gate reads per caster, and the enumerator must
// read the same one: never offer "both" to a seat without a commander
// (the engine would refuse it, #544), and do offer it to a seat with
// one.

const oracleJeskasWill = "0fd114c4-092b-4e28-b0dc-ef529f3bc73e"

func TestJeskasWillBothIsOfferedOnlyWithACommander(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	spell := handCard(active, game.Card{
		Name: "Jeska's Will", TypeLine: "Sorcery",
		OracleID: oracleJeskasWill, ManaCost: "{2}{R}",
	})
	lands(g, active, "Mountain", "Mountain", 3)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	combos := explosiveDerailmentModes(castMovesFor(moves, spell), spell)
	if !combos["0"] || !combos["1"] {
		t.Fatalf("each bullet alone is always offered; combos: %v", combos)
	}
	if combos["0,1"] || combos["1,0"] {
		t.Errorf("no commander on the battlefield: both must not be offered; combos: %v", combos)
	}
	dispatchAll(t, g, active.ID, castMovesFor(moves, spell))

	// An opponent's commander on their own side does not count.
	battlefieldCard(g, opp, game.Card{Name: "Their Commander", TypeLine: "Legendary Creature — Elf", Power: 2, Toughness: 2, IsCommander: true})
	moves = legal.EnumerateFor(g, active.ID)
	if combos := explosiveDerailmentModes(castMovesFor(moves, spell), spell); combos["0,1"] {
		t.Errorf("an opponent's commander must not unlock both; combos: %v", combos)
	}

	// Your own commander does.
	battlefieldCard(g, active, game.Card{Name: "My Commander", TypeLine: "Legendary Creature — Human", Power: 2, Toughness: 2, IsCommander: true})
	moves = legal.EnumerateFor(g, active.ID)
	combos = explosiveDerailmentModes(castMovesFor(moves, spell), spell)
	if !combos["0,1"] {
		t.Errorf("with your commander out, both is offered; combos: %v", combos)
	}
	// #544: every offer — "both" included — is one the engine accepts.
	dispatchAll(t, g, active.ID, castMovesFor(moves, spell))
}
