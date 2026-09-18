package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// once_each_turn_object_test.go — #936 from the catalog's side. The
// "only once each turn" gates in the catalog are b11TriggeredThisTurn
// and b15ResolvedThisTurn, and neither had to change: they ask
// Game.TriggeredThisTurn / ResolvedThisTurn, which read the tally
// through the object's own key. This is the proof that the reader
// they share now answers per object (CR 400.7).

const (
	gateProbeOracle = "test-once-each-turn-object-probe"
	gateProbeLabel  = "Gate Probe — gain 1 life"
)

func init() {
	Register(Spec{
		OracleID: gateProbeOracle,
		Name:     "Gate Probe",
		Triggered: []game.TriggeredAbility{
			On(game.EventETB,
				AllOf(AnotherCreatureEnteredUnderYourControl, notYetThisTurn(gateProbeLabel)),
				gateProbeLabel, Do(GainLife{Amount: 1})),
		},
	})
}

// notYetThisTurn is the printed "if this is the first time this has
// happened this turn" clause, written the way the catalog writes it.
func notYetThisTurn(label string) When {
	return func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return !b11TriggeredThisTurn(g, source.InstanceID, label)
	}
}

// TestOnceEachTurnGateReopensForTheReturningPermanent — the bug: the
// gate stayed shut for the rest of the turn because TurnTally is
// keyed by instance ID and an instance ID survives a zone change. The
// permanent that comes back is a new object (CR 400.7) and gets its
// own first time.
func TestOnceEachTurnGateReopensForTheReturningPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	probe := pushCatalogPermanent(g, me.ID, "Gate Probe", "Creature — Test", gateProbeOracle, false)
	start := me.Life

	enterFromHand(t, g, me.ID, "Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if me.Life != start+1 {
		t.Fatalf("life %d after the first creature entered, want %d", me.Life, start+1)
	}

	enterFromHand(t, g, me.ID, "Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if me.Life != start+1 {
		t.Fatalf("life %d after a second creature entered, want %d — the gate is once each turn",
			me.Life, start+1)
	}

	bounceAndReturnForTest(t, g, probe, me.ID)

	enterFromHand(t, g, me.ID, "Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if me.Life != start+2 {
		t.Errorf("life %d after the permanent left and came back, want %d — the returning "+
			"permanent is a new object and its once-each-turn clause may fire again (CR 400.7)",
			me.Life, start+2)
	}
}

// bounceAndReturnForTest sends a permanent to its owner's hand and
// puts it straight back, keeping the instance ID — the #630 shape a
// map keyed by instance ID cannot tell apart.
func bounceAndReturnForTest(t *testing.T, g *game.Game, id, owner uuid.UUID) {
	t.Helper()
	hand := game.ZoneRef{Kind: game.ZoneHand, Owner: owner}
	field := game.ZoneRef{Kind: game.ZoneBattlefield}
	if err := g.MoveCardByID(field, hand, id); err != nil {
		t.Fatalf("bounce to hand: %v", err)
	}
	if err := g.MoveCardByID(hand, field, id); err != nil {
		t.Fatalf("return to the battlefield: %v", err)
	}
}
