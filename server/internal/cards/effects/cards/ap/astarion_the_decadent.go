package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Astarion, the Decadent — {4}{W}{B} Legendary Creature — Vampire Elf Rogue
// 4/4
//
//	"Deathtouch, lifelink
//	 At the beginning of your end step, choose one —
//	 • Feed — Target opponent loses life equal to the amount of life they
//	   lost this turn.
//	 • Friends — You gain life equal to the amount of life you gained this
//	   turn."
//
// The amounts are the turn's life tally, read as the mode resolves (CR
// 608.2h): the life lost or gained so far this turn, by any source. Feed's
// target is chosen with its mode as the trigger goes on the stack and
// re-checked as it resolves; the loss is another life loss, so it is not
// counted into the amount already read.
func init() {
	end := effects.AtYourEndStep("Astarion, the Decadent — choose one", func(*game.Game, *game.StackItem) error { return nil })
	end.Modes = effects.ChooseOne(
		effects.ModeDoing("Feed — Target opponent loses life equal to the amount of life they lost this turn.",
			effects.TargetPlayer("target opponent", effects.Opponent()),
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				t, ok := effects.ModeTarget(ctx, 0)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				n := ctx.Game.TurnTallyFor(t.ID).LifeLost
				if n <= 0 {
					return nil
				}
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -n)
			}),
		effects.ModeDoing("Friends — You gain life equal to the amount of life you gained this turn.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				n := ctx.Game.TurnTallyFor(ctx.Controller()).LifeGained
				if n <= 0 {
					return nil
				}
				return effects.GainLife{Player: ctx.Controller(), Amount: n}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:        "03048844-32bb-45a2-a754-73f75eaa4566",
		Name:            "Astarion, the Decadent",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "lifelink"},
		Triggered:       []game.TriggeredAbility{end},
	})
}
