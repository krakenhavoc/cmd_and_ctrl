package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fit of Rage — Sorcery {1}{R} (Weatherlight, #5629):
//
//	"Target creature gets +3/+3 and gains first strike until end of turn."
//
// Interjection's pump and grant at +3/+3: a layer 7c change and a layer 6
// grant that both end at cleanup.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "69d08521-74e1-4215-9ed4-3f12137332a4",
		Name:         "Fit of Rage",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3, "first strike")),
		OnResolve:    effects.SocB27TargetGetsAndGains(3, 3, "first strike", "Fit of Rage"),
	})
}
