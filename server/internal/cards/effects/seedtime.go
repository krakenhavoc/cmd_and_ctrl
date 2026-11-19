package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Seedtime — Instant {1}{G}:
//
//	"Cast this spell only during your turn.
//	 Take an extra turn after this one if an opponent cast a blue spell this
//	 turn."
//
// The timing line is a cast condition (held whichever way it is paid). The
// "if" is read as the spell resolves, off each opponent's cast tally for the
// turn, whose colours are read off each spell as it became cast (CR 601.2i).
// A spell an opponent cast earlier this turn counts even if it has since
// been countered or has resolved.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3ba50802-57d0-450e-9e2f-a46a09d2f6bb",
		Name:         "Seedtime",
		Completeness: CompletenessFull,
		CastCondition: func(g *game.Game, controller uuid.UUID, _ game.Card) bool {
			return g.Turn.ActiveSeat >= 0 && g.Turn.ActiveSeat < len(g.Seats) &&
				g.Seats[g.Turn.ActiveSeat] != nil && g.Seats[g.Turn.ActiveSeat].ID == controller
		},
		CastConditionLabel: "Cast this spell only during your turn.",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, opp := range ctx.Opponents() {
				if ctx.Game.CastTallyFor(opp).Colors != "" && containsColorLetter(ctx.Game.CastTallyFor(opp).Colors, "U") {
					return TakeExtraTurn{}.Apply(ctx)
				}
			}
			return nil
		},
	})
}

func containsColorLetter(colors, letter string) bool {
	for _, r := range colors {
		if string(r) == letter {
			return true
		}
	}
	return false
}
