package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Terrarion — Artifact {1} (Jumpstart, #5510):
//
//	"This artifact enters tapped.
//	 {2}, {T}, Sacrifice this artifact: Add two mana in any combination of
//	 colors.
//	 When this artifact is put into a graveyard from the battlefield, draw a
//	 card."
//
// The Odyssey Eggs' enters-tapped, sacrifice-for-mana shape with
// Flamebraider's two mana of any combination. The mana is a mana ability
// (CR 605.1a), so it never uses the stack, and the sacrifice is a cost: the
// Terrarion reaches the graveyard and the draw trigger follows.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e7ade56-ab68-42ab-8a6f-bc38957f1563",
		Name:         "Terrarion",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Sacrifice: true, Mana: "{2}"},
			Produced: effects.AnyCombinationOfColors(2),
			Label:    "{2}, {T}, Sacrifice this artifact: Add two mana in any combination of colors.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Terrarion — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
