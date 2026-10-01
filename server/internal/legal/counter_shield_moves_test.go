package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// counter_shield_moves_test.go — ADR 0106 PR 3 (#1806), the enumerator
// half: the two new activations that take a target or a choice, Vexing
// Shusher's "{R/G}: Target spell can't be countered." and Domri, Anarch
// of Bolas's loyalty abilities, are offered to the bot exactly when the
// engine accepts them (#544), with the spell on the stack as the target.

const (
	oracleVexingShusher = "a20a7cf8-2075-47ad-9229-36264b112e61"
	oracleDomriAnarch   = "afc2269c-d3b5-487d-9445-800c7a8e526b"
	oracleMistrise      = "339f5334-b65a-445a-a016-20e997e0b4bb"
)

// targetsNamed reports whether a move's params name this object.
func targetsNamed(m legal.Move, id uuid.UUID) bool {
	return strings.Contains(string(m.Params), id.String())
}

func TestVexingShusherIsOfferedAgainstASpellOnTheStack(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	clearHand(active)
	clearHand(opp)
	shusher := battlefieldCard(g, active, game.Card{Name: "Vexing Shusher", TypeLine: "Creature — Goblin Shaman",
		ManaCost: "{R/G}{R/G}", OracleID: oracleVexingShusher, Power: 2, Toughness: 2})
	battlefieldCard(g, active, basic("Forest", "Forest"))
	battlefieldCard(g, active, basic("Forest", "Forest"))
	sorcery := handCard(active, game.Card{Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}"})
	advanceTo(t, g, game.StepPrecombatMain)

	if n := len(activationsOf(legal.EnumerateFor(g, active.ID), shusher)); n != 0 {
		t.Errorf("no spell on the stack, nothing to target: %d offered", n)
	}
	if err := g.CastSpell(active.ID, sorcery, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, shusher)
	if len(acts) != 1 || !targetsNamed(acts[0], sorcery) {
		t.Fatalf("one activation, against the sorcery: %v", labels(acts))
	}
	dispatchAll(t, g, active.ID, moves)
}

func TestDomriAndMistriseVillageAreOfferedAndAccepted(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	clearHand(active)
	mine := battlefieldCard(g, active, creature("Your Bear", "{1}{G}", 2, 2))
	theirs := battlefieldCard(g, opp, creature("Their Bear", "{1}{G}", 2, 2))
	domri := walker(g, active, "Domri, Anarch of Bolas", oracleDomriAnarch, 3)
	village := battlefieldCard(g, active, game.Card{Name: "Mistrise Village", TypeLine: "Land", OracleID: oracleMistrise})
	battlefieldCard(g, active, basic("Island", "Island"))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, domri)
	var plus, minus int
	for _, m := range acts {
		switch {
		case m.Cost != nil && m.Cost.Loyalty > 0:
			plus++
		case m.Cost != nil && m.Cost.Loyalty < 0:
			minus++
			if !targetsNamed(m, mine) || !targetsNamed(m, theirs) {
				t.Errorf("the −2 names your creature and theirs: %s", m.Params)
			}
		}
	}
	if plus != 1 || minus < 1 {
		t.Fatalf("Domri's +1 and −2: %v", labels(acts))
	}
	if n := len(activationsOf(moves, village)); n != 1 {
		t.Errorf("Mistrise Village's promise: %d offered", n)
	}
	dispatchAll(t, g, active.ID, moves)
}
