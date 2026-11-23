package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Auron's Inspiration — Instant {2}{W} (Final Fantasy, #4462):
//
//	"Attacking creatures get +2/+0 until end of turn.
//	 Flashback {2}{W}{W}"
//
// The set of attacking creatures is locked in as the spell resolves (CR
// 611.2c): every attacking creature, whoever controls it, gets the bonus,
// and a creature that starts attacking later in the turn (an added
// combat) does not. Flashback is the engine's alternative cost (CR
// 702.34): cast from the graveyard, then exiled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "53a4d9ee-6dfa-4a8b-a1e4-7f671a666188",
		Name:             "Auron's Inspiration",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{2}{W}{W}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{
				Match: effects.AttackingCreature(),
				Power: 2,
				Label: "Auron's Inspiration — attacking creatures get +2/+0",
			}.Apply(ctx)
		},
	})
}
