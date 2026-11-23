package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Earthbrawn — Instant {1}{G} (Morningtide, #5592):
//
//	"Target creature gets +3/+3 until end of turn.
//	 Reinforce 1—{1}{G} ({1}{G}, Discard this card: Put a +1/+1 counter on
//	 target creature.)"
//
// The spell is Giant Growth (boostTheTargetUntilEOT). Reinforce is the
// shared hand-only row (morReinforce): cast it or discard it, never both.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91edf44a-0f46-4936-909a-934e2c4fe176",
		Name:         "Earthbrawn",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.BoostTheTargetUntilEOT(3, 3, "Earthbrawn — +3/+3"),
		Activated: []effects.ActivatedAbility{
			effects.MorReinforce(1, "{1}{G}", "Reinforce 1—{1}{G} ({1}{G}, Discard this card: Put a +1/+1 counter on target creature.)"),
		},
	})
}
