package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Collective Override — Instant {2}{W}{U} (Star Trek Commander, #4974):
//
//	"Choose one. If you control a commander as you cast this spell, you
//	 may choose both instead.
//	 • Create a 1/1 colorless Borg artifact creature token for each
//	   attacking creature.
//	 • Draw a card for each attacking creature."
//
// The mode count is Jeska's Will's conditional one (OrUpToIf(2,
// YouControlACommander)), read as the spell is cast (CR 601.2b). Each
// bullet counts the attacking creatures, every player's, as it resolves;
// the Borg tokens are created after the count, not attacking, so with both
// bullets chosen the draw counts the same creatures. Outside combat no
// creature is attacking and each bullet does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "75af4154-d873-45d0-9ff2-c977169012a3",
		Name:         "Collective Override",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Create a 1/1 colorless Borg artifact creature token for each attacking creature.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					n := trekB39AttackingCreatureCount(ctx)
					if n <= 0 {
						return nil
					}
					return CreateToken{Template: TokenCard("1/1 colorless Borg artifact"), N: n}.Apply(ctx)
				}),
			ModeDoing("Draw a card for each attacking creature.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					n := trekB39AttackingCreatureCount(ctx)
					if n <= 0 {
						return nil
					}
					return DrawCards{N: n}.Apply(ctx)
				}),
		).OrUpToIf(2, YouControlACommander),
	})
}
