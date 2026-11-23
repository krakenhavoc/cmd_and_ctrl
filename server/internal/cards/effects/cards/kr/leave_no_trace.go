package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leave No Trace — Instant {1}{W}:
//
//	"Radiance — Destroy target enchantment and each other enchantment
//	 that shares a color with it."
//
// The radiance set (radianceSet, #5736) with the card's own kind,
// enchantments, read as the spell resolves, whoever controls them. They
// are destroyed as one simultaneous event, so a "whenever an enchantment
// is put into a graveyard" trigger sees them all, and an indestructible
// one stays (CR 702.12b). A colourless target is the only one destroyed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a9367efe-82f2-43e4-9692-205567be6a0c",
		Name:         "Leave No Trace",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target enchantment", effects.Enchantment()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if ids := effects.RadianceSet(ctx, effects.Enchantment()); len(ids) > 0 {
				ctx.Game.DestroyPermanentsForEffect(ids)
			}
			return nil
		},
	})
}
