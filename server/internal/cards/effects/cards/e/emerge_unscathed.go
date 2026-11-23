package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emerge Unscathed — Instant {W}:
//
//	"Target creature you control gains protection from the color of
//	 your choice until end of turn.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// Center Soul's text at a lower price. Rebound is the engine's keyword
// (game/rebound.go, #1854).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "894015d5-ee78-45df-8d37-53df963c59a6",
		Name:            "Emerge Unscathed",
		Purpose:         effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "protection")),
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Targets:         effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve:       effects.ProtectionFromAChosenColorForTheTarget("Emerge Unscathed"),
	})
}
