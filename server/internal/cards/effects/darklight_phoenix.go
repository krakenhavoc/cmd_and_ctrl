package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Darklight Phoenix — Creature — Phoenix {3}{B}, 3/2:
//
//	"Flying, haste
//	 At the beginning of combat on your turn, if two or more creatures
//	 died this turn, return this card from your graveyard to the
//	 battlefield."
//
// A graveyard trigger (CR 113.6k) with an intervening "if" (CR 603.4),
// checked as the step begins. The turn's creature-death tally only
// ever rises, so the condition cannot stop being true while the
// ability waits on the stack; the return itself does nothing if the
// card has left the graveyard in the meantime (CR 400.7). Tokens and
// the controller's own and opponents' creatures all count, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e4b51c2d-7e56-435b-89d6-85e33df65879",
		Name:            "Darklight Phoenix",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			InGraveyard(On(game.EventStepBegan,
				AllOf(StepBegan(game.StepBeginCombat, true),
					func(_ game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
						return b11CreaturesDiedThisTurn(g) >= 2
					}),
				"Darklight Phoenix — return it from your graveyard to the battlefield",
				returnThisCardFromYourGraveyard)),
		},
	})
}
