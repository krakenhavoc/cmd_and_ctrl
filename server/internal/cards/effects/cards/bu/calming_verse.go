package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Calming Verse — Sorcery {3}{G} (Prophecy, S82 OLD b672, #5617):
//
//	"Destroy all enchantments you don't control. Then if you control an untapped
//	 land, destroy all enchantments you control."
//
// Back to Nature's sweep, split by controller (Primeval Light's controller
// predicate): the enchantments of everyone else go first, and the controller's
// untapped land is looked for in the sweep's registered tail (Blood Money's
// shape), once it has finished (CR 608.2c), so a land that was one of the
// destroyed enchantments no longer counts. "You" is the spell's controller
// as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c70ae479-0bd1-481c-885f-16bfa096cf58",
		Name:         "Calming Verse",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepEnchantments, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{
				Match:    effects.And(effects.Enchantment(), effects.Not(effects.ControlledBy(ctx.Controller()))),
				ThenTail: ctx.ItemTail(calmingVerseThen),
			}.Apply(ctx)
		},
	})
}

// calmingVerseThen is "Then if you control an untapped land, destroy all
// enchantments you control", told once the first sweep has finished (a commander
// waiting on CR 903.9 holds it back). An on-disk key: never renamed, never reused.
var calmingVerseThen = effects.MassTail("mass/s82-old-b672-calming-verse-then", func(ctx *effects.Context, _ []game.Card, _ int) error {
	you := ctx.Controller()
	if effects.CountControlled(ctx.Game, you, func(c game.Card) bool { return c.IsLand() && !c.Tapped }) == 0 {
		return nil
	}
	return effects.DestroyAllMatching{Match: effects.And(effects.Enchantment(), effects.ControlledBy(you))}.Apply(ctx)
})
