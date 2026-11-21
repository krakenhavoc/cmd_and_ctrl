package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Depose // Deploy — split card, oracle c547c9ab-a303-48fb-9579-37f9de9a558b
// (Ravnica: Clue Edition):
//
//	Depose {1}{W/U} Instant: "Tap target creature. Draw a card."
//	Deploy {2}{W}{U} Instant: "Create two 1/1 colorless Thopter artifact
//	 creature tokens with flying, then you gain 1 life for each creature you
//	 control."
//
// Two specs, Deploy under "<oracle_id>#1" as Dusk // Dawn is. Deploy counts
// the creatures after the tokens exist, so the Thopters are counted.
//
// No simplification.
func init() {
	const oracle = "c547c9ab-a303-48fb-9579-37f9de9a558b"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Depose",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if id := effects.FirstLegalBattlefieldTarget(ctx); id != [16]byte{} {
				if err := (effects.TapTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Deploy",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := (effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 2}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Amount: effects.B14CreaturesControlled(ctx.Game, item.Controller)}.Apply(ctx)
		},
	})
}
