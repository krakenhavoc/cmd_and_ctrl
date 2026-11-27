package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mercadian Atlas — Artifact — Book {5} (Mercadian Masques, S82 OLD, #5619):
//
//	"At the beginning of your end step, if you didn't play a land this turn,
//	 you may draw a card."
//
// An intervening "if" (CR 603.4), checked as the end step begins and again as
// the ability resolves, read off the turn's land-play tally (a land played from
// the hand or by an effect counts, as for River of Tears). The "may" is asked
// as it resolves, and a draw limit that stops the draw stops the question
// (#4016).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5bae9ffb-99b3-4feb-8aed-265d874fd68b",
		Name:         "Mercadian Atlas",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(effects.InterveningIfYou("you didn't play a land this turn",
				func(g *game.Game, you uuid.UUID) bool { return g.LandsPlayedThisTurnFor(you) == 0 },
				effects.AtYourEndStep("Mercadian Atlas — you may draw a card", effects.Do(effects.DrawCards{N: 1}))),
				"Mercadian Atlas — draw a card?", 1),
		},
	})
}
