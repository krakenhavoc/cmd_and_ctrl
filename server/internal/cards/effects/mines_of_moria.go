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
// #1297: the Treasure ability is live. "Exile three cards from your
// graveyard" is AbilityCost.ExileCards read against the graveyard —
// the activator names the three cards at announce, they leave for
// exile before the ability is on the stack, and a graveyard with fewer
// than three cards offers nothing to pay with. Until the component
// existed the ability was left out rather than shipped with its cost
// omitted (#259), which is what the caveat this replaced said.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "583cdebe-0195-45be-bd2e-5765f07cb902",
		Name:         "Mines of Moria",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTappedUnless(b08ControlsLegendaryCreature)},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{3}{R}, {T}, Exile three cards from your graveyard: Create two Treasure tokens.",
			Cost:  Plus(ManaCost("{3}{R}"), TapCost(), ExileFromGraveyard(3, "three cards", nil)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 2}.Apply(NewContext(g, item))
			},
		}},
	})
}
