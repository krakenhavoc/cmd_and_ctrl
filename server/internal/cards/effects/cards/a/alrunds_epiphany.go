package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alrund's Epiphany — Sorcery {5}{U}{U} (Kaldheim):
//
//	"Create two 1/1 blue Bird creature tokens with flying. Take an extra turn
//	 after this one. Exile Alrund's Epiphany.
//	 Foretell {4}{U}{U} (During your turn, you may pay {2} and exile this card
//	 from your hand face down. Cast it on a later turn for its foretell cost.)"
//
// The Birds first, then the extra turn (CR 500.7), then the spell exiles
// itself instead of going to the graveyard (Temporal Mastery's body).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "bc842838-1993-4480-b8fe-f9d9f2dee7a4",
		Name:           "Alrund's Epiphany",
		Completeness:   effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{effects.Foretell("{4}{U}{U}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := (effects.CreateToken{Template: effects.TokenCard("1/1 blue Bird with flying"), N: 2}).Apply(ctx); err != nil {
				return err
			}
			return effects.TakeAnExtraTurnThenExileThisSpell(item, ctx)
		},
	})
}
