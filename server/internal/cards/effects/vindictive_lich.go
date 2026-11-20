package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vindictive Lich — Creature — Zombie Wizard {3}{B}, 4/1 (Commander 2017,
// #5533):
//
//	"When this creature dies, choose one or more. Each mode must target a
//	 different player.
//	 • Target opponent sacrifices a creature of their choice.
//	 • Target opponent discards two cards.
//	 • Target opponent loses 5 life."
//
// A modal dies trigger (CR 700.2): one to three bullets chosen as it goes on
// the stack (CR 603.3c), each with its own "target opponent" clause, and
// every clause Distinct (Donnie & April's shape), so no two chosen bullets
// can name the same opponent. The bullets are the ordinary edict (the
// opponent chooses, CR 701.21a), Cerebral Confiscation's discard of the
// opponent's choice, and the life loss; a target gone by resolution
// skips its bullet (CR 608.2b).
//
// No simplification.
func init() {
	dies := WhenThisDies("Vindictive Lich — choose one or more; each mode must target a different player",
		func(*game.Game, *game.StackItem) error { return nil })
	dies.Modes = ChooseN("Choose one or more", 1, 3,
		ModeDoing("Target opponent sacrifices a creature of their choice.",
			Distinct(TargetPlayer("target opponent", Opponent())),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				ctx.Game.PlayerSacrificesForEffect(ctx.Source(), t.ID, sacrificeSpec("a creature", Creature()),
					"Vindictive Lich — sacrifice a creature")
				return nil
			}),
		ModeDoing("Target opponent discards two cards.",
			Distinct(TargetPlayer("target opponent", Opponent())),
			TheModesTargetDiscards(2)),
		ModeDoing("Target opponent loses 5 life.",
			Distinct(TargetPlayer("target opponent", Opponent())),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				return GainLife{Player: t.ID, Amount: -5}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "6dc7f837-158f-4ccf-86a5-a07a20891ec5",
		Name:         "Vindictive Lich",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{dies},
	})
}
