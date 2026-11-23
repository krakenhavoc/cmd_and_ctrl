package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ring of the Lucii — Legendary Artifact for {4} (EDHREC rank 4435):
//
//	"{T}: Add {C}{C}.
//	 {2}, {T}, Pay 1 life: Tap target nonland permanent."
//
// A four-mana rock that also locks something down every turn — and
// because the two abilities share the tap, each turn is a choice
// between the mana and the tap-down. Roadmap batch 42 (#449), "no new
// machinery".
//
// The two abilities are different KINDS and the file is the place
// that distinction shows. "Add {C}{C}" is a mana ability: it does not
// use the stack (CR 605.3), resolves synchronously the moment it is
// activated, and cannot be responded to. The tap-down is an ordinary
// CR 602 activated ability that goes on the stack with its target
// chosen at announce, so it can be answered.
//
// The life is a COST, not a rider: it is validated before anything is
// paid, so an activation at 0 life is rejected outright and the Ring
// does not tap. The {2} is mana that has to be floating first, as the
// engine's activated-ability cost path requires.
//
// "Target NONLAND permanent" — the Ring cannot tap down a land, which
// is what keeps it out of the mana-denial slot it would otherwise
// occupy.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d37e595f-afe7-4ad9-87db-b9ceb3f435e8",
		Name:         "Ring of the Lucii",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "Add {C}{C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Pay 1 life: Tap target nonland permanent.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.PayLife(1)),
			Targets: effects.TargetPermanent("target nonland permanent", effects.Nonland()),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
