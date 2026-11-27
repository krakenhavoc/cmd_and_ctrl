package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Morality Shift — Sorcery {5}{B}{B} (Judgment):
//
//	"Exchange your graveyard and library. Then shuffle your library."
//
// Two simultaneous moves in turn (S82OldB627ExchangeGraveyardAndLibrary): the
// library goes to the graveyard, then the cards that were in the graveyard
// before go to the library, and the library is shuffled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "82e36ba9-9652-4594-8660-86c766b47709",
		Name:         "Morality Shift",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB627ExchangeGraveyardAndLibrary(ctx)
		},
	})
}
