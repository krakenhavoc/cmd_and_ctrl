package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Restless Bones — Creature — Skeleton {2}{B}, 1/1 (Guildpact):
//
//	"{3}{B}, {T}: Target creature gains swampwalk until end of turn. (It
//	 can't be blocked as long as defending player controls a Swamp.)
//	 {1}{B}: Regenerate this creature."
//
// Akki Drillmaster's target grant with swampwalk, and the shared "regenerate
// this creature" row (a shield on the Skeleton as the ability resolves).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b6c6e872-1363-480f-9a8d-c393361ccbc1",
		Name:         "Restless Bones",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}{B}, {T}: Target creature gains swampwalk until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "swampwalk")),
				Cost:    effects.Plus(effects.ManaCost("{3}{B}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.NeoB07TargetGetsUntilEOT("Restless Bones — swampwalk until end of turn", 0, 0, "swampwalk"),
			},
			effects.RegenerateThisRow("{1}{B}"),
		},
	})
}
