package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lt. Cmdr. Geordi La Forge — Legendary Creature — Human Artificer
// {2}{U}, 3/2 (Star Trek, #4973):
//
//	"If you would put one or more charge counters on an artifact you
//	 control, put twice that many charge counters on that artifact
//	 instead."
//
// A CR 614.1a replacement over the counter window, Branching Evolution's
// doubling with Lae'zel's "if YOU would put" test: the counters' placer
// must be Geordi's controller (Event.CounterPlacer), so an opponent's
// effect that charges your artifact is not doubled. The artifact is read
// through the layered view, so a Spacecraft that is also a creature
// still counts, and station's own counters (CR 702.184a) are doubled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cb3b4cfe-fd8d-4a26-905e-8ccd30f033a3",
		Name:         "Lt. Cmdr. Geordi La Forge",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventCounterPlaced},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventCounter || ev.CounterDelta <= 0 || ev.CounterName != game.CounterCharge {
					return false
				}
				if ev.CounterPlacer != src.Controller || ev.CounterPlayer != uuid.Nil {
					return false
				}
				target, ok := g.LookupCardForEffect(ev.CounterTarget)
				return ok && target.Controller == src.Controller && target.IsArtifact()
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.CounterDelta *= 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Lt. Cmdr. Geordi La Forge: twice that many charge counters",
		}},
	})
}
