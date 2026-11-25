package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Howltooth Hollow — Land (Secret Lair Drop, #5462):
//
//	"Hideaway 4. This land enters tapped.
//	 {T}: Add {B}.
//	 {B}, {T}: You may play the exiled card without paying its mana cost if
//	 each player has no cards in hand."
//
// The Lorwyn hideaway shape (hideaway_lands.go) with its own condition,
// asked as the ability resolves: every player still in the game has an empty
// hand.
//
// No simplification.
func init() {
	effects.Register(effects.HideawayLandSpec("463fc699-f4fc-4112-a6b3-6dcb642203e6", "Howltooth Hollow", "B",
		"if each player has no cards in hand", func(g *game.Game, _ uuid.UUID) bool {
			for _, p := range g.Seats {
				if p != nil && !p.Eliminated && p.Hand.Size() > 0 {
					return false
				}
			}
			return true
		}))
}
