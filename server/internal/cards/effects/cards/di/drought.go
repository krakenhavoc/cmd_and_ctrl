package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drought — Enchantment {2}{W}{W} (Ice Age, S82 OLD batch 848, #5636):
//
//	"At the beginning of your upkeep, sacrifice this enchantment unless you pay
//	 {W}{W}.
//	 Spells cost an additional "Sacrifice a Swamp" to cast for each black mana
//	 symbol in their mana costs.
//	 Activated abilities cost an additional "Sacrifice a Swamp" to activate for
//	 each black mana symbol in their activation costs."
//
// Aura Flux's pay-or-sacrifice upkeep tax, and the two #6205 imposed sacrifices:
// each is a Swamp per black mana symbol (hybrid and Phyrexian symbols that can
// be paid with black count) in the spell's mana cost or the ability's mana cost,
// for every player's spells and abilities, added to the cost as it is announced
// (CR 601.2f, 602.2b).
//
// No simplification.
func init() {
	swamp := game.SacrificeKind{Label: "a Swamp", Subtypes: []string{"Swamp"}}
	effects.Register(effects.Spec{
		OracleID:     "ecba74bd-850e-4cbe-9b1c-b0d4d6f27d1c",
		Name:         "Drought",
		Completeness: effects.CompletenessFull,
		ImposedSacrifices: []game.ImposedSacrifice{
			{
				Label:          `Spells cost an additional "Sacrifice a Swamp" to cast for each black mana symbol in their mana costs.`,
				Kind:           swamp,
				Spells:         true,
				PerBlackSymbol: true,
			},
			{
				Label:          `Activated abilities cost an additional "Sacrifice a Swamp" to activate for each black mana symbol in their activation costs.`,
				Kind:           swamp,
				Activations:    true,
				PerBlackSymbol: true,
			},
		},
		Triggered: []game.TriggeredAbility{effects.OldB848UpkeepPayOrSacrifice("Drought", "{W}{W}")},
	})
}
