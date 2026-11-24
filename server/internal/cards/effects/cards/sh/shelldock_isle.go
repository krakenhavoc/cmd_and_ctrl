package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shelldock Isle — Land (Secret Lair Drop, #5462):
//
//	"Hideaway 4. This land enters tapped.
//	 {T}: Add {U}.
//	 {U}, {T}: You may play the exiled card without paying its mana cost if a
//	 library has twenty or fewer cards in it."
//
// The Lorwyn hideaway shape (hideaway_lands.go) with its own condition,
// asked as the ability resolves: the library of any player still in the game
// holds twenty or fewer cards.
//
// No simplification.
func init() {
	effects.Register(effects.HideawayLandSpec("f748b2fb-6c2a-400a-8e96-fa4e4a1dfe80", "Shelldock Isle", "U",
		"if a library has twenty or fewer cards in it", func(g *game.Game, _ uuid.UUID) bool {
			for _, p := range g.Seats {
				if p != nil && !p.Eliminated && p.Library.Size() <= 20 {
					return true
				}
			}
			return false
		}))
}
