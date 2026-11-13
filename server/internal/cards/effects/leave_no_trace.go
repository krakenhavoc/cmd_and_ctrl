package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a9367efe-82f2-43e4-9692-205567be6a0c",
		Name:         "Leave No Trace",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target enchantment", Enchantment()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if ids := radianceSet(ctx, Enchantment()); len(ids) > 0 {
				ctx.Game.DestroyPermanentsForEffect(ids)
			}
			return nil
		},
	})
}
