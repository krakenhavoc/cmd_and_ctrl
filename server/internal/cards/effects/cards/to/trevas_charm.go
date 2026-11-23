package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treva's Charm — Instant {G}{W}{U} (Planeshift, #5615):
//
//	"Choose one —
//	 • Destroy target enchantment.
//	 • Exile target attacking creature.
//	 • Draw a card, then discard a card."
//
// A modal spell (CR 700.2) whose bullets are the catalog's mode bodies:
// DestroyTheModesTarget, ExileTheModesTarget over Ancient Stone Idol's
// AttackingCreature (so it is cast during combat, and a creature removed
// from combat in response is no longer a legal target, CR 608.2b), and
// Divining Duelist's loot, which draws before it discards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9795590-0918-49f9-b37c-1278c8616ab1",
		Name:         "Treva's Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Destroy target enchantment.",
				effects.TargetPermanent("target enchantment", effects.Enchantment()),
				effects.DestroyTheModesTarget),
			effects.ModeDoing("Exile target attacking creature.",
				effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
				effects.ExileTheModesTarget),
			effects.ModeWithPurpose(effects.ModeDoing("Draw a card, then discard a card.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.LootOne(ctx.Game, item, 1)
				}), game.Purpose{Draws: 1, Discards: 1}),
		),
	})
}
