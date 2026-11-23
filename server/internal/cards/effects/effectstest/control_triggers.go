package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// control_triggers_test.go — the three printed shapes #930 adds, each
// on its own probe card so one event cannot satisfy two of them at
// once. The engine half (one event per delta, from the one
// materialise step) is pinned in game/control_changed_event_test.go.

const (
	LoseControlProbeOracle  = "test-control-lose-probe"
	GainControlProbeOracle  = "test-control-gain-probe"
	OwnedControlProbeOracle = "test-control-owned-probe"
)

func init() {
	effects.Register(effects.Spec{
		OracleID: LoseControlProbeOracle,
		Name:     "Betrayer Probe",
		Triggered: []game.TriggeredAbility{
			effects.WhenYouLoseControlOfThis("Betrayer Probe — draw two cards", effects.Do(effects.DrawCards{N: 2})),
		},
	})
	effects.Register(effects.Spec{
		OracleID: GainControlProbeOracle,
		Name:     "Turncoat Probe",
		Triggered: []game.TriggeredAbility{
			effects.WhenYouGainControlOfThis("Turncoat Probe — gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
	})
	effects.Register(effects.Spec{
		OracleID: OwnedControlProbeOracle,
		Name:     "Donor Probe",
		Triggered: []game.TriggeredAbility{
			effects.WheneverAnOpponentGainsControlOfAPermanentYouOwn(
				"Donor Probe — gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}

// stealForTest hands `target` to `to` until end of turn and settles
// the layer pass, which is where the control change — and its event —
// happens.
func StealForTest(t *testing.T, g *game.Game, target, to uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if !g.GainControlForEffect(uuid.New(), target, to, g.UntilEndOfTurnDuration(), "test — steal") {
			t.Fatal("GainControlForEffect refused a battlefield permanent")
		}
	})
	g.ReadSnapshot(func() {})
}

// pendingTriggerFor returns the harvested trigger queued for
// `cardID`, before the next priority boundary drains it onto the
// stack. The queue is where a control-change trigger lands: the
// event is emitted from the layer pass, not from a resolution.
func PendingTriggerFor(g *game.Game, cardID uuid.UUID) *game.StackItem {
	for _, item := range g.PendingTriggers {
		if item != nil && item.SourceCardID == cardID {
			return item
		}
	}
	return nil
}
