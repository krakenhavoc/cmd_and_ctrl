package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	dies := effects.WhenThisDies("Vindictive Lich — choose one or more; each mode must target a different player",
		func(*game.Game, *game.StackItem) error { return nil })
	dies.Modes = effects.ChooseN("Choose one or more", 1, 3,
		effects.ModeDoing("Target opponent sacrifices a creature of their choice.",
			effects.Distinct(effects.TargetPlayer("target opponent", effects.Opponent())),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				ctx.Game.PlayerSacrificesForEffect(ctx.Source(), t.ID, effects.SacrificeSpec("a creature", effects.Creature()),
					"Vindictive Lich — sacrifice a creature")
				return nil
			}),
		effects.ModeDoing("Target opponent discards two cards.",
			effects.Distinct(effects.TargetPlayer("target opponent", effects.Opponent())),
			effects.TheModesTargetDiscards(2)),
		effects.ModeDoing("Target opponent loses 5 life.",
			effects.Distinct(effects.TargetPlayer("target opponent", effects.Opponent())),
			vindictiveLichTargetLosesFive),
	)
	effects.Register(effects.Spec{
		OracleID:     "6dc7f837-158f-4ccf-86a5-a07a20891ec5",
		Name:         "Vindictive Lich",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{dies},
	})
}

// vindictiveLichTargetLosesFive is the third bullet, "Target opponent loses
// 5 life": its own declaration, so its target read is never mistaken for a
// gate on the first bullet's sacrifice.
func vindictiveLichTargetLosesFive(_ *game.StackItem, ctx *effects.Context, occ int) error {
	t, ok := effects.ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	return effects.GainLife{Player: t.ID, Amount: -5}.Apply(ctx)
}
