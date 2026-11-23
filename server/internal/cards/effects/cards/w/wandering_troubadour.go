package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wandering Troubadour — Creature — Dragon Bard {3}{G}, 4/2:
//
//	"At the beginning of your end step, if you had a land enter the
//	 battlefield under your control this turn, venture into the
//	 dungeon."
//
// An intervening "if" (CR 603.4) on the turn tally's lands entered
// under your control — played, put or fetched alike.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1a304703-6b48-4f3f-bd41-7f10ce0a751c",
		Name:         "Wandering Troubadour",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.IfVentureCond(effects.CondYouHadALandEnter, effects.AtYourEndStep(effects.VentureLabel("Wandering Troubadour"), effects.Venture)),
		},
	})
}
