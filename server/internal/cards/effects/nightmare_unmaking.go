package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nightmare Unmaking — Sorcery {3}{B}{B}:
//
//	"Choose one —
//	 • Exile each creature with power greater than the number of cards in
//	   your hand.
//	 • Exile each creature with power less than the number of cards in your
//	   hand."
//
// The hand is counted as the spell resolves (it is on the stack, not in
// hand), and each creature's layered power is read then. Power may be
// negative, so "less than 0" is a real set.
//
// No simplification.
func init() {
	sweep := game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepExile, Partial: true}}
	Register(Spec{
		OracleID:     "285fecb7-f872-454d-ac2f-bc5615c70cd8",
		Name:         "Nightmare Unmaking",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(Mode("Exile each creature with power greater than the number of cards in your hand."), sweep),
			ModeWithPurpose(Mode("Exile each creature with power less than the number of cards in your hand."), sweep),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := cardsInHand(ctx.Game, ctx.Controller())
			if ctx.HasMode(0) {
				return ExileAllMatching{Match: And(Creature(), PowerGE(n+1))}.Apply(ctx)
			}
			return ExileAllMatching{Match: And(Creature(), PowerLE(n-1))}.Apply(ctx)
		},
	})
}
