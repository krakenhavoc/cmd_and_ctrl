package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Karakas — Legendary Land (Legends, #5639):
//
//	"{T}: Add {W}.
//	 {T}: Return target legendary creature to its owner's hand."
//
// A fixed {W} mana ability, and a tap ability whose target clause is a
// legendary creature on the battlefield, anyone's, a commander
// included: it goes to its owner's hand (and its owner may then move a
// commander to the command zone, CR 903.9b). The body is the shared
// bounceEachLegalTarget, so a creature that stopped being a legal target
// in response is left alone (CR 608.2b). A land is not summoning sick
// (CR 302.6 is for creatures), so both abilities work the turn it is
// played.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "59119143-c0fa-49dd-adf0-e2fd3029c48b",
		Name:         "Karakas",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Return target legendary creature to its owner's hand.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target legendary creature", effects.Legendary()),
			Effect:  effects.BounceEachLegalTarget,
		}},
	})
}
