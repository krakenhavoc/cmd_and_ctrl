package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "406eabe2-df62-49e2-bb39-c0227509d875",
		Name:          "Irrigated Farmland",
		Completeness:  CompletenessFull,
		Activated:     []ActivatedAbility{Cycling("{2}")},
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("W", "U")},
	})
}
