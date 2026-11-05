package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "78c423ad-1f8b-4262-b64f-a98899da9372",
		Name:         "Encounter the Unknown",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Come in Peace — Exile the top two cards of your library. Until the end of your next turn, you may play those cards.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return ExileTopNUntilYourNextTurn(ctx, 2)
				}),
			ModeWithPurpose(ModeDoing("Shoot to Kill — Encounter the Unknown deals 2 damage to each creature and each player.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return b23DamageEachCreatureAndEachPlayer(ctx, 2)
				}), game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2}}),
		),
	})
}
