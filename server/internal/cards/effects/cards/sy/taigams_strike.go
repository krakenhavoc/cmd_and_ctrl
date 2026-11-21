package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Taigam's Strike — Sorcery {3}{U}:
//
//	"Target creature gets +2/+0 until end of turn and can't be blocked
//	 this turn.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// Rebound is the engine's keyword (game/rebound.go, #1854).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f256c828-01d2-4077-823d-8dccd15120bd",
		Name:            "Taigam's Strike",
		Purpose:         effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 2, 0)),
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Targets:         effects.TargetCreature("target creature"),
		OnResolve:       effects.BoostTheTargetAndUnblockableThisTurn(2, "Taigam's Strike"),
	})
}
