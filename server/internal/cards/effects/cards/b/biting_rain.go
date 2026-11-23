package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Biting Rain — Sorcery {2}{B}{B} (Shadows over Innistrad, #5543):
//
//	"All creatures get -2/-2 until end of turn.
//	 Madness {2}{B} (If you discard this card, discard it into exile.
//	 When you do, cast it for its madness cost or put it into your
//	 graveyard.)"
//
// Infest's -2/-2 to every creature until end of turn. Madness is grown by
// the engine from the one string (CR 702.35a, #657).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "485cfaca-9348-418b-a751-2ecb19f3f210",
		Name:         "Biting Rain",
		Completeness: effects.CompletenessFull,
		Madness:      "{2}{B}",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2}},
		OnResolve:    effects.ResolveSteps(effects.BoostUntilEOT{Match: effects.Creature(), Power: -2, Toughness: -2, Label: "Biting Rain — -2/-2 until end of turn"}),
	})
}
