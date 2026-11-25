package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arcane Adaptation — Enchantment {2}{U} (Ixalan, S82 older sets):
//
//	"As this enchantment enters, choose a creature type.
//	 Creatures you control are the chosen type in addition to their other
//	 types. The same is true for creature spells you control and creature
//	 cards you own that aren't on the battlefield."
//
// Leyline of Transformation's three layer-4 statics (creatures you control,
// creature spells you control, creature cards you own off the battlefield)
// with the type taken from the "as this enters" choice (Card.NamedTribe).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2188c03f-11b6-4651-914a-b4fdd59127e3",
		Name:         "Arcane Adaptation",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseCreatureTypeAsEnters("Arcane Adaptation — choose a creature type"),
		Static:       effects.ChosenTypeInAdditionStatics(),
	})
}
