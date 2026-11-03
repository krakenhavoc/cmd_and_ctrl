package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brotherhood's End — Sorcery {1}{R}{R} (Secrets of Strixhaven
// Commander, #4834):
//
//	"Choose one —
//	 • Brotherhood's End deals 3 damage to each creature and each
//	   planeswalker.
//	 • Destroy all artifacts with mana value 3 or less."
//
// A plain choose-one of two untargeted sweeps. The damage is one
// instance to every creature and planeswalker on the battlefield, the
// caster's included (Splatter Technique's sweep), so a creature that
// is also a planeswalker is dealt 3 once. The destroy mode takes every
// artifact permanent whose mana value is 3 or less; a token with no mana
// cost has mana value 0 (CR 202.3a), so a Clue goes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fe7bf8ca-88ae-4c6d-8d71-ef170a87c0d0",
		Name:         "Brotherhood's End",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Brotherhood's End deals 3 damage to each creature and each planeswalker.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return damageEachMatching(ctx, Or(Creature(), Planeswalker()), 3)
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3}}),
			ModeWithPurpose(ModeDoing("Destroy all artifacts with mana value 3 or less.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return DestroyAllMatching{Match: And(Artifact(), ManaValueLE(3))}.Apply(ctx)
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifacts, How: game.SweepDestroy, Partial: true}}),
		),
	})
}
