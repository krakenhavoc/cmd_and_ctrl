package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
// fires, so mill payoffs and Bruvac stay out of it. The run is
// bounded by the library's size: an opponent whose whole library
// totals less than 20 simply exiles it all and does not lose for it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e352f5b9-6406-4914-bc79-f24608be6bc9",
		Name:         "Tasha's Hideous Laughter",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, opp := range ctx.Opponents() {
				if err := b27ExileTopUntilTotalManaValue(ctx, opp, 20); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
