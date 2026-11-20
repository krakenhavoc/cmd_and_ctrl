package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Folk Medicine — Instant {2}{G} (Judgment, #5611):
//
//	"You gain 1 life for each creature you control.
//	 Flashback {1}{W} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Mastery of the Unseen's count (cmdr2S37GainLifePerCreature), taken as
// the spell resolves. Flashback is the shared constructor at the
// off-colour {1}{W}, exiled as it leaves the stack (CR 702.34a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "135eb851-3f2d-45c5-bfb9-49d1e6c98344",
		Name:             "Folk Medicine",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{1}{W}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.Cmdr2S37GainLifePerCreature(ctx.Game, item)
		},
	})
}
