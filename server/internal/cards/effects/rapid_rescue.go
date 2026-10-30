package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rapid Rescue — Instant {G} (Marvel Super Heroes, #4455):
//
//	"Mill two cards. You may put a permanent card from among the milled
//	 cards into your hand. You gain 2 life."
//
// Rickety Gazebo's shape: the mill hands back the cards that actually
// reached the graveyard (a card a replacement sent elsewhere is not among
// them), and the pick is limited to those, not to the whole graveyard. The
// life gain does not depend on the pick, so it is the first line of the
// after-mill continuation.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "de1ad369-5aa3-4c54-8fb1-30b7f1d4d542",
		Name:         "Rapid Rescue",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return MillToZone{N: 2, ThenTail: ctx.ItemTail(mshB7RapidRescueMilled)}.Apply(ctx)
		},
	})
}

// mshB7RapidRescueMilled is everything after the mill: the life gain, which
// does not depend on the pick, and the pick among the milled cards.
var mshB7RapidRescueMilled = CardTail("rapid-rescue/after-mill", func(ctx *Context, o game.TailOutcome) error {
	var milled []uuid.UUID
	milled = append(milled, o.IDs...)
	if err := (GainLife{Amount: 2}).Apply(ctx); err != nil {
		return err
	}
	return mayTakeUpToFromAmongThem(ctx, milled, 1, Permanent(),
		"Rapid Rescue — you may put a permanent card from among the milled cards into your hand")
})
