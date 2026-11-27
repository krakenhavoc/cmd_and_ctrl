package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Disciple of Tevesh Szat — Creature — Human Cleric {2}{B}{B}, 3/1 (Coldsnap, S82 OLD b560):
//
//	"{T}: Target creature gets -1/-1 until end of turn.
//	 {4}{B}{B}, {T}, Sacrifice this creature: Target creature gets -6/-6 until
//	 end of turn."
//
// The sacrifice is a cost, so the second ability resolves even though the
// Disciple is gone (CR 112.7a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a09c951-9861-4e95-acf1-5ce1a3778493",
		Name:         "Disciple of Tevesh Szat",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Target creature gets -1/-1 until end of turn.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target creature"),
				Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -1, -1)),
				Effect:  effects.PumpAndGrantFirstTarget(-1, -1, "Disciple of Tevesh Szat — -1/-1 until end of turn"),
			},
			{
				Label:   "{4}{B}{B}, {T}, Sacrifice this creature: Target creature gets -6/-6 until end of turn.",
				Cost:    effects.Plus(effects.ManaCost("{4}{B}{B}"), effects.TapCost(), effects.SacrificeThis()),
				Targets: effects.TargetCreature("target creature"),
				Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -6, -6)),
				Effect:  effects.PumpAndGrantFirstTarget(-6, -6, "Disciple of Tevesh Szat — -6/-6 until end of turn"),
			},
		},
	})
}
