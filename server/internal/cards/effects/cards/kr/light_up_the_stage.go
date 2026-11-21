package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Light Up the Stage — Sorcery {2}{R} (Final Fantasy, #4462):
//
//	"Spectacle {R} (You may cast this spell for its spectacle cost rather
//	 than its mana cost if an opponent lost life this turn.)
//	 Exile the top two cards of your library. Until the end of your next
//	 turn, you may play those cards."
//
// Spectacle is the shared alternative cost (CR 702.137a): offered while an
// opponent has lost life this turn, life lost and not a lower total, and it
// changes neither the timing nor the mana value. The exile is Reckless
// Impulse's body (the S21 play grant, "play" so a land is not stranded, for
// the window of ADR 0063's "until the end of your next turn").
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "8d286376-f386-43ab-a9bd-8478fb9e2497",
		Name:             "Light Up the Stage",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Spectacle("{R}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B19ExileTopTwoUntilEndOfNextTurn(ctx.Game, item)
		},
	})
}
