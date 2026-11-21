package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wirewood Lodge — Land (EDHREC rank 2392):
//
//	"{T}: Add {C}.
//	 {G}, {T}: Untap target Elf."
//
// The Elf deck's second activation. The mana half is an ordinary tap
// ability; the untap is a CR 602 activated ability sharing the tap,
// targeting any permanent with the Elf subtype — anyone's, as
// printed, and a changeling counts (post-layer subtype).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1275653f-de4e-4fe9-aad8-88555fa11680",
		Name:         "Wirewood Lodge",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, {T}: Untap target Elf.",
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target Elf", effects.Subtype("Elf")),
			Effect:  effects.UntapTheTarget,
		}},
	})
}
