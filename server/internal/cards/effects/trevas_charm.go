package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "e9795590-0918-49f9-b37c-1278c8616ab1",
		Name:         "Treva's Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Destroy target enchantment.",
				TargetPermanent("target enchantment", Enchantment()),
				DestroyTheModesTarget),
			ModeDoing("Exile target attacking creature.",
				TargetCreature("target attacking creature", AttackingCreature()),
				ExileTheModesTarget),
			ModeWithPurpose(ModeDoing("Draw a card, then discard a card.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return lootOne(ctx.Game, item, 1)
				}), game.Purpose{Draws: 1, Discards: 1}),
		),
	})
}
