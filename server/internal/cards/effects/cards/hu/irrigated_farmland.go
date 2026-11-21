package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Irrigated Farmland — Land — Plains Island:
//
//	"({T}: Add {W} or {U}.)
//	 This land enters tapped.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The white-blue member of the Amonkhet cycling-land cycle, Canyon
// Slough's shape: a two-colour mana ability, a self enters-tapped
// replacement and the Cycling constructor.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "406eabe2-df62-49e2-bb39-c0227509d875",
		Name:          "Irrigated Farmland",
		Completeness:  effects.CompletenessFull,
		Activated:     []effects.ActivatedAbility{effects.Cycling("{2}")},
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("W", "U")},
	})
}
