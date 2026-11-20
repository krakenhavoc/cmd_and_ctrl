package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c29b997a-5be6-4d51-a81b-d00677ef401d",
		Name:         "Essence Feed",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		Purpose: game.Purpose{
			Targets:  game.ForTargets(game.TargetPurpose{Slot: 0, LifeLoss: 3}),
			LifeGain: 3, Tokens: 3, CreatureTokens: 3,
		},
		OnResolve: ResolveSteps(Instruction(func(ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetPlayer {
					return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -3)
				}
			}
			return nil
		}), GainLife{Amount: 3}, roeB1Spawns(3)),
	})
}
