package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crimson Operative — Artifact Creature — Human Villain {3}{R}, 3/2
// (Marvel Super Heroes Commander, #4455):
//
//	"Prowess
//	 When this creature enters, exile the top card of your library.
//	 Until the end of your next turn, you may play that card."
//
// Prowess is the printed keyword (Agent of Atlas). The enters trigger is
// the shared impulse exile of one card through the end of the
// controller's next turn (Blazing Crescendo, Burning Curiosity).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d8863742-6c1c-415a-ad27-7def6d119c6a",
		Name:            "Crimson Operative",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Crimson Operative — exile the top card; you may play it until the end of your next turn",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error { return effects.ExileTopNUntilYourNextTurn(ctx, 1) }))),
		},
	})
}
