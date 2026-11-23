package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prey's Vengeance — Instant {G}:
//
//	"Target creature gets +2/+2 until end of turn.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// Rebound is the engine's keyword (game/rebound.go, #1854).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "67f6e949-14d9-4253-9d18-4004cac2f900",
		Name:            "Prey's Vengeance",
		Purpose:         effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Targets:         effects.TargetCreature("target creature"),
		OnResolve:       effects.BoostTheTargetUntilEOT(2, 2, "Prey's Vengeance — +2/+2"),
	})
}
