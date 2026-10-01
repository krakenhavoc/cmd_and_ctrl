package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ephemerate — Instant {W}:
//
//	"Exile target creature you control, then return it to the
//	 battlefield under its owner's control.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// The flicker is Momentary Blink's: the creature comes back under its
// OWNER's control, as a new object (CR 400.7). Rebound is the engine's
// keyword (game/rebound.go, #1854).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0fd57894-b917-41c8-a394-360d1d31b236",
		Name:            "Ephemerate",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Targets:         TargetCreature("target creature you control", YouControl()),
		OnResolve:       flickerTheTargetToItsOwner,
	})
}
