package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glimpse of Tomorrow — Sorcery with no mana cost (Modern Horizons 2):
//
//	"Suspend 3—{R}{R}
//	 Shuffle all permanents you own into your library, then reveal that many
//	 cards from the top of your library. Put all non-Aura permanent cards
//	 revealed this way onto the battlefield, then do the same for Aura cards,
//	 then put the rest on the bottom of your library in a random order."
//
// Suspend is the whole of its playability (no mana cost, CR 118.6). Every
// permanent you own goes to your library at once, whoever controls it, and the
// library is shuffled; the number revealed is the number shuffled in, tokens
// included. The non-Aura permanent cards enter together; then the Aura cards
// enter, each attached to something it can enchant, or staying in the library if
// there is nothing; then everything not put onto the battlefield goes to the
// bottom in a random order. Each step waits for the one before it (the tuck, the
// shuffle and the entry questions are keyed tails).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "12778951-7b11-400b-98ed-06434d7ecc11",
		Name:         "Glimpse of Tomorrow",
		Completeness: effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{
			effects.Suspend(3, "{R}{R}"),
		},
		OnResolve: effects.OldB036Glimpse,
	})
}
