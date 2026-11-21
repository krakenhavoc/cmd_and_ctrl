package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hideous Visage — Sorcery {2}{B} (Magic 2012, #5573):
//
//	"Creatures you control gain intimidate until end of turn. (Each of
//	 those creatures can't be blocked except by artifact creatures
//	 and/or creatures that share a color with it.)"
//
// Crash Through's keyword grant, intimidate instead of trample, with no
// draw: the creatures you control as it resolves gain it (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd9c5b53-9666-45ee-8f9d-ee1ae74111b6",
		Name:         "Hideous Visage",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"intimidate"},
				Label:    "Hideous Visage — creatures you control gain intimidate",
			},
		),
	})
}
