package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tilonalli's Knight — Creature — Human Knight {1}{R}, 2/2 (Ixalan):
//
//	"Whenever this creature attacks, if you control a Dinosaur, this creature
//	 gets +1/+1 until end of turn."
//
// An intervening "if" (CR 603.4): checked as the Knight is declared as an
// attacker and again as the trigger resolves. A Dinosaur is read through the
// layers, so a changeling counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae5fe3bd-b6cf-47be-bfb0-05204387bb83",
		Name:         "Tilonalli's Knight",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you control a Dinosaur", func(g *game.Game, you uuid.UUID) bool {
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == you && c.HasSubtypeWord("Dinosaur") {
						return true
					}
				}
				return false
			}, effects.WheneverThisAttacks("Tilonalli's Knight — it gets +1/+1 until end of turn",
				effects.ThisGetsUntilEndOfTurn(1, 1, "Tilonalli's Knight — +1/+1"))),
		},
	})
}
