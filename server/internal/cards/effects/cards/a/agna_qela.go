package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "22d0a848-2126-48f0-9050-38daaf93b1d0",
		Name:         "Agna Qel'a",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.EntersTappedUnless(effects.B11ControlsBasicLand)},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}, {T}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{U}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.LootOne(g, item, 1)
			},
		}},
	})
}
