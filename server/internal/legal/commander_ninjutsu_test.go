package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// commander_ninjutsu_test.go — the enumerator half of #1278. Commander
// ninjutsu (CR 702.49c) reaches the bots through the same activation
// walk plain ninjutsu does, with the COMMAND ZONE as the pile the card
// sits in: abilityZones() has walked it since #1221, so the invariant
// to pin is #544's — offered exactly when the engine would accept it,
// and never before blockers.

const oracleYuriko = "a7043fbd-1dfd-42cf-be4b-cc343d0949e5"

// yurikoInCommand seeds the catalog Yuriko into p's command zone as a
// commander already cast `casts` times.
func yurikoInCommand(p *game.Player, casts int) uuid.UUID {
	id := uuid.New()
	p.Command.PushTop(game.Card{
		InstanceID: id, Name: "Yuriko, the Tiger's Shadow",
		TypeLine: "Legendary Creature — Human Ninja", OracleID: oracleYuriko,
		ManaCost: "{1}{U}{B}", Power: 1, Toughness: 3,
		Owner: p.ID, Controller: p.ID, IsCommander: true,
	})
	p.CommanderCasts[id] = casts
	return id
}

// Nothing offered in declare attackers, exactly one move from the
// command zone once blockers are declared — with only {U}{B} of land
// and a commander tax of {6} on the books, because the tax is not part
// of an activation (CR 903.8) — and the move is one the dispatcher
// accepts.
func TestCommanderNinjutsuIsEnumeratedFromTheCommandZone(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	battlefieldCard(g, active, basic("Island", "Island"))
	battlefieldCard(g, active, basic("Swamp", "Swamp"))
	attacker := battlefieldCard(g, active, vanillaAttacker("Sneaky Rat"))
	yuriko := yurikoInCommand(active, 3)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), yuriko); len(acts) != 0 {
		t.Fatalf("commander ninjutsu offered in declare_attackers, where nothing is unblocked yet: %v", labels(acts))
	}

	advanceTo(t, g, game.StepDeclareBlockers)
	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, yuriko)
	if len(acts) != 1 {
		t.Fatalf("want exactly one commander ninjutsu move from the command zone, got %v", labels(acts))
	}
	dispatchAll(t, g, active.ID, acts)
}

// Another seat's commander is not this seat's move: the command zone's
// "you" is its owner (CR 108.4), and the walk visits only the seat's
// own piles.
func TestCommanderNinjutsuIsNotEnumeratedForAnotherSeatsCommander(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	clearHand(active)
	battlefieldCard(g, active, basic("Island", "Island"))
	battlefieldCard(g, active, basic("Swamp", "Swamp"))
	attacker := battlefieldCard(g, active, vanillaAttacker("Sneaky Rat"))
	theirs := yurikoInCommand(other, 0)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), theirs); len(acts) != 0 {
		t.Fatalf("offered another seat's commander ninjutsu: %v", labels(acts))
	}
}
