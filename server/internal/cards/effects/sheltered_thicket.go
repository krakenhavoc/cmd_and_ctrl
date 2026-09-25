package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sheltered Thicket — Land — Mountain Forest (EDHREC rank 4200):
//
//	"({T}: Add {R} or {G}.)
//	 This land enters tapped.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The red-green member of the Amonkhet cycling-land cycle — same
// shape as Polluted Mire and Canyon Slough.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "db8d8643-3d0b-4f20-bf53-f4cd26a0e8df",
		Name:          "Sheltered Thicket",
		Completeness:  CompletenessFull,
		Activated:     []ActivatedAbility{Cycling("{2}")},
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("R", "G")},
	})
}
