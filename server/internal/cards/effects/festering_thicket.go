package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Festering Thicket — Land — Swamp Forest (Secrets of Strixhaven
// Commander, #4834):
//
//	"({T}: Add {B} or {G}.)
//	 This land enters tapped.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The black-green member of the Amonkhet cycling-land cycle, the
// Sheltered Thicket and Canyon Slough shape. Its two basic land types
// are on its type line, and it is not basic (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "bdb9b2ce-342c-4935-8943-d0c3971b1e38",
		Name:          "Festering Thicket",
		Completeness:  CompletenessFull,
		Activated:     []ActivatedAbility{Cycling("{2}")},
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("B", "G")},
	})
}
