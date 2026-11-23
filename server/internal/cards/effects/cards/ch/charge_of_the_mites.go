package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Charge of the Mites — Instant {2}{W}:
//
//	"Choose one —
//	 • Charge of the Mites deals damage equal to the number of creatures you
//	   control to target creature or planeswalker.
//	 • Create two 1/1 colorless Phyrexian Mite artifact creature tokens with
//	   toxic 1 and 'This token can't block.'"
//
// The damage reads the creature count as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa47c190-d66a-41b0-aa08-21036cda4223",
		Name:         "Charge of the Mites",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 2, CreatureTokens: 2},
		Modes: effects.ChooseOne(
			effects.Mode("Charge of the Mites deals damage equal to the number of creatures you control to target creature or planeswalker.",
				effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker()))),
			effects.Mode("Create two 1/1 colorless Phyrexian Mite artifact creature tokens with toxic 1 and \"This token can't block.\""),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if ctx.HasMode(0) {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				n := 0
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if c.IsCreature() && c.Controller == ctx.Controller() {
						n++
					}
				}
				return effects.DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: n}.Apply(ctx)
			}
			if ctx.HasMode(1) {
				return effects.CreateToken{Template: effects.PhyrexianMiteToken(), N: 2}.Apply(ctx)
			}
			return nil
		},
	})
}
