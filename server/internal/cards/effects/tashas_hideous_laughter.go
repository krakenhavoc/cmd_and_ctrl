package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tasha's Hideous Laughter — Sorcery {1}{U}{U} (EDHREC rank 2921):
//
//	"Each opponent exiles cards from the top of their library until
//	 that player has exiled cards with total mana value 20 or
//	 greater."
//
// The mill spell that punishes low curves. Each opponent in seat
// order, through MillToZone into exile with an Until that keeps a
// running mana-value total and stops on the card that reaches 20 —
// that card is exiled too, as printed. Exile, not mill: no EventMill
// fires, so mill payoffs and Bruvac stay out of it. An opponent whose
// whole library totals less than 20 simply exiles it all and does not
// lose for it (CR 701.17b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e352f5b9-6406-4914-bc79-f24608be6bc9",
		Name:         "Tasha's Hideous Laughter",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			// One opponent at a time in APNAP order (CR 101.4, #4268).
			item := ctx.Item
			return eachPlayerInTurn(ctx.Game, ctx.Opponents(),
				func(g *game.Game, opp uuid.UUID, next func(g *game.Game, cards []uuid.UUID) error) error {
					return b27ExileTopUntilTotalManaValueThen(NewContext(g, item), opp, 20, next)
				}, nil)
		},
	})
}
