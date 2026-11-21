package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hardbristle Bandit — Creature — Plant Rogue {1}{G}, 1/1 (Outlaws of
// Thunder Junction, #4468):
//
//	"{T}: Add one mana of any color.
//	 Whenever you commit a crime, untap this creature. This ability
//	 triggers only once each turn. (Targeting opponents, anything they
//	 control, and/or cards in their graveyards is a crime.)"
//
// Birds of Paradise's mana ability, and Marauding Sphinx's crime trigger
// (the once-per-announcement crime of CR 700.13, held to once each turn)
// with Blasting Station's untap as its effect: the Bandit untaps only if
// it is still the permanent the trigger came from (CR 400.7). The untap
// is not a mana ability, so it uses the stack and can be responded to.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "116eea43-4235-4608-8c6a-599e860ae2f1",
		Name:         "Hardbristle Bandit",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCommitACrimeOnceEachTurn("Hardbristle Bandit — untap this creature (crime)", effects.B33UntapSelf),
		},
	})
}
