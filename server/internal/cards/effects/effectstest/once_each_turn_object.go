package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// once_each_turn_object_test.go — #936 from the catalog's side. The
// "only once each turn" gates in the catalog are b11TriggeredThisTurn
// and Game.ResolvedThisTurn, and neither had to change: they ask
// Game.TriggeredThisTurn / ResolvedThisTurn, which read the tally
// through the object's own key. This is the proof that the reader
// they share now answers per object (CR 400.7).

const (
	GateProbeOracle = "test-once-each-turn-object-probe"
	gateProbeLabel  = "Gate Probe — gain 1 life"
)

func init() {
	effects.Register(effects.Spec{
		OracleID: GateProbeOracle,
		Name:     "Gate Probe",
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB,
				effects.AllOf(effects.AnotherCreatureEnteredUnderYourControl, notYetThisTurn(gateProbeLabel)),
				gateProbeLabel, effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}

// notYetThisTurn is the printed "if this is the first time this has
// happened this turn" clause, written the way the catalog writes it.
func notYetThisTurn(label string) effects.When {
	return func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return !effects.B11TriggeredThisTurn(g, source.InstanceID, label)
	}
}

// bounceAndReturnForTest sends a permanent to its owner's hand and
// puts it straight back, keeping the instance ID — the #630 shape a
// map keyed by instance ID cannot tell apart.
func BounceAndReturnForTest(t *testing.T, g *game.Game, id, owner uuid.UUID) {
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
