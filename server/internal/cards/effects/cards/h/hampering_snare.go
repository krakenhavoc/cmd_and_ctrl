package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hampering Snare — Instant {1}{U} (Ikoria: Lair of Behemoths, #5512):
//
//	"Creatures your opponents control get -2/-0 until end of turn.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Uncomfortable Chill's shrink without the draw, and cycling from hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4fd3dafe-e6a4-49d5-950a-e2fcc77eda98",
		Name:         "Hampering Snare",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(effects.BoostUntilEOT{
			Match: effects.And(effects.Creature(), effects.OpponentControls()), Power: -2,
			Label: "Hampering Snare — creatures your opponents control get -2/-0",
		}),
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
