package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Weatherseed Elf — Creature — Elf {G}, 1/1 (Urza's Legacy, S82 OLD b716, #5623):
//
//	"{T}: Target creature gains forestwalk until end of turn. (It can't be blocked as long as defending player controls a Forest.)"
//
// Ivy Dancer's grant: forestwalk through the layered view until end of turn,
// on the creature targeted as the ability resolves. A tap ability, so
// summoning sickness applies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "73d0b29a-5b7a-4835-9bd4-5d00e685b851",
		Name:         "Weatherseed Elf",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature gains forestwalk until end of turn. (It can't be blocked as long as defending player controls a Forest.)",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "forestwalk")),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.NeoB07TargetGetsUntilEOT("Weatherseed Elf — forestwalk until end of turn", 0, 0, "forestwalk"),
		}},
	})
}
