package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scene of the Crime — Artifact Land — Clue (EDHREC rank 2588):
//
//	"This land enters tapped.
//	 {T}: Add {C}.
//	 {T}, Tap an untapped creature you control: Add one mana of any
//	 color.
//	 {2}, Sacrifice this land: Draw a card."
//
// The land that is a Clue. The tapped entry is the real CR 614
// self-replacement; the colourless tap is an ordinary mana ability;
// the crack is the Clue token's own activated ability — "{2},
// Sacrifice this: draw a card", no tap in the cost, so a land that
// entered tapped can still be cracked at once — declared as a CR 602
// ability with a mana-and-sacrifice-self cost.
//
// DECLARED SIMPLIFICATION, weaker than printed: the coloured mana
// ability is not offered. "Tap an untapped creature you control" is
// a cost that taps ANOTHER permanent, and neither ManaAbilityCost nor
// AbilityCost has a tap-another component (the Springleaf Drum seam —
// The Shire's posture); a cost with no shape is left out rather than
// priced at nothing (#259). The land is still recognisably itself
// without it: an artifact land that taps for colourless and cracks
// for a card.
func init() {
	Register(Spec{
		OracleID:     "ba11a517-1dbd-4797-9f5e-46ce0f6c77c0",
		Name:         "Scene of the Crime",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The coloured mana ability isn't available — tapping a creature alongside this land for a mana of any color isn't a cost the engine can pay."},
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{2}, Sacrifice this land: Draw a card.",
			Cost:  Plus(ManaCost("{2}"), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
