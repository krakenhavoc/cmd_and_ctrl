package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prismatic Ending — Sorcery {X}{W} (Secrets of Strixhaven Mystical
// Archive, #4834):
//
//	"Converge — Exile target nonland permanent if its mana value is less
//	 than or equal to the number of colors of mana spent to cast this
//	 spell."
//
// Painful Truths' converge count (WantsDistinctColors pays the generic
// half, here {X}, with colours not yet spent; ReadsManaSpent makes a
// waived cast ask what was paid) on a removal spell. The mana value is not
// a targeting condition (the ruling): any nonland permanent is a legal
// target and the comparison is made as the spell resolves, on the target
// as it is then (a permanent's {X} is 0, CR 202.3e). A permanent that is
// too expensive stays where it is. A copy of the spell was cast with no
// mana, so it counts zero colours (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:            "2cb98ca9-d7bb-416b-a17e-ee5f8e4d78f2",
		Name:                "Prismatic Ending",
		ReadsManaSpent:      game.ReadsOwnPayment,
		WantsDistinctColors: true,
		Completeness:        CompletenessFull,
		Targets:             TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			id := item.Targets[0].ID
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok || !ManaValueLE(ctx.ColorsSpentCount())(ctx.Game, ctx.Controller(), c) {
				return nil
			}
			return ExileTarget{Target: id}.Apply(ctx)
		},
	})
}
