package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dakmor Salvage — Land (Commander deck reprint):
//
//	"This land enters tapped.
//	 {T}: Add {B}.
//	 Dredge 2 (If you would draw a card, you may mill two cards instead.
//	 If you do, return this card from your graveyard to your hand.)"
//
// Canyon Slough's shape with one colour: enters tapped, one mana
// ability, and the row below.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cdc4048a-73ec-4ec1-a179-2b36c397bf1a",
		Name:         "Dakmor Salvage",
		Completeness: CompletenessFull,
		// Dredge 2
		Replacements: []game.ReplacementEffect{SelfEntersTapped(), Dredge(2)},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
	})
}
