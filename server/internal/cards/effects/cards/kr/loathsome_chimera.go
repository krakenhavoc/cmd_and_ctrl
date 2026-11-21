package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loathsome Chimera — Creature — Chimera {2}{G}, 4/1:
//
//	"Escape—{4}{G}, Exile three other cards from your graveyard.
//	 (You may cast this card from your graveyard for its escape cost.)
//	 This creature escapes with a +1/+1 counter on it."
//
// The Typhon's cheaper cousin, and the card that shows why escape
// pairs with a fragile body: a 4/1 trades with almost anything, and
// escape is what turns each of those trades into a three-card tax
// rather than a two-for-one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "fe983088-6e0d-4544-90bd-f5ebb95c3418",
		Name:             "Loathsome Chimera",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.EscapeWithCounters("{4}{G}", 3, 1)},
	})
}
