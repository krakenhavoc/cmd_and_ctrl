package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fearless Fledgling — Creature — Griffin {1}{W}, 1/1 (Zendikar Rising,
// #5508):
//
//	"Landfall — Whenever a land you control enters, put a +1/+1 counter
//	 on this creature. It gains flying until end of turn."
//
// Landfall is an ability word (CR 207.2c). Both halves land on the
// Griffin as it resolves, the counter first, and neither happens if it
// left the battlefield in response (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75ae9b87-7137-47a0-b3e1-8915ccd0edf1",
		Name:         "Fearless Fledgling",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Fearless Fledgling — put a +1/+1 counter on this creature; it gains flying until end of turn (landfall)",
				effects.Do(
					effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1},
					effects.Instruction(func(ctx *effects.Context) error {
						return effects.ThisCreatureUntilEOT("Fearless Fledgling — flying", 0, 0, "flying")(ctx.Game, ctx.Item)
					}),
				)),
		},
	})
}
