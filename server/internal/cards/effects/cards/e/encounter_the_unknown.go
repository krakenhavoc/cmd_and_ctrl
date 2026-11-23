package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Encounter the Unknown — {1}{R}{R} Instant (Star Trek, #4973):
//
//	"Choose one —
//	 • Come in Peace — Exile the top two cards of your library. Until the
//	   end of your next turn, you may play those cards.
//	 • Shoot to Kill — Encounter the Unknown deals 2 damage to each
//	   creature and each player."
//
// Atsushi, the Blazing Sky's first bullet (ExileTopNUntilYourNextTurn,
// ADR 0063's "your next turn" window) and Jaya Ballard's sweep at 2: one
// damage instance to every creature and every player still in the game.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "78c423ad-1f8b-4262-b64f-a98899da9372",
		Name:         "Encounter the Unknown",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Come in Peace — Exile the top two cards of your library. Until the end of your next turn, you may play those cards.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.ExileTopNUntilYourNextTurn(ctx, 2)
				}),
			effects.ModeWithPurpose(effects.ModeDoing("Shoot to Kill — Encounter the Unknown deals 2 damage to each creature and each player.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.B23DamageEachCreatureAndEachPlayer(ctx, 2)
				}), game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2}}),
		),
	})
}
