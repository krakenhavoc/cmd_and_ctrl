package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scrollshift — Instant {2}{W} (March of the Machine, S82 #5477):
//
//	"Exile up to one target artifact, creature, or enchantment you control,
//	 then return it to the battlefield under its owner's control.
//	 Draw a card."
//
// "Up to one": with no target chosen the spell still resolves and you still
// draw (CR 601.2c, 608.2b). The permanent returns as a new object (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f762b9f-7525-4ff8-bc1d-fbab66fd4013",
		Name:         "Scrollshift",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets: TargetPermanent("up to one target artifact, creature, or enchantment you control",
			And(Or(Artifact(), Creature(), Enchantment()), YouControl())).WithCount(0, 1),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if id, ok := b16FirstLegalTargetCard(ctx); ok {
				if err := (Flicker{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
