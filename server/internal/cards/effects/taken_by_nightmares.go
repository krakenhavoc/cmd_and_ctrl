package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Taken by Nightmares — Instant {2}{B}{B} (Wilds of Eldraine, #4471):
//
//	"Exile target creature. If you control an enchantment, scry 2."
//
// Angelic Edict's exile, then Roadside Reliquary's "if you control an
// enchantment", asked of the board as the spell resolves, after the
// exile (CR 608.2c): an enchantment you control then scries, whatever
// happened to the creature. The scry does not depend on the exile, so a
// creature a replacement kept from exile still lets you scry. A target
// that left in response leaves the spell with no legal target, and it
// does nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e5557c9f-e4c2-41af-9fcc-371409e1dec6",
		Name:         "Taken by Nightmares",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := exileTheFirstTargetIfLegal(ctx); err != nil {
				return err
			}
			if countControlled(ctx.Game, item.Controller, func(c game.Card) bool { return c.IsEnchantment() }) == 0 {
				return nil
			}
			return Scry{N: 2}.Apply(ctx)
		},
	})
}
