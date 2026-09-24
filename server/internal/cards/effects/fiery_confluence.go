package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fiery Confluence — Sorcery {2}{R}{R}:
//
//	"Choose three. You may choose the same mode more than once.
//	 • Fiery Confluence deals 1 damage to each creature.
//	 • Fiery Confluence deals 2 damage to each opponent.
//	 • Destroy target artifact."
//
// Mystic Confluence's shape (CR 700.2d, ChooseNRepeating): three
// picks from a three-bullet menu, any bullet repeatable, each
// occurrence's body running in announce order. Choosing the sweep
// bullet three times is three separate 1-damage passes rather than
// one 3-damage pass — the same distinction Languish and Pyroclasm
// already rely on (a 3-toughness creature that lives through two
// passes still dies to the third; a 2-toughness one is already gone
// and the third pass finds nothing left to hit).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3c22e031-4804-4c31-bd3c-c3f29d456b34",
		Name:         "Fiery Confluence",
		Completeness: CompletenessFull,
		Modes: ChooseNRepeating("Choose three (you may choose the same mode more than once)", 3, 3,
			ModeDoing("Fiery Confluence deals 1 damage to each creature.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return damageEachMatching(ctx, Creature(), 1)
				}),
			ModeDoing("Fiery Confluence deals 2 damage to each opponent.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return damageToEachOpponent(ctx.Game, item, 2)
				}),
			ModeDoing("Destroy target artifact.",
				TargetPermanent("target artifact", Artifact()),
				DestroyTheModesTarget),
		),
	})
}
