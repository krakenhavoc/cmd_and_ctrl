package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Surge of Thoughtweft — Kindred Instant — Kithkin {1}{W} (Lorwyn, S82 OLD
// b546, #5593):
//
//	"Creatures you control get +1/+1 until end of turn. If you control a
//	 Kithkin, draw a card."
//
// The creatures are picked as the spell resolves (CR 611.2c); the draw
// is decided afterwards (CR 608.2c), and a changeling you control is a
// Kithkin (CR 702.73a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76aea8c6-3bab-45e9-afb4-65e5597bb2fa",
		Name:         "Surge of Thoughtweft",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1, Label: "Surge of Thoughtweft — +1/+1"},
			effects.S82OldB546DrawIfYouControl("Kithkin"),
		),
	})
}
