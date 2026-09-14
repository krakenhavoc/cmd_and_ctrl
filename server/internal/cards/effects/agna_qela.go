package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Agna Qel'a — Land (EDHREC rank 2542):
//
//	"This land enters tapped unless you control a basic land.
//	 {T}: Add {U}.
//	 {2}{U}, {T}: Draw a card, then discard a card."
//
// The Avatar: The Last Airbender "abandoned temple" cycle's blue
// member — Abandoned Air Temple's shape with a loot instead of a
// counter placement. The tapped entry is the EntersTappedUnless
// self-replacement on the b11ControlsBasicLand condition; the
// activation is a CR 602 ability with a mana-and-tap cost whose
// effect draws first and then opens the discard prompt (lootOne), so
// the drawn card is a legal discard exactly as in paper.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "22d0a848-2126-48f0-9050-38daaf93b1d0",
		Name:         "Agna Qel'a",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{EntersTappedUnless(b11ControlsBasicLand)},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{2}{U}, {T}: Draw a card, then discard a card.",
			Cost:  Plus(ManaCost("{2}{U}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return lootOne(g, item, 1)
			},
		}},
	})
}
