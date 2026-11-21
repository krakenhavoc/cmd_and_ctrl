package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Maximize Altitude — Sorcery {U} (Guilds of Ravnica, #5525):
//
//	"Target creature gets +1/+1 and gains flying until end of turn.
//	 Jump-start"
//
// The shared socB27TargetGetsAndGains (both halves end at cleanup, CR
// 514.2), with jump-start (cmdr2S39JumpStart, CR 702.133a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "70e8eba9-a106-49c5-932e-b4390edb1c98",
		Name:             "Maximize Altitude",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetCreature("target creature"),
		Purpose:          effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "flying")),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Cmdr2S39JumpStart("{U}")},
		OnResolve:        effects.SocB27TargetGetsAndGains(1, 1, "flying", "Maximize Altitude"),
	})
}
