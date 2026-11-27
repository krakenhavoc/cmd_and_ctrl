package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Riding Red Hare — Sorcery {2}{W}
// (Portal Three Kingdoms, #5622):
//
//	"Target creature gets +3/+3 and gains horsemanship until end of turn. (It can't be blocked except by creatures with horsemanship.)"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ddd45d52-20c3-4731-8860-142af1ccaaa0",
		Name:         "Riding Red Hare",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3, "horsemanship")),
		OnResolve:    effects.SocB27TargetGetsAndGains(3, 3, "horsemanship", "Riding Red Hare"),
	})
}
