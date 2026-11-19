package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     oracle,
		Name:         "Depose",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if id := FirstLegalBattlefieldTarget(ctx); id != [16]byte{} {
				if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Deploy",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 2}).Apply(ctx); err != nil {
				return err
			}
			return GainLife{Amount: b14CreaturesControlled(ctx.Game, item.Controller)}.Apply(ctx)
		},
	})
}
