package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Optimization Annex — {1}{U} Artifact Creature — Borg 1/3 (Star Trek
// Commander, #4974):
//
//	"If one or more charge counters would be put on an artifact you
//	 control, that many plus one charge counters are put on it instead."
//
// Lt. Cmdr. Geordi La Forge's charge-counter window with Conclave
// Mentor's "that many plus one" (CR 614.1a). Unlike Geordi it does not
// ask who puts the counters: an opponent's effect that charges your
// artifact gets the extra one too. The artifact is read through the
// layered view, so the Annex itself and a Spacecraft that has become a
// creature both count, and station's counters (CR 702.184a) get the
// extra one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f38914f-7870-44bf-887e-ee893e98989c",
		Name:         "Optimization Annex",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventCounterPlaced},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventCounter || ev.CounterDelta <= 0 || ev.CounterName != game.CounterCharge {
					return false
				}
				if ev.CounterPlayer != uuid.Nil {
					return false
				}
				target, ok := g.LookupCardForEffect(ev.CounterTarget)
				return ok && target.Controller == src.Controller && target.IsArtifact()
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.CounterDelta++
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Optimization Annex: one more charge counter",
		}},
	})
}
