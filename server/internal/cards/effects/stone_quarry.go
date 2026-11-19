package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stone Quarry — Land:
//
//	"This land enters tapped.
//	 {T}: Add {R} or {W}."
//
// A plain tapped Boros dual: the tapped entry is the CR 614 self
// replacement, the mana ability is the two-colour tap.
func init() {
	Register(Spec{
		OracleID:     "90bccf66-58ec-445d-ac96-c6013054a1b4",
		Name:         "Stone Quarry",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R|W}",
			Label:    "Add {R} or {W}",
		}},
	})
}
