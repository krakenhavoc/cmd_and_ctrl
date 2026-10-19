package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "1a304703-6b48-4f3f-bd41-7f10ce0a751c",
		Name:         "Wandering Troubadour",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			IfVentureCond(condYouHadALandEnter, AtYourEndStep(ventureLabel("Wandering Troubadour"), Venture)),
		},
	})
}
