package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mines of Moria — Legendary Land (EDHREC rank 1002):
//
//	"Mines of Moria enters tapped unless you control a legendary
//	 creature.
//	 {T}: Add {R}.
//	 {3}{R}, {T}, Exile three cards from your graveyard: Create two
//	 Treasure tokens."
//
// The legends deck's red land. The conditional entry is the Castle
// shape with a different question — "unless you control a legendary
// creature", read post-layer so an animated legendary artifact
// counts — and the mana is a plain {R}.
//
// DECLARED SIMPLIFICATION: the Treasure ability is not implemented.
// "Exile three cards from your graveyard" is a cost component the
// engine cannot express — AbilityCost carries tap, sacrifice, mana,
// life and loyalty, and nothing that picks cards out of a graveyard
// — and shipping the ability with the cost omitted would make the
// card STRONGER than printed (#259). The land is still recognisably
// itself without it (a conditional-untapped Mountain for a legends
// deck), which is the Ketria Triome / Raffine's Tower posture, and
// the gap runs the weaker way. The ability lands when a
// graveyard-exile cost component does.
func init() {
	Register(Spec{
		OracleID:     "583cdebe-0195-45be-bd2e-5765f07cb902",
		Name:         "Mines of Moria",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The Treasure-making ability isn't implemented — the land can only be played and tapped for {R}."},
		Replacements: []game.ReplacementEffect{SelfEntersTappedUnless(b08ControlsLegendaryCreature)},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
	})
}
