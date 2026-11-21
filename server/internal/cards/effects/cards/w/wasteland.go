package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wasteland — Land (EDHREC rank 1204):
//
//	"{T}: Add {C}.
//	 {T}, Sacrifice this land: Destroy target nonbasic land."
//
// Strip Mine with the basics off the menu — the whole difference
// between the two cards is the word "nonbasic", and it is the whole
// reason one is banned in Legacy and the other is not. Same
// tap-and-sacrifice cost paid at announce, so the Wasteland is gone
// before the ability resolves. "Nonbasic" reads the Basic supertype
// off the printed type line: a land something else has turned into a
// Swamp is still a nonbasic Swamp, and a Wastes is still basic.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "09a70ae8-3859-4a09-901d-dce063fa3b5f",
		Name:         "Wasteland",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this land: Destroy target nonbasic land.",
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target nonbasic land", effects.Land(), effects.B10NonbasicLand()),
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
