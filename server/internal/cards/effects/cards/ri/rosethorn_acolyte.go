package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rosethorn Acolyte // Seasonal Ritual — an adventure card (Throne of Eldraine):
//
//	Rosethorn Acolyte {2}{G}, Creature — Elf Druid 2/3:
//	"{T}: Add one mana of any color."
//
//	Seasonal Ritual {G}, Sorcery — Adventure:
//	"Add one mana of any color. (Then exile this card. You may cast the creature
//	 later from exile.)"
//
// The adventure layout is Brazen Borrower's (CR 715.3d): the creature under the
// bare oracle ID and the Adventure half under "<oracle_id>#1". The creature's
// mana ability is Birds of Paradise's; the Adventure is a sorcery, so it uses
// the stack (CR 605.3a) and adds its mana as it resolves.
//
// No simplification.
func init() {
	const oracleID = "639a8804-8513-4ba4-8683-ea1e2e5a9530"
	effects.Register(effects.Spec{
		OracleID:     oracleID,
		Name:         "Rosethorn Acolyte",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracleID + "#1",
		Name:         "Seasonal Ritual",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve:    effects.ResolveSteps(effects.AddMana{Produced: "{W|U|B|R|G}"}),
	})
}
