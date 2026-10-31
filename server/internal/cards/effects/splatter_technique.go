package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Splatter Technique — Sorcery {1}{U}{U}{R}{R} (Secrets of Strixhaven,
// #4456):
//
//	"Choose one —
//	 • Draw four cards.
//	 • Splatter Technique deals 4 damage to each creature and
//	   planeswalker."
//
// A plain choose-one. The damage is one instance to every creature and
// planeswalker on the battlefield, the caster's included (Invasion of
// Karsus's sweep), so a creature that is also a planeswalker is dealt 4
// once, not twice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2620dcfe-8738-4f81-b23d-0c1c60cf9047",
		Name:         "Splatter Technique",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Draw four cards.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return DrawCards{N: 4}.Apply(ctx)
				}), game.Purpose{Draws: 4}),
			ModeWithPurpose(ModeDoing("Splatter Technique deals 4 damage to each creature and planeswalker.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return damageEachMatching(ctx, Or(Creature(), Planeswalker()), 4)
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 4}}),
		),
	})
}
