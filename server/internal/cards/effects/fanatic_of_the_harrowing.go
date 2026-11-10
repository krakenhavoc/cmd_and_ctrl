package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fanatic of the Harrowing — Creature — Human Cleric {3}{B}, 2/2
// (Duskmourn, #4466):
//
//	"When this creature enters, each player discards a card. If you
//	 discarded a card this way, draw a card."
//
// Kefka's fan-out (EachPlayerDiscardsThenKeyedForEffect): every player is
// asked in turn order and a player with an empty hand is not (CR 701.9b),
// and the draw is the run's continuation, which reads whether the
// CONTROLLER (carried on the run) really discarded.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "878e5e0d-beb9-4f8d-a640-62140c907e52",
		Name:         "Fanatic of the Harrowing",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Fanatic of the Harrowing — each player discards a card; if you discarded a card this way, draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return g.EachPlayerDiscardsThenKeyedForEffect(uuid.Nil,
						game.DiscardPrompt{Source: item.SourceCardID, N: 1, Question: "Fanatic of the Harrowing — discard a card"},
						fanaticOfTheHarrowingDraw, []uuid.UUID{item.Controller})
				}),
		},
	})
}

// fanaticOfTheHarrowingDraw is "If you discarded a card this way, draw a
// card", run once after the last discard has landed. The controller is the
// carry. An on-disk key.
var fanaticOfTheHarrowingDraw = game.RegisterPromptRunThen("prompt-run/fanatic-of-the-harrowing-draw", func(g *game.Game, r game.RunDone) error {
	if len(r.Carry) == 0 || !r.Discards().Discarded(r.Carry[0]) {
		return nil
	}
	return g.DrawNForEffect(r.Carry[0], 1)
})
