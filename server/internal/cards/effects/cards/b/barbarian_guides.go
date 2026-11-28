package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Barbarian Guides — Creature — Human Barbarian {2}{R}, 1/2 (Ice Age, S82 OLD,
// #5636):
//
//	"{2}{R}, {T}: Choose a land type. Target creature you control gains snow
//	 landwalk of the chosen type until end of turn. Return that creature to its
//	 owner's hand at the beginning of the next end step."
//
// The type is chosen as it resolves, from every CR 205.3i land type, and the
// grant is the "snow <type>walk" token (#6143). The delayed return is pinned to
// the creature as the object it is now (CR 603.7c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef8c324a-582b-401e-add1-1985c5596908",
		Name:         "Barbarian Guides",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, {T}: Choose a land type. Target creature you control gains snow landwalk of the chosen type until end of turn. Return that creature to its owner's hand at the beginning of the next end step.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "snow landwalk")),
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect:  effects.OldB833ChooseLandTypeThenSnowwalk("Barbarian Guides"),
		}},
	})
}
