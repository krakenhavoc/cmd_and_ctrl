package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rhonas's Last Stand — Sorcery {G}{G}:
//
//	"Create a 5/4 green Snake creature token. Lands you control don't
//	 untap during your next untap step."
//
// A plain token, then #3995's player-wide skip over the caster's lands.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "01c729d7-9b80-479d-8d30-ae47c76c2c50",
		Name:         "Rhonas's Last Stand",
		Purpose:      game.Purpose{Tokens: 1},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (CreateToken{Template: TokenCard("5/4 green Snake"), N: 1}).Apply(ctx); err != nil {
				return err
			}
			return LandsYouControlDontUntapNextStep(ctx)
		},
	})
}
