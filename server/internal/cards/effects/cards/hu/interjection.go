package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Interjection — Instant {W} (Secrets of Strixhaven, #4456):
//
//	"Target creature gets +2/+2 and gains first strike until end of turn."
//
// One target, two layered until-end-of-turn changes that end together at
// cleanup. Any creature, yours or not: the card says only "target creature".
// The body is shared with Honorbound Page's prepare spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "47bbadcf-11c6-4a4d-9575-1eb8284a174c",
		Name:         "Interjection",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "first strike")),
		OnResolve:    effects.SocB27TargetGetsAndGains(2, 2, "first strike", "Interjection"),
	})
}
