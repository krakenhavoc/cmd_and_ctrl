package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seeker of Skybreak — Creature — Elf {1}{G}, 2/1 (Mystery Booster 2):
//
//	"{T}: Untap target creature."
//
// An ordinary activated ability (it targets, so it uses the stack). The
// {T} is a creature's own, so summoning sickness applies (CR 302.6), and
// the Seeker may target itself: tapped as a cost, then untapped on
// resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "704723f3-0490-4d96-930a-aacc6d19ba8d",
		Name:         "Seeker of Skybreak",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Untap target creature.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.UntapTheTarget,
		}},
	})
}
