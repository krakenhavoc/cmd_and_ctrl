package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dark_confidant_test.go — the upkeep flip, and the three things
// about it that are easy to get wrong:
//
//   - it REVEALS, so the whole table learns the card;
//   - it does not DRAW, so nothing watching draws fires;
//   - the life is the flipped card's mana value, which is zero for a
//     land and is charged after the card has reached hand.

const DarkConfidantOracle = "2068185c-1b50-47d0-aa3f-bf505d199428"

// seedLibraryTop puts one card with a real mana cost on top of the
// player's library. seedLibrary makes costless Sorceries, which is
// the wrong shape for a card whose whole clause is "equal to its mana
// value".
func SeedLibraryTop(p *game.Player, name, typeLine, manaCost string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		ManaCost:   manaCost,
		Owner:      p.ID,
		Controller: p.ID,
	})
	return id
}
