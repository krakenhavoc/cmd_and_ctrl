package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Essence Feed — Sorcery {5}{B} (Rise of the Eldrazi, #5581):
//
//	"Target player loses 3 life. You gain 3 life and create three 0/1
//	 colorless Eldrazi Spawn creature tokens. They have "Sacrifice this
//	 token: Add {C}.""
//
// In printed order: the loss, the gain, the tokens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c29b997a-5be6-4d51-a81b-d00677ef401d",
		Name:         "Essence Feed",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose: game.Purpose{
			Targets:  game.ForTargets(game.TargetPurpose{Slot: 0, LifeLoss: 3}),
			LifeGain: 3, Tokens: 3, CreatureTokens: 3,
		},
		OnResolve: effects.ResolveSteps(effects.Instruction(func(ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetPlayer {
					return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -3)
				}
			}
			return nil
		}), effects.GainLife{Amount: 3}, effects.RoeB1Spawns(3)),
	})
}
