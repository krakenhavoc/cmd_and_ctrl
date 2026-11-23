package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// A Good Day to Die — Instant {B} (Star Trek, #4973 batch 12):
//
//	"Target creature gets +2/+0 until end of turn. When that creature
//	 dies this turn, you draw a card."
//
// The second sentence is a delayed triggered ability the spell creates
// as it resolves (CR 603.7a), bound to that one object: if the creature
// blinks and comes back it is a new object (CR 400.7) and its death no
// longer draws. "You" is the spell's controller, who controls the
// delayed trigger (CR 603.7d), whoever controls the creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4c4559a5-a7a4-4a40-8a8f-e7375c741748",
		Name:         "A Good Day to Die",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || !effects.OnBattlefield(ctx.Game, t.ID) {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Label: "A Good Day to Die — +2/+0"}).Apply(ctx); err != nil {
				return err
			}
			return effects.WhenThatCreatureDiesThisTurn(ctx, t.ID, "A Good Day to Die — you draw a card", effects.DrawOneBody)
		},
	})
}
